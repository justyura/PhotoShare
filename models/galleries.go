package models

import (
	"database/sql"
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
