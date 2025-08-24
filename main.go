package main

import (
    "fmt"
    "net/http"

    "github.com/gorilla/csrf"

    "github.com/go-chi/chi/v5"
    "github.com/z-wentao/PhotoShare/controllers"
    "github.com/z-wentao/PhotoShare/migrations"
    "github.com/z-wentao/PhotoShare/models"
    "github.com/z-wentao/PhotoShare/templates"
    "github.com/z-wentao/PhotoShare/views"
)

func main() {
    // Set up the DB connection
    cfg := models.DefaultPostgresConfig()
    db, err := models.Open(cfg)
    if err != nil {
	panic(err)
    }
    defer db.Close()

    // use the embed migration files
    err = models.MigrateFS(db, migrations.FS, ".")
    if err != nil {
	panic(err)
    }

    // Set up services
    userService := models.UserService{
	DB: db,
    }

    sessionService := models.SessionService{
	DB: db,
    }
    
    // Set up middlewares
    umw := controllers.UserMiddleware{
	SessionService: &sessionService,
    }

    csrfKey := "abc32abcdefghijklmnopqrstuvwxyz1"
    csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false))

    // Set up controllers
    usersC := controllers.Users{
	UserService:    &userService,
	SessionService: &sessionService,
    }

    usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
    usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))

    // Set up router and routes
    r := chi.NewRouter()

    r.Use(csrfMw)
    r.Use(umw.SetUser)
    r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
    r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
    r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
    r.Get("/signup", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))))
    r.Get("/signup", usersC.New)
    r.Post("/signup", usersC.Create)
    r.Get("/signin", usersC.SignIn)
    r.Post("/signin", usersC.ProcessSignIn)
    r.Route("/users/me", func(r chi.Router) {
	r.Use(umw.RequireUser)
	r.Get("/", usersC.CurrentUser)
    })
    r.Post("/signout", usersC.ProcessSignOut)
    r.NotFound(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Page not found!", http.StatusNotFound)
    })

    // Start the server
    fmt.Println("Starting the server on :3000...")
    http.ListenAndServe(":3000", r)
}
