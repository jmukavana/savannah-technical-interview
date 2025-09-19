package Database

import (	
	"time"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

var DB *gorm.DB

func NewPostgres(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// ping
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
