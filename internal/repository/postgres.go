package repository

import (
	"fmt"

	"github.com/BountyM/effectiveMobileTestTask/internal/config"
	"github.com/jmoiron/sqlx"
)

func NewPostgresDB(cfg config.DB) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Dbname, cfg.Password, cfg.Sslmode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
