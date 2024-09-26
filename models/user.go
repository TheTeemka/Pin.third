package models

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
}

type UserService struct {
	DB *sql.DB
}

func (us *UserService) Create(email string, password string) (*User, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Create User: %v", err)
	}

	var user User
	user.Email = strings.ToLower(email)
	user.PasswordHash = string(hashedBytes)

	row := us.DB.QueryRow(`INSERT INTO users(email, password_hash)
		VALUES($1, $2) RETURNING id;`, user.Email, user.PasswordHash)
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("Create User: %v", err)
	}

	return &user, nil
}

func (us *UserService) Authenticate(email string, password string) (*User, error) {
	var user User
	user.Email = strings.ToLower(email)
	row := us.DB.QueryRow(`
		SELECT id, password_hash 
		FROM users
		WHERE email = $1`, user.Email)

	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authentication : %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authentication : %v", err)
	}

	return &user, nil
}
