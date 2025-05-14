package controllers

import (
	"html/template"
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

func FAQ(tpl views.Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   template.HTML
	}{
		{
			Question: "What is this website about?",
			Answer:   "As it's named, it help sharing photo with your family and friend.",
		},
		{
			Question: "How to use it",
			Answer:   "Sign up and log in, upload your photo and share!",
		},
		{
			Question: "How to get support?",
			Answer:   `Email me - <a href="mailto:support@photoshare.com">support@photoshare.com</a>`,
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, questions)
	}
}
