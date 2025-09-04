package controllers

import (
	"net/http"

	"github.com/z-wentao/PhotoShare/models"
)

type Galleries struct {
    Templates struct {
	New Template
    } 
    GalleryService *models.GalleryService
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
    var data struct {
	Title string
    }
    data.Title = r.FormValue("title")
    g.Templates.New.Execute(w, r, data)
}
