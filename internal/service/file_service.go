package service

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/repository"
	"github.com/jwtly10/jambda/internal/utils"
	"github.com/spf13/afero"
)

type FileService struct {
	repo repository.IFileRepository
	log  logging.Logger
	fs   afero.Fs
}

func NewFileService(repo repository.IFileRepository, log logging.Logger, fs afero.Fs) *FileService {
	return &FileService{
		repo: repo,
		log:  log,
		fs:   fs,
	}
}

func (fs *FileService) ProcessFileUpload(r *http.Request) (*data.FileEntity, error) {
	genId := utils.GenerateShortID()
	fs.log.Infof("Processing file for jambda function  %s", genId)

	r.ParseMultipartForm(10 << 20) // Limit upload size

	file, _, err := r.FormFile("upload")
	if err != nil {
		return &data.FileEntity{}, fmt.Errorf("error retrieving the file from request: %v", err)
	}
	defer file.Close()

	// Validate it's a zip file
	if !fs.isValidZipFile(file) {
		return &data.FileEntity{}, fmt.Errorf("file is not a valid zip archive")
	}

	// Extract, validate and save uploaded binary
	err = fs.handleFile(genId, file)
	if err != nil {
		return &data.FileEntity{}, err
	}

	return fs.repo.InitFileMetaData(genId)
}

func (fs *FileService) IsValidFunctionId(functionId string) bool {
	fileEntity, err := fs.repo.GetFileFromExternalId(functionId)
	if err != nil {
		fs.log.Error("error getting file meta data from external id", err)
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

func (fs *FileService) extractAndValidateZip(zipPath, genId string) error {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fs.log.Infof("Found file %s", f.Name)
		// TODO: More file validation, support jars/python scripts
		if f.Name == "bootstrap" && f.FileInfo().Mode().IsRegular() {
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
		fs.log.Errorf("Failed to set '%s' as executable: %s", outputPath, err)
		return err
	}

	fs.log.Infof("Successfully extracted and set executable permissions for '%s'", outputPath)
	return err
}
