package controllers

import (
	"net/http"

	"github.com/z-wentao/PhotoShare/views"
)

type Static struct {
	Template views.Template
}

func StaticHandler(tpl views.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	}
}

func (static Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	static.Template.Execute(w, nil)
}
