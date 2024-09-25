package main

import (
	"fmt"
	"log"
	"net/http"
	"third/controllers"
	"third/templates"
	"third/views"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func tim(f http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%-15s%-15s%-5s", r.RemoteAddr, r.URL.Path, r.Method)
			f.ServeHTTP(w, r)
		},
	)
}
func main() {
	fmt.Println("Shit, here we go again")
	r := chi.NewRouter()
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "home.gohtml"))))
	r.Get("/faq", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "faq.gohtml"))))
	r.Get("/signup", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "signup.gohtml"))))
	r.Get("/signin", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "signin.gohtml"))))

	r.Post("/signup", controllers.ProcessSignUp)
	r.Post("/signin", controllers.ProcessSignIn)
	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		csrf.Secure(false),
	)
	fmt.Println("Server is starting :3000 ....")
	http.ListenAndServe(":8000", csrfMw(tim(r)))
}
