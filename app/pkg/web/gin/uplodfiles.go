package gin

import (
	"bytes"
	"io"
	"net/http"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/files"

	"github.com/gin-gonic/gin"
)

// PostFiles godoc
// @Summary Stream audio files.
// @Description Streams audio files in the specified directory as MP3 or FLAC.
// @Tags track-controller
// @Accept */*
// @Produce */*
// @Param control path string false "Upload media files"
// @Success 201 {array} model.Track "OK"
// @Failure 400 {object} model.ErrorResponse "empty required fields"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /audio/upload [post]
func (a *WebApp) PostFiles(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "empty required fields `file`"})
		return
	}

	// Create a buffer to store the file data
	var buffer bytes.Buffer

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: "error opening file"})
		return
	}
	defer file.Close()

	// Copy the data from the uploaded file to the buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: "error reading file"})
		return
	}

	fileType := http.DetectContentType(buffer.Bytes())

	musicTypes := getMusicTypes()
	if _, ex := musicTypes[fileType]; !ex {
		c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "file type is not supported"})
		return
	}

	pathFile := a.cfg.AppConfig.MusicPath + "/" + fileHeader.Filename

	// Pass the buffer (as io.Reader) to the UploadFile function
	err = files.UploadFile(&buffer, pathFile)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, model.OkResponse{Message: "file upload"})
}
