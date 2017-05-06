package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var (
	staticDir   = strDefault(os.Getenv("STATIC_ASSETS"), "static")
	templateDir = strDefault(os.Getenv("TEMPLATE_DIR"), "templates")
	schemaFile  = strDefault(os.Getenv("SCHEMA_FILE"), "./schema.sql")
	csrfKey     = ""

	tpls = template.Must(template.ParseGlob(templateDir + "/*.html"))
)

type Comment struct {
	Author          string
	Body            string
	SafeBody        template.HTML
	Id              int
	NeedsModeration bool
	ArticleId       string
}

// Populate the SafeBody field of c, by sanitizing the Body field.
func (c *Comment) Sanitize() {
	unsafeHtml := blackfriday.MarkdownCommon([]byte(c.Body))
	safeHtml := bluemonday.UGCPolicy().SanitizeBytes(unsafeHtml)
	c.SafeBody = template.HTML(safeHtml)
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
