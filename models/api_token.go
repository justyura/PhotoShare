package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/z-wentao/PhotoShare/rand"
)

type APIToken struct {
	ID        int
	UserID    int
	GalleryID int
	Token     string
	TokenHash string
}

type APITokenService struct {
	DB            *sql.DB
	BytesPerToken int
}

func (s *APITokenService) Create(userID, galleryID int) (*APIToken, error) {
	bytesPerToken := s.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create api token %w", err)
	}
	apiToken := APIToken{
		UserID:    userID,
		GalleryID: galleryID,
		Token:     token,
		TokenHash: s.hash(token),
	}
	row := s.DB.QueryRow(`
		INSERT INTO api_tokens (user_id, gallery_id, token_hash)
		VALUES ($1, $2, $3) RETURNING id; `, userID, galleryID, apiToken.TokenHash)
	err = row.Scan(&apiToken.ID)
	if err != nil {
		return nil, fmt.Errorf("create api token %w", err)
	}
	return &apiToken, nil
}

func (s *APITokenService) ByToken(token string) (*APIToken, error) {
	apiToken := APIToken{
		TokenHash: s.hash(token),
	}
	row := s.DB.QueryRow(`
		SELECT id, user_id, gallery_id
		FROM api_tokens
		WHERE token_hash = $1; `, apiToken.TokenHash)
	err := row.Scan(&apiToken.ID, &apiToken.UserID, &apiToken.GalleryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("api token by token: %w", err)
	}
	return &apiToken, nil
}

func (s *APITokenService) ByGalleryID(galleryID int) ([]APIToken, error) {
	rows, err := s.DB.Query(`
		SELECT id, user_id, gallery_id
		FROM api_tokens
		WHERE gallery_id = $1
		ORDER BY id DESC; `, galleryID)
	if err != nil {
		return nil, fmt.Errorf("api tokens by gallery: %w", err)
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		err = rows.Scan(&t.ID, &t.UserID, &t.GalleryID)
		if err != nil {
			return nil, fmt.Errorf("api tokens by gallery: %w", err)
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("api tokens by gallery: %w", err)
	}
	return tokens, nil
}

func (s *APITokenService) Delete(id, userID int) error {
	_, err := s.DB.Exec(`
		DELETE FROM api_tokens
		WHERE id = $1 AND user_id = $2; `, id, userID)
	if err != nil {
		return fmt.Errorf("delete api token: %w", err)
	}
	return nil
}

func (s *APITokenService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
