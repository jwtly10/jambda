package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/errors"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/repository"
	"github.com/jwtly10/jambda/internal/utils"
	"github.com/spf13/afero"
)

type FileService struct {
	repo repository.IFunctionRepository
	log  logging.Logger
	fs   afero.Fs
	cv   ConfigValidator
}

func NewFileService(repo repository.IFunctionRepository, log logging.Logger, fs afero.Fs, cv ConfigValidator) *FileService {
	return &FileService{
		repo: repo,
		log:  log,
		fs:   fs,
		cv:   cv,
	}
}

func (fs *FileService) ProcessNewFunction(r *http.Request) (*data.FunctionEntity, error) {
	genId := utils.GenerateShortID()

	// Get config from request
	configData := r.FormValue("config")
	var config *data.FunctionConfig
	err := json.Unmarshal([]byte(configData), &config)
	if err != nil {
		fs.log.Error("Error unmarshaling config json: ", err)
		return nil, errors.NewValidationError(fmt.Sprintf("error unmarshaling config json: %v", err))
	}

	fs.log.Infof("Validating uploaded config for '%s'", genId)
	err = fs.cv.ValidateConfig(config)
	if err != nil {
		fs.log.Error("Config validation failed with error: ", err)
		return nil, errors.NewValidationError(fmt.Sprintf("error validating config json: %v", err))
	}

	fs.log.Infof("Processing file for jambda function '%s'", genId)
	r.ParseMultipartForm(10 << 20) // Limit upload size 10MB

	file, _, err := r.FormFile("upload")
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("error retrieving the file from request: %v", err))
	}
	defer file.Close()

	// Validate it's a zip file
	if !fs.isValidZipFile(file) {
		return nil, errors.NewValidationError("uploaded file is not a valid zip archive")
	}

	err = fs.handleFile(genId, file)
	if err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("error unpacking and extracting binary: %v", err))
	}

	fileEntity, err := fs.repo.SaveFunction(genId, *config)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("error saving function to db: %v", err))
	}

	return fileEntity, nil
}

func (fs *FileService) IsValidExternalId(functionId string) bool {
	fileEntity, err := fs.repo.GetFunctionEntityFromExternalId(functionId)
	if err != nil {
		fs.log.Error("Error getting file meta data from external id: ", err)
		return false
	}

	if fileEntity == nil {
		return false
	}

	return true
}

func (fs *FileService) isValidZipFile(file multipart.File) bool {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return false
	}
	file.Seek(0, 0) // Rewind file after reading

	contentType := http.DetectContentType(buffer)
	fs.log.Debugf("Detected file type %s", contentType)

	return contentType == "application/zip"
}

func (fs *FileService) handleFile(genId string, file multipart.File) error {
	tmpFile, err := os.CreateTemp("", "*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	fs.log.Debug("Created temp zip file locally")

	if _, err = io.Copy(tmpFile, file); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %v", err)
	}

	return fs.extractAndValidateZip(tmpFile.Name(), genId)
}

// Validating the executable name
func (fs *FileService) extractAndValidateZip(zipPath, genId string) error {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fs.log.Infof("Found file %s", f.Name)
		// TODO: More file validation, support jars/python scripts
		if (f.Name == "bootstrap" || f.Name == "bootstrap.jar") && f.FileInfo().Mode().IsRegular() {
			extractPath := filepath.Join("binaries", genId, f.Name)
			if err := fs.extractFile(f, extractPath); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("bootstrap executable not found in zip")
}

func (fs *FileService) extractFile(f *zip.File, outputPath string) error {
	fs.log.Debugf("Extracting file '%s' to '%s'", f.Name, outputPath)
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	os.MkdirAll(filepath.Dir(outputPath), 0755)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)

	// Set the output file to be executable
	if err := os.Chmod(outputPath, 0755); err != nil {
		fs.log.Errorf("Failed to set '%s' as executable: %v", outputPath, err)
		return err
	}

	fs.log.Infof("Successfully extracted and set executable permissions for '%s'", outputPath)
	return err
}
