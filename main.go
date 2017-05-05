package main

import (
	"database/sql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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
	csrfKey     = ""

	tpls = template.Must(template.ParseGlob(templateDir + "/*.html"))

	db *sql.DB

	kvDefaults = map[string]string{
		"require-moderation": "true",
		"require-sign-in":    "true",
	}
)

type Comment struct {
	Author          string
	Body            string
	SafeBody        template.HTML
	Id              int
	NeedsModeration bool
	ArticleId       string
}

type CommentPageArgs struct {
	ArticleId  string
	Comments   []Comment
	CSRFField  template.HTML
	NeedsLogin bool
	Moderated  bool
}

type Settings struct {
	RequireModeration bool
	RequireSignIn     bool
}

type AdminPageArgs struct {
	Settings  Settings
	Comments  []Comment
	CSRFField template.HTML
}

func getKey(key string) (string, error) {
	row := db.QueryRow("SELECT value FROM key_val WHERE key = ?", key)
	ret := ""
	err := row.Scan(&ret)
	if err != nil {
		ret = kvDefaults[key]
		err = setKey(key, ret)
	}
	return ret, err
}

func mustGetKey(w http.ResponseWriter, key string) (string, error) {
	val, err := getKey(key)
	if err != nil {
		serverErr(w, "Getting/setting key "+key, err)
	}
	return val, err
}

func setKey(key string, value string) error {
	_, err := db.Exec(
		"INSERT OR REPLACE INTO key_val (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

func mustSetKey(w http.ResponseWriter, key string, value string) error {
	err := setKey(key, value)
	if err != nil {
		serverErr(w, "Setting key "+key, err)
	}
	return err
}

// Populate the SafeBody field of c, by sanitizing the Body field.
func (c *Comment) Sanitize() {
	unsafeHtml := blackfriday.MarkdownCommon([]byte(c.Body))
	safeHtml := bluemonday.UGCPolicy().SanitizeBytes(unsafeHtml)
	c.SafeBody = template.HTML(safeHtml)
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

func initDB() {
	schema, err := ioutil.ReadFile(schemaFile)
	chkfatal(err)
	db, err = sql.Open("sqlite3", dbPath)
	chkfatal(err)
	_, err = db.Exec(string(schema))
	chkfatal(err)
}

func addComment(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Invalid form body"))
	}

	val, err := mustGetKey(w, "require-moderation")
	if err != nil {
		return
	}

	requireModeration := val != "false"

	articleId := req.PostForm.Get("article_id")
	comment := Comment{
		Author: req.PostForm.Get("author"),
		Body:   req.PostForm.Get("body"),
	}
	_, err = db.Exec(
		"INSERT INTO comments (article, author, body, needsMod) VALUES (?, ?, ?, ?)",
		articleId,
		comment.Author,
		comment.Body,
		requireModeration,
	)
	if err != nil {
		serverErr(w, "Saving comment to the database", err)
		return
	}
	http.Redirect(w, req, req.PostForm.Get("redirect"), http.StatusSeeOther)
}

func adminPage(w http.ResponseWriter, req *http.Request) {
	requireModeration, err := mustGetKey(w, "require-moderation")
	if err != nil {
		return
	}
	requireSignIn, err := mustGetKey(w, "require-sign-in")
	if err != nil {
		return
	}

	rows, err := db.Query(
		`SELECT id, author, body, needsMod, article
				FROM comments`)
	if err != nil {
		serverErr(w, "Loading comments from admin page", err)
		return
	}
	defer rows.Close()
	comments := []Comment{}
	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(
			&comment.Id,
			&comment.Author,
			&comment.Body,
			&comment.NeedsModeration,
			&comment.ArticleId)
		if err != nil {
			serverErr(w, "Loading comments from admin page", err)
			return
		}
		(&comment).Sanitize()
		comments = append(comments, comment)
	}

	err = tpls.ExecuteTemplate(w, "index.html", AdminPageArgs{
		Settings: Settings{
			RequireModeration: requireModeration != "false",
			RequireSignIn:     requireSignIn != "false",
		},
		Comments:  comments,
		CSRFField: csrfField(csrfKey, req),
	})
	if err != nil {
		log.Print("Error rendering template:", err)
	}
}

func commentSignIn(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Sign in page not yet implemeneted"))
}

func postSettings(w http.ResponseWriter, req *http.Request) {
	readCheckBox := func(val *string) {
		if *val == "on" {
			*val = "true"
		} else {
			*val = "false"
		}
	}
	requireModeration := req.PostForm.Get("require-moderation")
	requireSignIn := req.PostForm.Get("require-sign-in")
	readCheckBox(&requireModeration)
	readCheckBox(&requireSignIn)
	if mustSetKey(w, "require-moderation", requireModeration) != nil {
		return
	}
	if mustSetKey(w, "require-sign-in", requireSignIn) != nil {
		return
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func deleteComment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	_, err := db.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		serverErr(w, "deleting comment", err)
		return
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
func approveComment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	_, err := db.Exec("UPDATE comments SET needsMod = 0 WHERE id = ?", id)
	if err != nil {
		serverErr(w, "approving comment", err)
		return
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func main() {
	var err error
	initDB()
	csrfKey, err = loadCsrfKey()
	chkfatal(err)

	r := mux.NewRouter()

	r.Methods("GET").PathPrefix("/static/").
		Handler(http.FileServer(http.Dir(staticDir + "/..")))
	r.Methods("GET").Path("/").
		MatcherFunc(havePermission("admin")).
		HandlerFunc(adminPage)
	r.Methods("POST").Path("/settings").
		MatcherFunc(havePermission("admin")).
		Handler(csrfGuard{csrfKey, "/", http.HandlerFunc(postSettings)})
	r.Methods("POST").Path("/delete/{id:[0-9]+}").
		MatcherFunc(havePermission("admin")).
		Handler(csrfGuard{csrfKey, "/", http.HandlerFunc(deleteComment)})
	r.Methods("POST").Path("/approve/{id:[0-9]+}").
		MatcherFunc(havePermission("admin")).
		Handler(csrfGuard{csrfKey, "/", http.HandlerFunc(approveComment)})
	r.Methods("POST").Path("/new-comment").
		MatcherFunc(havePermission("post")).
		Handler(csrfGuard{csrfKey, "/comments", http.HandlerFunc(addComment)})
	r.Methods("GET").Path("/comment-sign-in").
		MatcherFunc(havePermission("post")).
		Handler(http.HandlerFunc(commentSignIn))
	r.Methods("GET").Path("/comments").HandlerFunc(showComments)
	http.Handle("/", r)
	http.ListenAndServe(":8000", nil)
}

func serverErr(w http.ResponseWriter, ctx string, err error) {
	log.Printf("Error %s: %v", ctx, err)
	w.WriteHeader(500)
	w.Write([]byte("Internal Server Error"))
}

func showComments(w http.ResponseWriter, req *http.Request) {
	requireModeration, err := mustGetKey(w, "require-moderation")
	if err != nil {
		return
	}
	requireSignIn, err := mustGetKey(w, "require-sign-in")
	if err != nil {
		return
	}

	articleId := req.URL.Query().Get("articleId")
	if articleId == "" {
		w.WriteHeader(400)
		w.Write([]byte("Error: articleId not set."))
		return
	}
	comments, err := getComments(articleId)
	if err != nil {
		serverErr(w, "getting comments for article: "+articleId, err)
		return
	}
	for i := range comments {
		(&comments[i]).Sanitize()
	}
	err = tpls.ExecuteTemplate(w, "comments.html", CommentPageArgs{
		ArticleId:  articleId,
		Comments:   comments,
		CSRFField:  csrfField(csrfKey, req),
		Moderated:  requireModeration != "false",
		NeedsLogin: requireSignIn != "false",
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
