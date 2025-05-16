package controllers

import (
	"net/http"

	"github.com/z-wentao/PhotoShare/views"
)

type Users struct {
	Template struct {
		New views.Template
	}
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	//TODO: render the sign up page
}
