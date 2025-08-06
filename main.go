package main

import (
	"fmt"
	"github.com/gorilla/csrf"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/z-wentao/PhotoShare/controllers"
	"github.com/z-wentao/PhotoShare/models"
	"github.com/z-wentao/PhotoShare/templates"
	"github.com/z-wentao/PhotoShare/views"
)

func main() {
	r := chi.NewRouter()
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	userService := models.UserService{
		DB: db,
	}

	sessionService := models.SessionService{
		DB: db,
	}

	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}

	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Get("/users/me", usersC.CurrentUser)
	r.Post("/signout", usersC.ProcessSignOut)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found!", http.StatusNotFound)
	})

	csrfKey := "abc32abcdefghijklmnopqrstuvwxyz1"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false))
	fmt.Println("Starting the server on :3000...")

	http.ListenAndServe(":3000", csrfMw(r))
}
