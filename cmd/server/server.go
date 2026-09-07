package main

import (
	"errors"
	"fmt"
	"io/fs"
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
		Key    string
		Secure bool
	}
	Server struct {
		Address   string
		BaseURL   string
		ImagesDir string
	}
}

func loadENVConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return cfg, fmt.Errorf("load .env: %w", err)
	}
	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("no psql config provided")
	}
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"

	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	cfg.Server.BaseURL = os.Getenv("BASE_URL")
	cfg.Server.ImagesDir = os.Getenv("IMAGES_DIR")

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

	pwResetService := &models.PasswordResetService{
		DB: db,
	}

	gallerySevice := &models.GalleryService{
		DB: db,
	}

	imageService := &models.ImageService{
		DB:        db,
		ImagesDir: cfg.Server.ImagesDir,
	}

	apiTokenService := &models.APITokenService{
		DB: db,
	}

	emailService := models.NewEmailService(cfg.SMTP)

	// Set up middlewares
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	csrfMw := csrf.Protect([]byte(cfg.CSRF.Key), csrf.Secure(cfg.CSRF.Secure), csrf.Path("/"))

	// Set up controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
	}

	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS, "forgot-pw.gohtml", "tailwind.gohtml"))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(templates.FS, "check-your-email.gohtml", "tailwind.gohtml"))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS, "reset-pw.gohtml", "tailwind.gohtml"))

	galleriesC := controllers.Galleries{
		GalleryService:  gallerySevice,
		ImageService:    imageService,
		APITokenService: apiTokenService,
		BaseURL:         cfg.Server.BaseURL,
	}
	galleriesC.Templates.New = views.Must(views.ParseFS(templates.FS, "galleries/new.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(templates.FS, "galleries/edit.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Index = views.Must(views.ParseFS(templates.FS, "galleries/index.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Show = views.Must(views.ParseFS(templates.FS, "galleries/show.gohtml", "tailwind.gohtml"))

	imagesC := controllers.Images{
		ImageService:   imageService,
		GalleryService: gallerySevice,
	}

	apiC := controllers.API{
		APITokenService: apiTokenService,
		ImageService:    imageService,
		BaseURL:         cfg.Server.BaseURL,
	}

	// Set up router and routes
	r := chi.NewRouter()

	// Public direct links: no session, no CSRF. Kept outside the web group so
	// that serving an image does not pay for a session lookup.
	r.Get("/i/{key}", imagesC.Show)

	// The API branch. RequireToken reads the Authorization header and never
	// touches the session cookie — if it did, browsers would attach that
	// cookie automatically and the CSRF hole would be right back.
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apiC.RequireToken)
		r.Post("/upload", apiC.Upload)
	})

	// The web branch: session cookie plus CSRF protection.
	r.Group(func(r chi.Router) {
		r.Use(csrfMw)
		r.Use(umw.SetUser)

		r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
		r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
		r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
		r.Get("/signup", usersC.New)
		r.Post("/signup", usersC.Create)
		r.Get("/signin", usersC.SignIn)
		r.Post("/signin", usersC.ProcessSignIn)
		r.Route("/users/me", func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/", usersC.CurrentUser)
		})
		r.Post("/signout", usersC.ProcessSignOut)
		r.Get("/forgot-pw", usersC.ForgotPassword)
		r.Post("/forgot-pw", usersC.ProcessForgotPassword)
		r.Get("/reset-pw", usersC.ResetPassword)
		r.Post("/reset-pw", usersC.ProcessResetPassword)

		r.Route("/galleries", func(r chi.Router) {
			r.Get("/{id}", galleriesC.Show)
			r.Group(func(r chi.Router) {
				r.Use(umw.RequireUser)
				r.Get("/", galleriesC.Index)
				r.Get("/new", galleriesC.New)
				r.Post("/", galleriesC.Create)
				r.Get("/{id}/edit", galleriesC.Edit)
				r.Post("/{id}", galleriesC.Update)
				r.Post("/{id}/delete", galleriesC.Delete)
				r.Post("/{id}/images", galleriesC.UploadImage)
				r.Post("/{id}/tokens", galleriesC.CreateToken)
				r.Post("/{id}/tokens/{tokenID}/delete", galleriesC.DeleteToken)
			})
		})

		// An image is identified by its uuid alone, so deleting one does not
		// need to know which gallery it sits in.
		r.Route("/images", func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Post("/{key}/delete", imagesC.Delete)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found!", http.StatusNotFound)
	})

	// Start the server
	fmt.Printf("Starting the server on %s...\n", cfg.Server.Address)
	http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}
