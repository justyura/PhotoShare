package controllers

import (
	"fmt"
	"net/http"

	"github.com/z-wentao/PhotoShare/context"
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

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
    var data struct {
	UserID int
	Title string
    }
    data.UserID = context.User(r.Context()).ID
    data.Title = r.FormValue("title")

    gallery, err := g.GalleryService.Create(data.Title, data.UserID)
    if err != nil {
	g.Templates.New.Execute(w, r, data, err)
	return
    }
    editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
    http.Redirect(w, r, editPath, http.StatusFound)
}
