package repository

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jwtly10/jambda/api/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func NewMock() (*sql.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	return db, mock, nil
}

func TestGetAllActiveFunctions(t *testing.T) {
	db, mock, err := NewMock()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "external_id", "state", "configuration", "created_at", "updated_at"}).
		AddRow(1, "Test Function", "ext123", "ACTIVE", json.RawMessage(`{"trigger":"cron","image":"openjdk:21-jdk","type":"SINGLE","port":0}`), time.Now(), time.Now())
	mock.ExpectQuery(`^SELECT id, name, external_id, state, configuration, created_at, updated_at FROM functions_tb WHERE state = 'ACTIVE'$`).WillReturnRows(rows)

	repo := NewFunctionRepository(db)
	functions, err := repo.GetAllActiveFunctions()
	assert.NoError(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, "Test Function", functions[0].Name)
	assert.Equal(t, "ext123", functions[0].ExternalId)
	assert.Equal(t, "ACTIVE", functions[0].State)
	assert.Equal(t, "ACTIVE", functions[0].State)
	expectedConfig := &data.FunctionConfig{
		Type:    "SINGLE",
		Trigger: "cron",
		Image:   "openjdk:21-jdk",
		Port:    new(int),
		EnvVars: nil,
	}
	assert.Equal(t, expectedConfig, functions[0].Configuration)
}

func TestDeleteFunctionByExternalId(t *testing.T) {
	db, mock, err := NewMock()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE functions_tb SET state = 'DELETED', updated_at = NOW\(\) WHERE external_id = \$1 AND state <> 'DELETED'`).
		WithArgs("ext123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewFunctionRepository(db)
	err = repo.DeleteFunctionByExternalId("ext123")
	assert.NoError(t, err)
}

func TestSaveFunction(t *testing.T) {
	db, mock, err := NewMock()
	require.NoError(t, err)

	defer db.Close()

	now := time.Now().UTC()

	function := data.FunctionEntity{
		ExternalId: "ext123",
		Name:       "Test Function",
		State:      "ACTIVE",
		Configuration: &data.FunctionConfig{
			Type:    "SINGLE",
			Trigger: "cron",
			Image:   "openjdk:21-jdk",
			Port:    new(int),
			EnvVars: nil,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	configJSON, err := json.Marshal(function.Configuration)
	require.NoError(t, err)

	mock.ExpectQuery(`INSERT INTO functions_tb`).
		WithArgs(
			function.ExternalId,
			function.Name,
			function.State,
			configJSON,
			AnyTime{},
			AnyTime{},
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "external_id", "state", "configuration", "created_at", "updated_at"}).
			AddRow(1, function.Name, function.ExternalId, function.State, configJSON, time.Now(), time.Now()))

	repo := NewFunctionRepository(db)
	savedFunction, err := repo.SaveFunction(function.ExternalId, function.Name, *function.Configuration)
	require.NoError(t, err)
	require.NotNil(t, savedFunction)
	assert.Equal(t, "Test Function", savedFunction.Name)
	assert.Equal(t, "ext123", savedFunction.ExternalId)
	assert.Equal(t, "ACTIVE", savedFunction.State)
	assert.Equal(t, "ACTIVE", savedFunction.State)
	expectedConfig := &data.FunctionConfig{
		Type:    "SINGLE",
		Trigger: "cron",
		Image:   "openjdk:21-jdk",
		Port:    new(int),
		EnvVars: nil,
	}
	assert.Equal(t, expectedConfig, savedFunction.Configuration)
}

func TestGetConfigurationFromExternalId(t *testing.T) {
	db, mock, err := NewMock()
	require.NoError(t, err)
	defer db.Close()

	config := data.FunctionConfig{
		Trigger: "http",
		Type:    "SINGLE",
	}
	configJSON, _ := json.Marshal(config)
	rows := sqlmock.NewRows([]string{"configuration"}).AddRow(configJSON)

	mock.ExpectQuery(`SELECT configuration FROM functions_tb WHERE external_id = \$1`).WithArgs("123").
		WillReturnRows(rows)

	repo := NewFunctionRepository(db)
	result, err := repo.GetConfigurationFromExternalId("123")
	require.NoError(t, err)
	assert.Equal(t, "http", result.Trigger)
	assert.Equal(t, "SINGLE", result.Type)
}

func TestGetFunctionEntityFromExternalId(t *testing.T) {
	db, mock, err := NewMock()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "external_id", "state", "created_at", "updated_at"}).
		AddRow(1, "LambdaFunc", "123", "ACTIVE", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, name, external_id, state, created_at, updated_at FROM functions_tb WHERE external_id = \$1`).
		WithArgs("123").WillReturnRows(rows)

	repo := NewFunctionRepository(db)
	function, err := repo.GetFunctionEntityFromExternalId("123")
	require.NoError(t, err)
	assert.Equal(t, "LambdaFunc", function.Name)
	assert.Equal(t, "123", function.ExternalId)
	assert.Equal(t, "ACTIVE", function.State)
	assert.Nil(t, function.Configuration, "config should be nil (not returned by sql)")
}

func TestUpdateConfigByExternalId(t *testing.T) {
	db, mock, err := NewMock()
	require.NoError(t, err)
	defer db.Close()

	config := data.FunctionConfig{
		Trigger: "http",
		Type:    "SINGLE",
	}
	configJSON, _ := json.Marshal(config)
	rows := sqlmock.NewRows([]string{"id", "name", "external_id", "state", "configuration", "created_at", "updated_at"}).
		AddRow(1, "LambdaFuncUpdated", "123", "ACTIVE", configJSON, time.Now(), time.Now())

	mock.ExpectQuery(`UPDATE functions_tb SET configuration = \$2, updated_at = NOW\(\), name = \$3 WHERE external_id = \$1 RETURNING id, name, external_id, state, configuration, created_at, updated_at`).
		WithArgs("123", configJSON, "LambdaFuncUpdated").WillReturnRows(rows)

	repo := NewFunctionRepository(db)
	updatedFunction, err := repo.UpdateConfigByExternalId("123", "LambdaFuncUpdated", config)
	require.NoError(t, err)
	assert.Equal(t, "LambdaFuncUpdated", updatedFunction.Name)
	assert.Equal(t, "http", updatedFunction.Configuration.Trigger)
}
