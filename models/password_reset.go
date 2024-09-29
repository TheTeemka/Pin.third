package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"third/rand"
	"time"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID        int
	UserID    int
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB            *sql.DB
	BytesPerToken int
	Duration      time.Duration
}

func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {
	email = strings.ToLower(email)
	row := service.DB.QueryRow(`
		SELECT id
		FROM users
		WHERE email = $1`, email)
	var userID int
	err := row.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create: %v", err)
	}

	bytesPerToken := service.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = minTokenBytes
	}

	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %v", err)
	}

	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}

	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = service.DB.QueryRow(`
		INSERT INTO password_resets(user_id, token_hash, expires_at)
		VALUES($1, $2, $3) ON CONFLICT (user_id) DO
		UPDATE SET token_hash = $2, expires_at = $3
		RETURNING id`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %v", err)
	}
	return &pwReset, nil
}

func (service *PasswordResetService) Consume(token string) (*User, error) {
	tokenHash := service.hash(token)
	row := service.DB.QueryRow(`
		SELECT password_resets.id ,
			password_resets.expires_at,
			users.id,
			users.email,
			users.password_hash
		FROM password_resets
			JOIN users ON users.id = password_resets.user_id
		WHERE password_resets.token_hash = $1;`, tokenHash)

	var user User
	var password_reset PasswordReset
	err := row.Scan(&password_reset.ID, &password_reset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %v ", err)
	}

	if time.Now().After(password_reset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", token)
	}

	err = service.delete(password_reset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %v ", err)
	}
	return &user, nil
}

func (service *PasswordResetService) delete(id int) error {
	_, err := service.DB.Exec(`
		DELETE FROM password_resets
		WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("Delete: %v", err)
	}
	return nil
}

func (service *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
