package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UploadFile(data io.Reader, filename string) error {
	// Checking the existence of the file
	_, err := os.Stat(filename)
	if err == nil {
		return fmt.Errorf("file %s already exists", filename)
	}

	// If the file already exists, return an error
	if !os.IsNotExist(err) {
		return err
	}

	file, errCreate := os.Create(filename)
	if errCreate != nil {
		return err
	}

	// Use defer to close the file and handle the error
	defer func(file *os.File) {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}(file)

	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}

	return nil
}

// FileExistsAndIsAudio checks if a file exists and has an MP3 or FLAC extension.
func FileExistsAndIsAudio(filePath string) (bool, error) {
	// Check if the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist
			return false, nil
		}
		// Error occurred while checking
		return false, err
	}

	// Check if the file extension is MP3 or FLAC (case-insensitive)
	extension := strings.ToLower(filepath.Ext(filePath))
	return extension == ".mp3" || extension == ".flac", nil
}

func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	default:
		return "application/octet-stream"
	}
}
