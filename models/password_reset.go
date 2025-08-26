package models

import (
	"database/sql"
	"fmt"
	"time"
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

type PasswordRestService struct {
    DB *sql.DB
    // similar as session
    BytesPerToken int 
    // Duration is the amount of time that a PasswordRest(Token) valid for
    Duration time.Duration
}

const (
    DefaultResetDuration = 1 * time.Hour
)

func (service *PasswordRestService) Create(email string) (*PasswordRest, error) {
    return nil, fmt.Errorf("TODO: Implement the PasswordRestService.Create") 
}

func (service *PasswordRestService) Consume(token string) (*User, error) {
    return nil, fmt.Errorf("TODO: Implement the PasswordRestService.Consume") 
}
