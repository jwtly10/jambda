// db.go
package db

import (
	"database/sql"
	"fmt"

	"github.com/jwtly10/jambda/config"
	_ "github.com/lib/pq"
)

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SaveFileMetadata(db *sql.DB, filename string, filepath string) error {
	query := `INSERT INTO files (filename, filepath) VALUES ($1, $2)`
	_, err := db.Exec(query, filename, filepath)
	return err
}
