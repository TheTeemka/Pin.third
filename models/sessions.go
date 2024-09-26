package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"third/rand"
)

const (
	minTokenBytes = 32
)

type Session struct {
	ID     int
	UserID int
	//Token is set only when creating token, at other times it is empty
	Token     string
	TokenHash string
}

type SessionService struct {
	DB            *sql.DB
	BytesPerToken int
}

func (ss *SessionService) Create(userId int) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < minTokenBytes {
		bytesPerToken = minTokenBytes
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("Create : %v", err)
	}
	tokenHash := ss.hash(token)
	session := Session{
		UserID:    userId,
		Token:     token,
		TokenHash: tokenHash,
	}
	row := ss.DB.QueryRow(`INSERT INTO sessions(user_id, token_hash)
	VALUES($1, $2) ON CONFLICT (user_id)
	DO UPDATE 
	SET token_hash = $2
	RETURNING id;`, session.UserID, session.TokenHash)

	err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("Create : %v", err)
	}

	return &session, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	_, err := ss.DB.Exec(`DELETE FROM sessions
		WHERE token_hash = $1`, tokenHash)

	if err != nil {
		return fmt.Errorf("delete cookie: %v", err)
	}
	return nil
}
func (ss *SessionService) User(token string) (*User, error) {
	tokenHash := ss.hash(token)
	row := ss.DB.QueryRow(`SELECT users.id, users.email, users.password_hash
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1;`, tokenHash)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %v", err)
	}

	return &user, nil
}

func (ss *SessionService) hash(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hashedToken[:])
}
