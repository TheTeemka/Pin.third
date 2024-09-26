package controllers

import (
	"fmt"
	"log"
	"net/http"
	"third/models"
)

type Users struct {
	UserService *models.UserService
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Fprintf(w, "email: %s \n pass: %s", email, password)
}

func (u Users) ProcessSignUp(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Create(email, password)
	if err != nil {
		log.Printf("ProcessSignUp: %v", err)
		fmt.Fprint(w, "Unable to parse form submissions")
		return
	}
	fmt.Fprint(w, user)
}
