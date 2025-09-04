package models

import (
	"database/sql"
	"errors"
	"fmt"
)

type Gallery struct {
    ID int
    UserID int
    Title string
}

type GalleryService struct {
    DB *sql.DB
}

func (service *GalleryService) Create(title string, userID int) (*Gallery, error) {
    gallery := Gallery{
	UserID: userID,
	Title: title,
    }

    row := service.DB.QueryRow(`
	INSERT INTO galleries (title, user_id)
	VALUE ($1, $2) RETURNING id;`, title, userID)
    err := row.Scan(&gallery.ID)
    if err != nil {
	return nil, fmt.Errorf("create gallery: %w", err)
    }
    return &gallery, nil
} 

func (service *GalleryService) ByID(id int) (*Gallery, error) {
    gallery := Gallery{
	ID: id,
    }
    row := service.DB.QueryRow(`
	SELECT title, user_id
	FROM galleries
	WHERE id = $1;
	`, gallery.ID)
    err := row.Scan(&gallery.Title, &gallery.UserID)
    if err != nil {
	if errors.Is(err, sql.ErrNoRows) {
	    return nil, ErrNotFound
	}
	return nil, fmt.Errorf("query gallery by id %w", err)
    }
    return &gallery, nil
}
