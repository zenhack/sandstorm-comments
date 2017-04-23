package main

import (
	"crypto/rand"
	"database/sql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var (
	dbPath      = strDefault(os.Getenv("DB_PATH"), "./db.sqlite3")
	staticDir   = strDefault(os.Getenv("STATIC_ASSETS"), "static")
	templateDir = strDefault(os.Getenv("TEMPLATE_DIR"), "templates")
	schemaFile  = strDefault(os.Getenv("SCHEMA_FILE"), "./schema.sql")
	csrfKeyfile = strDefault(os.Getenv("CSRF_KEYFILE"), "./csrfKey")

	tpls = template.Must(template.ParseGlob(templateDir + "/*.html"))

	db *sql.DB
)

type Comment struct {
	Author string
	Body   string
}

type SafeComment struct {
	Author string
	Body   template.HTML
}

type CommentPageArgs struct {
	ArticleId string
	Comments  []SafeComment
	CSRFField template.HTML
}

func (c Comment) Sanitize() SafeComment {
	unsafeHtml := blackfriday.MarkdownCommon([]byte(c.Body))
	safeHtml := bluemonday.UGCPolicy().SanitizeBytes(unsafeHtml)
	return SafeComment{
		Author: c.Author,
		Body:   template.HTML(safeHtml),
	}
}

func strDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func chkfatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func sanitizeComments(comments []Comment) []SafeComment {
	ret := make([]SafeComment, len(comments))
	for i := range comments {
		ret[i] = comments[i].Sanitize()
	}
	return ret
}

func getComments(articleId string) ([]Comment, error) {
	rows, err := db.Query(
		"SELECT author, body FROM comments WHERE article = ? and needsMod = 0",
		articleId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := []Comment{}
	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(&comment.Author, &comment.Body)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func initCSRF() func(http.Handler) http.Handler {
	key, err := ioutil.ReadFile(csrfKeyfile)
	if err != nil || len(key) != 32 {
		log.Print("Generating new CSRF Key")
		key = make([]byte, 32)
		rand.Read(key)
		chkfatal(ioutil.WriteFile(csrfKeyfile, key, 0600))
	}
	return csrf.Protect(key, csrf.Secure(os.Getenv("DEV_MODE") != "1"))
}

func initDB() {
	schema, err := ioutil.ReadFile(schemaFile)
	chkfatal(err)
	db, err = sql.Open("sqlite3", dbPath)
	chkfatal(err)
	_, err = db.Exec(string(schema))
	chkfatal(err)
}

func addPost(needsModeration bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Invalid form body"))
		}

		articleId := req.PostForm.Get("article_id")
		comment := Comment{
			Author: req.PostForm.Get("author"),
			Body:   req.PostForm.Get("body"),
		}
		_, err := db.Exec(
			"INSERT INTO comments (article, author, body, needsMod) VALUES (?, ?, ?, ?)",
			articleId,
			comment.Author,
			comment.Body,
			needsModeration,
		)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error"))
			return
		}
	})
}

func main() {
	// CSRF := initCSRF()
	CSRF := func(h http.Handler) http.Handler { return h }
	initDB()
	r := mux.NewRouter()

	r.Methods("GET").PathPrefix("/static/").
		Handler(http.FileServer(http.Dir(staticDir + "/..")))
	r.Methods("GET").Path("/"). // MatcherFunc(havePermission("admin")).
					HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			tpls.ExecuteTemplate(w, "index.html", nil)
		})
	r.Methods("POST").Path("/settings").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, "/", http.StatusSeeOther)
		})
	r.Methods("POST").Path("/new-comment").
		MatcherFunc(havePermission("post-unmoderated")).
		Handler(CSRF(addPost(false)))
	r.Methods("POST").Path("/new-comment").
		MatcherFunc(havePermission("post-moderated")).
		Handler(CSRF(addPost(true)))
	r.Methods("GET").Path("/comments").HandlerFunc(showComments)
	http.Handle("/", r)
	http.ListenAndServe(":8000", nil)
}

func showComments(w http.ResponseWriter, req *http.Request) {
	articleId := req.URL.Query().Get("articleId")
	if articleId == "" {
		w.WriteHeader(400)
		w.Write([]byte("Error: articleId not set."))
		return
	}
	comments, err := getComments(articleId)
	if err != nil {
		log.Printf("Error getting comments for article %q: %v", articleId, err)
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
		return
	}
	safeComments := sanitizeComments(comments)
	log.Printf("CSRF Token: %q", csrf.Token(req))
	err = tpls.ExecuteTemplate(w, "comments.html", CommentPageArgs{
		ArticleId: articleId,
		Comments:  safeComments,
		CSRFField: csrf.TemplateField(req),
	})
	if err != nil {
		log.Print("Rendering template:", err)
	}
}

func havePermission(name string) mux.MatcherFunc {
	return mux.MatcherFunc(func(req *http.Request, match *mux.RouteMatch) bool {
		if os.Getenv("SANDSTORM") != "1" {
			return true
		}
		perms := strings.Split(req.Header.Get("X-Sandstorm-Permissions"), ",")
		for _, p := range perms {
			if p == name {
				return true
			}
		}
		return false
	})
}
