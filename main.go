package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
	"github.com/z-wentao/PhotoShare/controllers"
	"github.com/z-wentao/PhotoShare/migrations"
	"github.com/z-wentao/PhotoShare/models"
	"github.com/z-wentao/PhotoShare/templates"
	"github.com/z-wentao/PhotoShare/views"
)

type config struct {
    PSQL models.PostgresConfig
    SMTP models.SMTPConfig
    CSRF struct {
	Key string
	Secure bool
    }
    Server struct {
	Address string
    }
}

func loadENVConfig() (config, error) {
    var cfg config
    err := godotenv.Load()
    if err != nil {
	return cfg, err
    } 
    //TODO: setup PSQL config
    cfg.PSQL = models.DefaultPostgresConfig()
    //TODO: setup SMTP config
    cfg.SMTP.Host = os.Getenv("SMTP_HOST")
    portStr := os.Getenv("SMTP_PORT")
    cfg.SMTP.Port, err = strconv.Atoi(portStr)
    if err != nil {
	return cfg, err
    }
    cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
    cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
    //TODO: setup CSRF config
    cfg.CSRF.Key = "abc32abcdefghijklmnopqrstuvwxyz1"
    cfg.CSRF.Secure = false
    //TODO: setup Server config
    cfg.Server.Address = ":3000"
    
    return cfg, nil
}

func main() {
    cfg, err := loadENVConfig()
    if err != nil {
	panic(err)
    }

    db, err := models.Open(cfg.PSQL)
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
    userService := &models.UserService{
	DB: db,
    }

    sessionService := &models.SessionService{
	DB: db,
    }

    pwResetService := &models.PasswordRestService{
	DB: db,
    }

    emailService := models.NewEmailService(cfg.SMTP)
    
    // Set up middlewares
    umw := controllers.UserMiddleware{
	SessionService: sessionService,
    }

    csrfMw := csrf.Protect([]byte(cfg.CSRF.Key), csrf.Secure(cfg.CSRF.Secure))

    // Set up controllers
    usersC := controllers.Users{
	UserService:    userService,
	SessionService: sessionService,
	PasswordResetService: pwResetService,
	EmailService: emailService,
    }

    usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
    usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
    usersC.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS, "forgot-pw.gohtml", "tailwind.gohtml"))
    usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(templates.FS, "check-your-email.gohtml", "tailwind.gohtml"))
    usersC.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS, "reset-pw.gohtml", "tailwind.gohtml"))

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
    r.Get("/forgot-pw", usersC.ForgotPassword)
    r.Post("/forgot-pw", usersC.ProcessForgotPassword)
    r.Get("/reset-pw", usersC.ResetPassword)
    r.Post("/reset-pw", usersC.ProcessResetPassword) 

    // Start the server
    fmt.Printf("Starting the server on %s...\n", cfg.Server.Address)
    http.ListenAndServe(cfg.Server.Address, r)
    if err != nil {
	panic(err)
    }
}
