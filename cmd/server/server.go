package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"third/controllers"
	"third/models"
	"third/templates"
	"third/views"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, nil
	}
	//TODO: READ FROM ENV
	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLmode:  os.Getenv("PSQL_SSL_MODE"),
	}

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}

	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	// TODO: Read the CSRF values from an ENV variable
	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"

	// TODO: Read the server values from an ENV variable
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	return cfg, nil
}

func tim(f http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%-20s %-5s", r.URL.Path, r.Method)
			f.ServeHTTP(w, r)
		},
	)
}
func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	passwordResetService := &models.PasswordResetService{
		DB: db,
	}
	galleryService := &models.GalleryService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)
	userControllers := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: passwordResetService,
		EmailService:         emailService,
	}
	umw := controllers.UserMiddleWare{
		SessionService: sessionService,
	}
	galleryControllers := controllers.Galleries{
		GalleryService: galleryService,
	}
	err = models.Migrate(db, "migrations")
	if err != nil {
		panic(err)
	}

	csrfKey := cfg.CSRF.Key
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"),
	)

	r := chi.NewRouter()
	r.Use(tim, csrfMw, umw.SetUser)

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "home.gohtml"))))
	r.Get("/faq", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "faq.gohtml"))))

	userControllers.Templates.SignUp = views.Must(views.ParseFS(templates.FS, "base.gohtml", "signup.gohtml"))
	r.Get("/signup", userControllers.SignUp)
	r.Post("/signup", userControllers.ProcessSignUp)

	userControllers.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "base.gohtml", "signin.gohtml"))
	r.Get("/signin", userControllers.SignIn)
	r.Post("/signin", userControllers.ProcessSignIn)

	r.Post("/signout", userControllers.ProcessSignOut)

	userControllers.Templates.CheckYourEmail = views.Must(views.ParseFS(templates.FS, "base.gohtml", "check-your-email.gohtml"))
	r.Get("/forgot-pw", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "base.gohtml", "forgot-pw.gohtml"))))
	r.Post("/forgot-pw", userControllers.ProcessForgotPassword)

	userControllers.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS, "base.gohtml", "reset-pw.gohtml"))
	r.Get("/reset-pw", userControllers.ResetPassword)
	r.Post("/reset-pw", userControllers.ProcessResetPassword)

	userControllers.Templates.CurrentUser = views.Must(views.ParseFS(templates.FS, "base.gohtml", "current-user.gohtml"))
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", userControllers.CurrentUser)
	})

	galleryControllers.Templates.New = views.Must(views.ParseFS(templates.FS, "base.gohtml", "gallery/new.gohtml"))
	galleryControllers.Templates.Edit = views.Must(views.ParseFS(templates.FS, "base.gohtml", "gallery/edit.gohtml"))
	galleryControllers.Templates.Index = views.Must(views.ParseFS(templates.FS, "base.gohtml", "gallery/index.gohtml"))
	galleryControllers.Templates.Show = views.Must(views.ParseFS(templates.FS, "base.gohtml", "gallery/show.gohtml"))
	galleryControllers.Templates.PublicGallery = views.Must(views.ParseFS(templates.FS, "base.gohtml", "gallery/public-gallery.gohtml"))
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleryControllers.Show)
		r.Get("/{id}/images/{filename}", galleryControllers.Image)
		r.Get("/common", galleryControllers.PublicGallery)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/new", galleryControllers.New)
			r.Post("/new", galleryControllers.Create)
			r.Get("/", galleryControllers.Index)
			r.Get("/{id}/edit", galleryControllers.Edit)
			r.Post("/{id}/edit", galleryControllers.ProcessEdit)
			r.Post("/{id}/delete", galleryControllers.Delete)
			r.Post("/{id}/images/{filename}/delete", galleryControllers.DeleteImage)
			r.Post("/{id}/images/upload", galleryControllers.UploadImage)

		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Page Not Found 404")
	})

	fmt.Printf("Server is starting %s ....\n", cfg.Server.Address)
	http.ListenAndServe(cfg.Server.Address, r)
}
