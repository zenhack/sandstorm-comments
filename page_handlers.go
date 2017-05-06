package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Settings struct {
	RequireModeration bool
	RequireSignIn     bool
}

type CommentPageArgs struct {
	ArticleId  string
	Comments   []Comment
	CSRFField  template.HTML
	NeedsLogin bool
	Moderated  bool
}

type AdminPageArgs struct {
	Settings  Settings
	Comments  []Comment
	CSRFField template.HTML
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

	err = insertComment(Comment{
		Author:          req.PostForm.Get("author"),
		Body:            req.PostForm.Get("body"),
		NeedsModeration: requireModeration,
		ArticleId:       req.PostForm.Get("article_id"),
	})
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

func commentSignIn(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Sign in page not yet implemeneted"))
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
	comments, err := getPublishedComments(articleId)
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
