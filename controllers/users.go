package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/z-wentao/PhotoShare/context"
	"github.com/z-wentao/PhotoShare/models"
)

type Users struct {
    Templates struct {
	New    Template
	SignIn Template
	ForgotPassword Template
	CheckYourEmail Template
	ResetPassword Template
    }
    UserService    *models.UserService
    SessionService *models.SessionService
    PasswordResetService *models.PasswordRestService
    EmailService *models.EmailService
}

type UserMiddleware struct {
    SessionService *models.SessionService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//TODO: Add logic for the SetUser middleware, then call next.ServeHTTP(w, r)
	token, err := readCookie(r, CookieSession)
	if err != nil {
	    next.ServeHTTP(w, r)
	    return
	}

	user, err := umw.SessionService.User(token)
	if err != nil {
	    next.ServeHTTP(w, r)
	    return
	}

	ctx := r.Context()
	ctx = context.WithUser(ctx, user)
	r = r.WithContext(ctx)
	next.ServeHTTP(w, r)
    })
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	if user == nil {
	    http.Redirect(w, r, "/signin", http.StatusFound)
	    return
	}
	next.ServeHTTP(w, r)
    })
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
    var data struct {
	Email string
    }
    data.Email = r.FormValue("email")
    u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
    var data struct {
	Email string
    }
    data.Email = r.FormValue("email")
    u.Templates.New.Execute(w, r, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")
    password := r.FormValue("password")
    user, err := u.UserService.Create(email, password)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	return
    }
    session, err := u.SessionService.Create(user.ID)
    if err != nil {
	fmt.Println(err)
	http.Redirect(w, r, "/signin", http.StatusFound)
	return
    }
    setCookie(w, CookieSession, session.Token)
    http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
    var data struct {
	Email    string
	Password string
    }
    data.Email = r.FormValue("email")
    data.Password = r.FormValue("password")
    user, err := u.UserService.Authenticate(data.Email, data.Password)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	return
    }
    session, err := u.SessionService.Create(user.ID)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	return
    }
    setCookie(w, CookieSession, session.Token)
    http.Redirect(w, r, "/users/me", http.StatusFound)
}

// SetUser & RequireUser middleware need this 'CurrentUser' func
func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
    user := context.User(r.Context())
    if user == nil {
	http.Redirect(w, r, "/signin", http.StatusFound)
	return
    }
    fmt.Fprintf(w, "CurrentUser: %s\n", user.Email)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
    token, err := readCookie(r, CookieSession)
    if err != nil {
	http.Redirect(w, r, "/signin", http.StatusFound)
	return
    }
    err = u.SessionService.Delete(token)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	return
    }
    //	TODO: Delete the user's cookie
    deleteCookie(w, CookieSession)
    http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
    var data struct {
	Email string
    }
    data.Email = r.FormValue("email")
    u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
    var data struct {
	Email string
    }
    data.Email = r.FormValue("email")

    pwReset, err := u.PasswordResetService.Create(data.Email)
    if err != nil {
	//TODO: Handle other cases.
	//like: a user don't exist with the email address.
	fmt.Println(err)
	http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	return
    }

    vals := url.Values{
	"token": {pwReset.Token},
    }

    //TODO: make the url here configurable
    resetURL := "https://www.photoshare.com/reset-pw?" + vals.Encode()
    err = u.EmailService.ForgotPassword(data.Email, resetURL)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "something went wrong", http.StatusInternalServerError)
	return
    }

    u.Templates.ForgotPassword.Execute(w, r, data) 
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
	Token string
	Password string
    }
    data.Token = r.FormValue("token")
    data.Password = r.FormValue("password")

    user, err := u.PasswordResetService.Consume(data.Token)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "Something went wrong!", http.StatusInternalServerError)
	return
    }

    // TODO: update the user's password
    
    // sign the user in
    // redirect to the signin page
    session, err := u.SessionService.Create(user.ID)
    if err != nil {
	fmt.Println(err)
	http.Redirect(w, r, "/signin", http.StatusInternalServerError)
	return
    }

    setCookie(w, CookieSession, session.Token)
    http.Redirect(w, r, "/users/me", http.StatusInternalServerError)
}
