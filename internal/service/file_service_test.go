package service

import (
	"archive/zip"
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/repository"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCanProcessFileAndCallDBWithNewRecord(t *testing.T) {
	logger := logging.NewLogger(false, slog.LevelDebug.Level())
	defer os.RemoveAll("binaries") // TODO: INJECT AFERO FS INSTEAD OF THIS

	fs := afero.NewOsFs()

	// Mocks
	repo := new(repository.MockFileRepository)
	repo.On("InitFileMetaData", mock.Anything).Return(nil)

	fileService := NewFileService(repo, logger, fs)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("upload", "test.zip")
	assert.NoError(t, err)

	err = createTestZipFile(part)
	assert.NoError(t, err)

	writer.Close()

	req, err := http.NewRequest("POST", "/", &b)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Call the method with mocked request
	err = fileService.ProcessFileUpload(req)
	assert.NoError(t, err)

	repo.AssertCalled(t, "InitFileMetaData", mock.Anything)
}

func TestCanHandleFile(t *testing.T) {
	logger := logging.NewLogger(false, slog.LevelDebug.Level())
	defer os.RemoveAll("binaries") // TODO: INJECT AFERO FS INSTEAD OF THIS

	fs := afero.NewOsFs()

	// Mocks
	fileService := NewFileService(nil, logger, fs)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("upload", "test.zip")
	assert.NoError(t, err)

	err = createTestZipFile(part)
	assert.NoError(t, err)
	writer.Close()

	file := createMockMPFile(b.Bytes())

	// Call the method with mocked file
	err = fileService.handleFile("test-id", file)
	assert.NoError(t, err)
}

func TestCanExtractAndValidateAZip(t *testing.T) {
	logger := logging.NewLogger(false, slog.LevelDebug.Level())
	defer os.RemoveAll("binaries") // TODO: INJECT AFERO FS INSTEAD OF THIS

	// Mocks
	fs := afero.NewOsFs()
	fileService := NewFileService(nil, logger, fs)

	tmpFile, err := os.CreateTemp("", "*.zip")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = createTestZipFile(tmpFile)
	assert.NoError(t, err)
	tmpFile.Close()

	// Call the method with the mocked file
	err = fileService.extractAndValidateZip(tmpFile.Name(), "test-id")
	assert.NoError(t, err)

	extractPath := filepath.Join("binaries", "test-id", "bootstrap")
	_, err = os.Stat(extractPath)
	assert.NoError(t, err)
	os.RemoveAll(filepath.Dir(extractPath))
}

// TEST HELPER UTILS

// createTestZipFile creates a test zip file
func createTestZipFile(w io.Writer) error {
	zipWriter := zip.NewWriter(w)

	// Add a file named "bootstrap"
	f, err := zipWriter.Create("bootstrap")
	if err != nil {
		return err
	}
	_, err = f.Write([]byte("this is a test file"))
	if err != nil {
		return err
	}

	return zipWriter.Close()
}

func createMockMPFile(data []byte) multipart.File {
	return &fakeFile{data: data, reader: bytes.NewReader(data)}
}

type fakeFile struct {
	data   []byte
	reader *bytes.Reader
}

func (f *fakeFile) Read(p []byte) (n int, err error) {
	return f.reader.Read(p)
}

func (f *fakeFile) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

func (f *fakeFile) Close() error {
	return nil
}

func (f *fakeFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.reader.ReadAt(p, off)
}
