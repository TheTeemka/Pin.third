package controllers

import (
	"fmt"
	"log"
	"net/http"
	"third/models"
)

type Users struct {
	UserService    *models.UserService
	SessionService *models.SessionService
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Authenticate(email, password)
	if err != nil {
		log.Printf("ProcessSignIn: %v", err)
		fmt.Fprint(w, "Unable to parse form submissions")
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("ProcessSignIn: %v", err)
		fmt.Fprint(w, "Unable to parse form submissions")
		return
	}
	setCookie(w, CookieSession, session.Token)
	fmt.Fprint(w, user)
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

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("ProcessSignUp: %v", err)
		fmt.Fprint(w, "Unable to parse form submissions")
		return
	}
	setCookie(w, CookieSession, session.Token)
	fmt.Fprint(w, user)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	err = u.SessionService.Delete(token)
	if err != nil {
		http.Redirect(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}
	deleteCookie(w, CookieSession)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {

		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	user, err := u.SessionService.User(token)
	if err != nil {

		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	fmt.Fprint(w, user)
}
