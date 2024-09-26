package controllers

import "net/http"

func StaticHandler(tpl template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, nil)
	}
}

func FaqHandler(tpl template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := "the world "
		tpl.Execute(w, r, data)
	}
}
