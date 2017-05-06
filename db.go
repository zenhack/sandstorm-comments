package main

// Code dealing with database access.

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
)

var (
	dbPath = strDefault(os.Getenv("DB_PATH"), "./db.sqlite3")
	db     *sql.DB

	kvDefaults = map[string]string{
		"require-moderation": "true",
		"require-sign-in":    "true",
	}
)

// Get a key from the key value store. If it is not set, sets it to the value
// defined in kvDefaults.
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

// Sets a key in the key value store.
func setKey(key string, value string) error {
	_, err := db.Exec(
		"INSERT OR REPLACE INTO key_val (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

// Initialize the database. This connects and loads the schema, The schema is
// written such that this is idempotent.
func initDB() {
	schema, err := ioutil.ReadFile(schemaFile)
	chkfatal(err)
	db, err = sql.Open("sqlite3", dbPath)
	chkfatal(err)
	_, err = db.Exec(string(schema))
	chkfatal(err)
}

// Insert the comment into the database.
func insertComment(comment Comment) error {
	_, err := db.Exec(
		"INSERT INTO comments (article, author, body, needsMod) VALUES (?, ?, ?, ?)",
		comment.ArticleId,
		comment.Author,
		comment.Body,
		comment.NeedsModeration,
	)
	return err
}

// Get all the comments for articleId which are published (needsMod is false).
func getPublishedComments(articleId string) ([]Comment, error) {
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
