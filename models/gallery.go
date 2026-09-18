package models

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/z-wentao/PhotoShare/rand"
)

const galleryShareIDBytes = 16

type Gallery struct {
	ID       int
	UserID   int
	Title    string
	IsPublic bool
	ShareID  string
}

type GalleryService struct {
	DB *sql.DB
}

func (service *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		UserID: userID,
		Title:  title,
	}

	row := service.DB.QueryRow(`
	INSERT INTO galleries (title, user_id)
	VALUES ($1, $2) RETURNING id;`, title, userID)
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
	SELECT title, user_id, is_public, COALESCE(share_id, '')
	FROM galleries
	WHERE id = $1;
	`, gallery.ID)
	err := row.Scan(&gallery.Title, &gallery.UserID, &gallery.IsPublic, &gallery.ShareID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallery by id %w", err)
	}
	return &gallery, nil
}

// ByShareID only returns galleries whose link sharing is currently enabled.
// A disabled share link therefore looks the same as a link that never existed.
func (service *GalleryService) ByShareID(shareID string) (*Gallery, error) {
	var gallery Gallery
	row := service.DB.QueryRow(`
	SELECT id, title, user_id, is_public, share_id
	FROM galleries
	WHERE share_id = $1 AND is_public = TRUE;
	`, shareID)
	err := row.Scan(
		&gallery.ID,
		&gallery.Title,
		&gallery.UserID,
		&gallery.IsPublic,
		&gallery.ShareID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallery by share id: %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByUserID(userID int) ([]Gallery, error) {
	rows, err := service.DB.Query(`
	SELECT id, title 
	FROM galleries
	WHERE user_id = $1;
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}
	defer rows.Close()

	var galleries []Gallery
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		err := rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, fmt.Errorf("query galleries by user: %w", err)
		}
		galleries = append(galleries, gallery)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}
	return galleries, nil
}

func (service *GalleryService) Update(gallery *Gallery) error {
	_, err := service.DB.Exec(`
	UPDATE galleries
	SET title = $2
	WHERE id = $1;
	`, gallery.ID, gallery.Title)
	if err != nil {
		return fmt.Errorf("Update galleries: %w", err)
	}
	return nil
}

// SetPublic enables or disables link sharing. The 128-bit share ID is created
// once, the first time sharing is enabled, so it cannot be sequentially
// enumerated like the gallery's internal integer ID.
func (service *GalleryService) SetPublic(gallery *Gallery, isPublic bool) error {
	shareID := gallery.ShareID
	if isPublic && shareID == "" {
		var err error
		shareID, err = rand.URLSafeString(galleryShareIDBytes)
		if err != nil {
			return fmt.Errorf("create gallery share id: %w", err)
		}
	}

	_, err := service.DB.Exec(`
	UPDATE galleries
	SET is_public = $2, share_id = $3
	WHERE id = $1;
	`, gallery.ID, isPublic, nullableString(shareID))
	if err != nil {
		return fmt.Errorf("set gallery sharing: %w", err)
	}

	gallery.IsPublic = isPublic
	gallery.ShareID = shareID
	return nil
}

func (service *GalleryService) Delete(id int) error {
	_, err := service.DB.Exec(`
	DELETE FROM galleries
	WHERE id = $1;
	`, id)
	if err != nil {
		return fmt.Errorf("Delete gallery by id: %w", err)
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func hasExtension(file string, extensions []string) bool {
	for _, ext := range extensions {
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}
