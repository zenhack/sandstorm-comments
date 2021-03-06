package main

// Misc. helpers.

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Return s, or def if s == "".
func strDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// If err is not nil, log it and exit the program.
func chkfatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Wrapper around getKey that calls serverErr on failure.
func mustGetKey(w http.ResponseWriter, key string) (string, error) {
	val, err := getKey(key)
	if err != nil {
		serverErr(w, "Getting/setting key "+key, err)
	}
	return val, err
}

// Wrapper around setKey that calls serverErr on failure.
func mustSetKey(w http.ResponseWriter, key string, value string) error {
	err := setKey(key, value)
	if err != nil {
		serverErr(w, "Setting key "+key, err)
	}
	return err
}

// Log a server error and report a generic 500 to the client.
func serverErr(w http.ResponseWriter, ctx string, err error) {
	log.Printf("Error %s: %v", ctx, err)
	w.WriteHeader(500)
	w.Write([]byte("Internal Server Error"))
}

// Return whether the client has the named permission.
func havePermission(name string, req *http.Request) bool {
	perms := strings.Split(req.Header.Get("X-Sandstorm-Permissions"), ",")
	for _, p := range perms {
		if p == name {
			return true
		}
	}
	return false
}

// Matcher func which checks that the client has the named permission, according to
// `havePermission`, above.
func matchPermission(name string) mux.MatcherFunc {
	return mux.MatcherFunc(func(req *http.Request, match *mux.RouteMatch) bool {
		return havePermission(name, req)
	})
}
