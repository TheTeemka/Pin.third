package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"third/context"
	"third/models"
)

type Users struct {
	Templates struct {
		CheckYourEmail template
		ResetPassword  template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
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
	user := context.User(r.Context())
	fmt.Fprint(w, user)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	passwordReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	vals := url.Values{
		"token": {passwordReset.Token},
	}

	err = u.EmailService.ForgetPassword(data.Email, "localhost:8000/reset-pw?"+vals.Encode())
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}
func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")
	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}

	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

type UserMiddleWare struct {
	SessionService *models.SessionService
}

func (umw UserMiddleWare) SetUser(f http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			token, err := readCookie(r, CookieSession)
			if err != nil {
				f.ServeHTTP(w, r)
				return
			}

			user, err := umw.SessionService.User(token)
			if err != nil {
				f.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = context.WithUser(ctx, user)
			r = r.WithContext(ctx)
			f.ServeHTTP(w, r)
		},
	)
}

func (umw UserMiddleWare) RequireUser(f http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user := context.User(ctx)
			if user == nil {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}
			f.ServeHTTP(w, r)
		},
	)
}
