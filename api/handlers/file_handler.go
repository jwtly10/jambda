package handlers

import (
	"net/http"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
)

type FileHandler struct {
	log     logging.Logger
	service service.FileService
}

func NewFileHandler(l logging.Logger, fs service.FileService) *FileHandler {
	return &FileHandler{
		log:     l,
		service: fs,
	}
}

// @Summary Upload and process a file
// @Description Uploads a zip file, validates its contents, and processes it in storage. The zip file must contain a "bootstrap" executable.
// @Tags files
// @Accept multipart/form-data
// @Produce text/plain
// @Param upload formData file true "File to upload"
// @Success 200 {string} string "File uploaded and processed successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /file/upload [post]
func (nfh *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	err := nfh.service.ProcessFileUpload(r)
	if err != nil {
		nfh.log.Error("uploading file failed with error: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest) // Customize the status code based on error type
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded and processed successfully"))
}
