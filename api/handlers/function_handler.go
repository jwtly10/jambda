package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
	"github.com/jwtly10/jambda/internal/utils"
)

type FunctionHandler struct {
	log     logging.Logger
	service service.FunctionService
}

func NewFunctionHandler(l logging.Logger, fs service.FunctionService) *FunctionHandler {
	return &FunctionHandler{
		log:     l,
		service: fs,
	}
}

// @Summary Upload and process a file
// @Description Uploads a zip file, validates its contents, and processes it in storage. The zip file must contain a "bootstrap" executable. Returns ExternalId
// @Tags Functions
// @Accept multipart/form-data
// @Produce application/json
// @Param zip formData file true "File to upload"
// @Param config formData string true "JSON configuration data"
// @Param name formData string true "Display name of the function"
// @Success 201 {object} data.FunctionEntity "File uploaded and processed successfully"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /function [post]
func (nfh *FunctionHandler) UploadFunction(w http.ResponseWriter, r *http.Request) {
	res, err := nfh.service.UploadFunction(r)
	if err != nil {
		nfh.log.Error("uploading file failed with error: ", err)
		utils.HandleCustomErrors(w, err)
		return
	}

	jsonResponse, err := json.Marshal(res)
	if err != nil {
		nfh.log.Error("marshaling response failed with error: ", err)
		utils.HandleCustomErrors(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

// @Summary Update an existing function config
// @Description Updates an existing function config by submitting new config data
// @Tags Functions
// @Accept multipart/form-data
// @Produce application/json
// @Param id path string true "Function ID"
// @Param config formData string true "JSON configuration data"
// @Success 200 {object} data.FunctionEntity "Function updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /function/{id} [put]
func (nfh *FunctionHandler) UpdateFunction(w http.ResponseWriter, r *http.Request) {

	configData := r.FormValue("config")
	var config *data.FunctionConfig
	err := json.Unmarshal([]byte(configData), &config)
	if err != nil {
		nfh.log.Error("Error unmarshaling config json", err)
		utils.HandleBadRequest(w, fmt.Errorf("error unmarshling config json: %v", err))
		return
	}

	name := r.FormValue("name")
	if name == "" {
		nfh.log.Error("Name missing or empty from update")
		utils.HandleValidationError(w, fmt.Errorf("missing name from form data"))
		return
	}

	externalId, err := getIdFromUrl(r.URL)
	if err != nil {
		utils.HandleBadRequest(w, fmt.Errorf("error parsing externalId from URL"))
		return
	}

	res, err := nfh.service.UpdateConfig(externalId, name, config)
	if err != nil {
		nfh.log.Error("error updating config for id '%s': ", externalId, err)
		utils.HandleCustomErrors(w, err)
		return
	}

	jsonResponse, err := json.Marshal(res)
	if err != nil {
		nfh.log.Error("marshaling response failed with error: ", err)
		utils.HandleInternalError(w, fmt.Errorf("error marshalling response json: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

// @Summary List all functions
// @Description Retrieves a list of all function entities stored in the system.
// @Tags Functions
// @Produce application/json
// @Success 200 {array} data.FunctionEntity "List of all functions"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /function [get]
func (nfh *FunctionHandler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	functions, err := nfh.service.GetAllActiveFunctions()
	if err != nil {
		utils.HandleCustomErrors(w, err)
		return
	}

	jsonResponse, err := json.Marshal(functions)
	if err != nil {
		nfh.log.Error("Error marshaling functions to JSON: ", err)
		utils.HandleInternalError(w, fmt.Errorf("error marshalling response json: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// @Summary Delete a function
// @Description Deletes a specific function entity identified by its ID.
// @Tags Functions
// @Produce application/json
// @Param id path string true "Function ID"
// @Success 204 {string} string "Function deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /function/{id} [delete]
func (nfh *FunctionHandler) DeleteFunction(w http.ResponseWriter, r *http.Request) {

	externalId, err := getIdFromUrl(r.URL)
	if err != nil {
		utils.HandleBadRequest(w, fmt.Errorf("error parsing externalId from URL"))
		return
	}

	err = nfh.service.DeleteFunction(externalId)
	if err != nil {
		nfh.log.Error("Failed to delete function: ", err)
		utils.HandleCustomErrors(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Function deleted successfully"))
	return
}

func getIdFromUrl(url *url.URL) (string, error) {
	pathParts := strings.Split(url.Path, "/")
	// Assuming the URL pattern is v1/api/function/{id} and split should return 5 parts
	if len(pathParts) != 5 {
		// http.Error(w, "Invalid request path", http.StatusBadRequest)
		// return
		return "", fmt.Errorf("Invalid request path")
	}
	externalId := pathParts[4] // The forth part should be the ID

	if externalId == "" {
		return "", fmt.Errorf("Invalid function ID")
	}

	return externalId, nil
}
