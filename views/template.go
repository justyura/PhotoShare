package views

import (
	"github.com/z-wentao/PhotoShare/context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/z-wentao/PhotoShare/models"
)

type Template struct {
	htmlTpl *template.Template
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
    tpl := template.New(pattern[0])
    tpl = tpl.Funcs(template.FuncMap{
	"csrfField": func() (template.HTML, error) {
	    return "", fmt.Errorf("csrfField not implemented")
	},
	"currentUser": func() (*models.User, error) {
	    return nil, fmt.Errorf("currentUser not implemented")
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

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any) {
    tpl, err := t.htmlTpl.Clone()
    if err != nil {
	log.Printf("Error cloning template: %v", err)
	http.Error(w, "Internal Server Error while rendering the page", http.StatusInternalServerError)
	return
    }
    tpl = tpl.Funcs(template.FuncMap{
	"csrfField": func() template.HTML {
	    return csrf.TemplateField(r)
	},
	"currentUser": func() *models.User {
	    return context.User(r.Context()) 
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
