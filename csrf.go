package main

import (
	"bytes"
	"crypto/rand"
	"golang.org/x/net/xsrftoken"
	"html/template"
	"net/http"
)

func loadCsrfKey() (string, error) {
	// TODO: naming is a bit confusing here; getKey refers to a key in
	// a dictionary, while loadCsrfKey refers to a cryptographic key.
	// would be nice to rework these to be a bit less confusing.
	val, err := getKey("csrf-key")
	if err != nil || val == "" {
		var buf [32]byte
		_, err = rand.Read(buf[:])
		if err != nil {
			return "", err
		}
		val = string(buf[:])
		err = setKey("csrf-key", val)
	}
	return val, err
}

type csrfGuard struct {
	key     string
	okPath  string // path from whence it is ok to post
	handler http.Handler
}

func (g csrfGuard) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !xsrftoken.Valid(req.FormValue("csrf"), g.key, "", g.okPath) {
		w.WriteHeader(400)
		w.Write([]byte("Invalid CSRF Token"))
		return
	}
	g.handler.ServeHTTP(w, req)
}

func csrfGen(key, path string) string {
	return xsrftoken.Generate(key, "", path)
}

func csrfField(key string, req *http.Request) template.HTML {
	token := csrfGen(key, req.URL.Path)
	buf := &bytes.Buffer{}
	template.HTMLEscape(buf, []byte(token))
	return template.HTML(`<input type="hidden" name="csrf" value="` + buf.String() + `" />`)
}
