package controllers

import (
	"fmt"
	"net/http"
)

func ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Fprintf(w, "email: %s \n pass: %s", email, password)
}

func ProcessSignUp(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Fprintf(w, "email: %s \n pass: %s", email, password)
}
