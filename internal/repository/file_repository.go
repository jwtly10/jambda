package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jwtly10/jambda/api/data"
)

type IFunctionRepository interface {
	InitFunctionEntity(externalId string) (*data.FunctionEntity, error)
	GetFunctionEntityFromExternalId(externalId string) (*data.FunctionEntity, error)
	GetConfigurationFromExternalId(externalId string) (*data.FunctionConfig, error)
}

type FunctionRepository struct {
	Db *sql.DB
}

func NewFunctionRepository(db *sql.DB) *FunctionRepository {
	return &FunctionRepository{Db: db}
}

func (repo *FunctionRepository) GetConfigurationFromExternalId(externalId string) (*data.FunctionConfig, error) {
	query := `SELECT configuration FROM functions_tb WHERE external_id = $1`

	var configData []byte
	row := repo.Db.QueryRow(query, externalId)
	err := row.Scan(&configData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no file found with external_id %s", externalId)
		}
		return nil, err
	}

	var config data.FunctionConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling configuration: %v", err)
	}

	return &config, nil
}

func (repo *FunctionRepository) InitFunctionEntity(externalId string) (*data.FunctionEntity, error) {
	query := `INSERT INTO functions_tb (external_id, state) VALUES ($1, $2) RETURNING id, external_id, state, created_at, updated_at`
	row := repo.Db.QueryRow(query, externalId, "ACTIVE")

	var fileEntity data.FunctionEntity
	err := row.Scan(&fileEntity.ID, &fileEntity.ExternalId, &fileEntity.State, &fileEntity.CreatedAt, &fileEntity.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &fileEntity, nil
}

func (repo *FunctionRepository) GetFunctionEntityFromExternalId(externalId string) (*data.FunctionEntity, error) {
	query := `SELECT id, external_id, state, created_at, updated_at FROM functions_tb WHERE external_id = $1`
	row := repo.Db.QueryRow(query, externalId)

	file := &data.FunctionEntity{}
	err := row.Scan(&file.ID, &file.ExternalId, &file.State, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return file, nil
}
