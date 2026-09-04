package controllers

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/z-wentao/PhotoShare/context"
	"github.com/z-wentao/PhotoShare/models"
)

type Images struct {
	ImageService   *models.ImageService
	GalleryService *models.GalleryService
}

// Show serves the public direct link, GET /i/{key}, where key looks like
// "kR3nPx7Qw2ZmLvB8dTaYcg.jpg". The extension is decoration for humans and
// link previewers; the uuid is what identifies the image.
func (i Images) Show(w http.ResponseWriter, r *http.Request) {
	img, err := i.imageByKey(w, r)
	if err != nil {
		return
	}

	contentType := mime.TypeByExtension(img.Ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// http.ServeFile only sniffs a type when we have not set one.
	w.Header().Set("Content-Type", contentType)
	// A uuid is assigned once and never reused, so the bytes behind this URL
	// can never change. That is what makes immutable caching safe here.
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	// Stop browsers from second-guessing the Content-Type we just set.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, i.ImageService.Path(*img))
}

func (i Images) Delete(w http.ResponseWriter, r *http.Request) {
	img, err := i.imageByKey(w, r)
	if err != nil {
		return
	}
	err = i.userMustOwnImage(w, r, img)
	if err != nil {
		return
	}
	err = i.ImageService.Delete(*img)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", img.GalleryID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (i Images) imageByKey(w http.ResponseWriter, r *http.Request) (*models.Image, error) {
	key := chi.URLParam(r, "key")
	uuid := strings.TrimSuffix(key, path.Ext(key))

	img, err := i.ImageService.ByUUID(uuid)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return nil, err
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, err
	}
	return img, nil
}

// userMustOwnImage costs one extra query because images do not carry a
// user_id of their own; ownership is reached through the gallery.
func (i Images) userMustOwnImage(w http.ResponseWriter, r *http.Request, img *models.Image) error {
	gallery, err := i.GalleryService.ByID(img.GalleryID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return err
	}
	user := context.User(r.Context())
	if user.ID != gallery.UserID {
		http.Error(w, "You are not authorized to delete this image", http.StatusForbidden)
		return fmt.Errorf("user does not own this image")
	}
	return nil
}
