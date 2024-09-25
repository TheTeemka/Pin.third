package main

import (
	"fmt"
	"log"
	"net/http"
	"third/controllers"
	"third/templates"
	"third/views"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Shit, here we go again")
	r := chi.NewRouter()
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "home.gohtml"))))
	r.Get("/faq", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "faq.gohtml"))))

	r.Get("/signup", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "signup.gohtml"))))
	r.Get("/signin", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "signin.gohtml"))))

	log.Println("Server is starting :3000 ....")
	http.ListenAndServe(":8000", r)
}
