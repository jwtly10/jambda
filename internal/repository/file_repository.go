package repository

import (
	"database/sql"
	"time"
)

type FileEntity struct {
	ID         int
	ExternalId string
	State      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type IFileRepository interface {
	InitFileMetaData(externalId string) error
	GetFileFromExternalId(externalId string) (*FileEntity, error)
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

func (repo *FileRepository) GetFileFromExternalId(externalId string) (*FileEntity, error) {
	query := `SELECT id, external_id, state, created_at, updated_at FROM files WHERE external_id = $1`
	row := repo.DB.QueryRow(query, externalId)

	file := &FileEntity{}
	err := row.Scan(&file.ID, &file.ExternalId, &file.State, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return file, nil
}
