package models

import (
	"database/sql"
	"fmt"
	"strings"
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
	var user User
	user.Email = strings.ToLower(email)
	user.PasswordHash = password
	row := us.DB.QueryRow(`INSERT INTO users(email, password_hash)
		VALUES($1, $2) RETURNING id;`, user.Email, user.PasswordHash)
	err := row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("Create User: %v", err)
	}
	return &user, nil
}
