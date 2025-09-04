package views

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/z-wentao/PhotoShare/context"

	"github.com/gorilla/csrf"
	"github.com/z-wentao/PhotoShare/models"
)

type Template struct {
	htmlTpl *template.Template
}

type public interface {
    Public() string 
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
    tpl := template.New(filepath.Base(pattern[0]))
    tpl = tpl.Funcs(template.FuncMap{
	"csrfField": func() (template.HTML, error) {
	    return "", fmt.Errorf("csrfField not implemented")
	},
	"currentUser": func() (*models.User, error) {
	    return nil, fmt.Errorf("currentUser not implemented")
	},
	"errors": func() []string {
	    return nil
	},
    })
    tpl, err := tpl.ParseFS(fs, pattern...)
    if err != nil {
	return Template{}, fmt.Errorf("parsing template: %w", err)
    }
    return Template{
	htmlTpl: tpl,
    }, nil

}

func errMessages(errs ...error) []string {
    var msgs []string
    for _, err := range errs {
	var pubErr public
	if errors.As(err, &pubErr) {
	    msgs = append(msgs, pubErr.Public())
	} else {
	    fmt.Println(err)
	    msgs = append(msgs, "Something went wrong.")
	}
    }
    return msgs
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error) {
    tpl, err := t.htmlTpl.Clone()
    if err != nil {
	log.Printf("Error cloning template: %v", err)
	http.Error(w, "Internal Server Error while rendering the page", http.StatusInternalServerError)
	return
    }
    errMsgs := errMessages(errs...)
    tpl = tpl.Funcs(template.FuncMap{
	"csrfField": func() template.HTML {
	    return csrf.TemplateField(r)
	},
	"currentUser": func() *models.User {
	    return context.User(r.Context()) 
	},
	"errors": func() []string {
	    return errMsgs
	},
    })
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = tpl.Execute(w, data)
    if err != nil {
	log.Printf("executing template: %v", err)
	http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
	return
    }
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}
