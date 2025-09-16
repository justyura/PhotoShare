package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/z-wentao/PhotoShare/context"
	"github.com/z-wentao/PhotoShare/models"
)

type Galleries struct {
    Templates struct {
	Show Template
	New Template
	Edit Template
	Index Template
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

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
    gallery, err := g.galleryByID(w, r, userMustOwnGallery)
    if err != nil {
	return
    }

    data := struct {
	ID int
	Title string
    } {
	ID: gallery.ID,
	Title: gallery.Title,
    }
    g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
    gallery, err := g.galleryByID(w, r, userMustOwnGallery)
    if err != nil {
	return
    }

    // parse the updated info and update
    title := r.FormValue("title")
    gallery.Title = title
    err = g.GalleryService.Update(gallery)
    if err != nil {
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	return
    }
    editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
    http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
    type Gallery struct {
	ID int
	Title string
    }
    var data struct {
	Galleries []Gallery
    }

    user := context.User(r.Context())
    galleries, err := g.GalleryService.ByUserID(user.ID)
    if err != nil {
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	return
    }
    for _, gallery := range galleries {
	data.Galleries = append(data.Galleries, Gallery{
	    ID: gallery.ID,
	    Title: gallery.Title,
	})
    }

    g.Templates.Index.Execute(w, r, data)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
    gallery, err := g.galleryByID(w, r)
    if err != nil {
	return
    }
    type Image struct {
	GalleryID int
	Filename string
    }

    var data struct {
	ID int
	Title string
	Images []Image
    }
    data.ID = gallery.ID
    data.Title = gallery.Title
    images, err := g.GalleryService.Images(gallery.ID)
    if err != nil {
	fmt.Println(err)
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	return
    }
    for _, image := range images {
	data.Images = append(data.Images, Image{
	    GalleryID: image.GalleryID,
	    Filename: image.Filename,
	})
    }
    g.Templates.Show.Execute(w, r, data)
}

type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

func (g Galleries) galleryByID (w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
    id, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil {
	http.Error(w, "Invalid ID", http.StatusNotFound)
	return nil, err
    }
    gallery, err := g.GalleryService.ByID(id)
    if err != nil {
	if errors.Is(err, models.ErrNotFound) {
	    http.Error(w, "Gallery not found", http.StatusNotFound)
	    return nil, err
	}
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	return nil, err
    }
    for _, opt := range opts {
	err = opt(w, r, gallery)
	if err != nil {
	    return nil, err
	}
    }
    return gallery, nil
}

func userMustOwnGallery (w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
    user := context.User(r.Context())
    if user.ID != gallery.UserID {
	http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
	return fmt.Errorf("user does not have access to this gallery")
    }
    return nil
}

func (g Galleries) Delete (w http.ResponseWriter, r *http.Request) {
    gallery, err := g.galleryByID(w, r, userMustOwnGallery)
    if err != nil {
	return
    }
    err = g.GalleryService.Delete(gallery.ID)
    if err != nil {
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	return
    }
    http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
    filename := chi.URLParam(r, "filename")
    galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil {
	http.Error(w, "Invalid ID", http.StatusNotFound)
	return
    }
    image, err := g.GalleryService.Image(galleryID, filename)
    if err != nil {
	if errors.Is(err, models.ErrNotFound) {
	    http.Error(w, "Image not found", http.StatusNotFound)
	    return
	}
	fmt.Println(err)
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	return
    }

    http.ServeFile(w, r, image.Path)
}
