package main

import (
	"fmt"
	"log"
	"net/http"
	"third/controllers"
	"third/models"
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
	cfg := models.DefaultConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userControllers := controllers.Users{
		UserService: &models.UserService{
			DB: db,
		},
		SessionService: &models.SessionService{
			DB: db,
		},
	}
	err = models.Migrate(db, "migrations")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "home.gohtml"))))
	r.Get("/faq", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "faq.gohtml"))))
	r.Get("/signup", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "signup.gohtml"))))
	r.Get("/signin", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "signin.gohtml"))))

	r.Post("/signup", userControllers.ProcessSignUp)
	r.Post("/signin", userControllers.ProcessSignIn)
	r.Get("/signout", userControllers.ProcessSignOut)
	r.Get("/me", userControllers.CurrentUser)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Page Not Found 404")
	})

	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		csrf.Secure(false),
	)
	fmt.Println("Server is starting :3000 ....")
	http.ListenAndServe(":8000", csrfMw(tim(r)))
}
