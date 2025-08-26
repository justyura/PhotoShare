package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/z-wentao/PhotoShare/rand"
)

type PasswordRest struct {
    ID int
    UserID int
    // Token is only set when a PasswordRest is being created.
    Token string
    TokenHash string
    ExpiresAt time.Time
}

type NullTime struct {
    Time time.Time
    Valid bool // Valid is true when Time is not NULL 
}

type PasswordResetService struct {
    DB *sql.DB
    // similar as session
    BytesPerToken int 
    // Duration is the amount of time that a PasswordRest(Token) valid for
    Duration time.Duration
}

const (
    DefaultResetDuration = 1 * time.Hour
)

func (service *PasswordResetService) hash(token string) string {
    tokenHash := sha256.Sum256([]byte(token))
    return base64.RawURLEncoding.EncodeToString(tokenHash[:])
}

func (service *PasswordResetService) Create(email string) (*PasswordRest, error) {
    email = strings.ToLower(email)
    var userID int
    row := service.DB.QueryRow(`
	SELECT id FROM users where email = $1
	`, email)
    err := row.Scan(&userID)
    if err != nil {
	// TODO: Consider returning a specific error when the user does not exist.
	return nil, fmt.Errorf("create: %w", err)
    }

    // build the passwordreset
    bytesPerToken := service.BytesPerToken
    if bytesPerToken == 0 {
	bytesPerToken = MinBytesPerToken
    }
    token, err := rand.String(bytesPerToken)
    if err != nil {
	return nil, fmt.Errorf("create: %w", err)
    }

    duration := service.Duration
    if duration == 0 {
	duration = DefaultResetDuration
    }

    pwReset := PasswordRest {
	UserID: userID,
	Token: token,
	TokenHash: service.hash(token),
	ExpiresAt: time.Now().Add(duration),
    }

    //Insert the passwordreset into DB
    row = service.DB.QueryRow(`
	INSERT INTO password_resets (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3) ON CONFLICT (user_id) DO
	UPDATE
	SET token_hash = $2, expires_at = $3
	RETURNING id;
	`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
    err = row.Scan(&pwReset.ID)
    if err != nil {
	return nil, fmt.Errorf("create: %w", err)
    }

    return &pwReset, nil
}

func (service *PasswordResetService) Consume(token string) (*User, error) {
    tokenHash := service.hash(token)
    var user User
    var pwReset PasswordRest 
    row := service.DB.QueryRow(`
	SELECT password_resets.id, password_resets.expires_at, users.id, users.email, users.password_hash
	FROM password_resets JOIN users ON users.id = password_resets.user_id
	WHERE password_resets.token_hash = $1;
	`, tokenHash)
    err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
    if err != nil {
	return nil, fmt.Errorf("consume: %w", err)
    }
    if time.Now().After(pwReset.ExpiresAt) {
	return nil, fmt.Errorf("token expired: %v", token)
    }

    err = service.delete(pwReset.ID)
    if err != nil {
	return nil, fmt.Errorf("consume: %w", err)
    }
    return &user, nil
}

func (service *PasswordResetService) delete(id int) error {
    _, err := service.DB.Exec(`
	DELETE FROM password_resets
	WHERE id = $1;
	`, id)
    if err != nil {
	return fmt.Errorf("delete: %w", err)
    }
    return nil
}
