package repository

import (
	"database/sql"

	"github.com/jwtly10/jambda/api/data"
)

type IFileRepository interface {
	InitFileMetaData(externalId string) (*data.FileEntity, error)
	GetFileFromExternalId(externalId string) (*data.FileEntity, error)
}

type FileRepository struct {
	DB *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{DB: db}
}

func (repo *FileRepository) InitFileMetaData(externalId string) (*data.FileEntity, error) {
	query := `INSERT INTO files (external_id, state) VALUES ($1, $2) RETURNING id, external_id, state, created_at, updated_at`
	row := repo.DB.QueryRow(query, externalId, "ACTIVE")

	var fileEntity data.FileEntity
	err := row.Scan(&fileEntity.ID, &fileEntity.ExternalId, &fileEntity.State, &fileEntity.CreatedAt, &fileEntity.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &fileEntity, nil
}

func (repo *FileRepository) GetFileFromExternalId(externalId string) (*data.FileEntity, error) {
	query := `SELECT id, external_id, state, created_at, updated_at FROM files WHERE external_id = $1`
	row := repo.DB.QueryRow(query, externalId)

	file := &data.FileEntity{}
	err := row.Scan(&file.ID, &file.ExternalId, &file.State, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return file, nil
}
