package models

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLmode  string
}

func DefaultConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     "5433",
		User:     "ol",
		Password: "ol",
		Database: "olar",
		SSLmode:  "disable",
	}
}

func (p PostgresConfig) ToString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLmode)
}

func Open(config PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.ToString())
	if err != nil {
		return nil, fmt.Errorf("DB open: %v", err)
	}
	return db, nil
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("Migrate db: %v", err)
	}
	err = goose.Up(db, dir)

	if err != nil {
		return fmt.Errorf("Migrate db: %v", err)
	}
	return nil
}