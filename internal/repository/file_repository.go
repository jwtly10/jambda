package repository

import (
	"database/sql"
)

type IFileRepository interface {
	InitFileMetaData(externalId string) error
}

type FileRepository struct {
	DB *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{DB: db}
}

func (repo *FileRepository) InitFileMetaData(externalId string) error {
	query := `INSERT INTO files (external_id, state) VALUES ($1, $2)`
	_, err := repo.DB.Exec(query, externalId, "ACTIVE")
	return err
}
