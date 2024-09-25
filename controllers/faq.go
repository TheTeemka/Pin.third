package controllers

import "net/http"

func FaqHandler(tpl template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := "the world "
		tpl.Execute(w, data)
	}
}
