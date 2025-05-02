package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFile(fileURL, destination string) string {
	// If no destination is given, default to "data" folder
	if destination == "" {
		destination = "data"
	}

	// Make sure directory exists
	dir := filepath.Dir(destination)
	if destination == "data" {
		dir = "data"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// Download the file
	resp, err := http.Get(fileURL)
	if err != nil {
		return "Error downloading file: " + err.Error()
	}
	defer resp.Body.Close()

	// Try to get filename from Content-Disposition
	contentDisp := resp.Header.Get("Content-Disposition")
	var fileName string
	if strings.Contains(contentDisp, "filename=") {
		parts := strings.Split(contentDisp, "filename=")
		fileName = strings.Trim(parts[1], "\" ")
	} else {
		fileName = filepath.Base(fileURL)
		if !strings.Contains(fileName, ".") {
			fileName += ".bin"
		}
	}

	// Final file path
	var filePath string
	if destination == "data" || strings.HasSuffix(destination, string(os.PathSeparator)) {
		filePath = filepath.Join(dir, fileName)
	} else if strings.HasSuffix(destination, fileName) {
		filePath = destination
	} else {
		filePath = filepath.Join(destination, fileName)
	}

	// Create file
	out, err := os.Create(filePath)
	if err != nil {
		return "Error creating file: " + err.Error()
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "Error saving file: " + err.Error()
	}

	return "File saved as: " + filePath
}

func main() {
	// Example usage
	url := "https://github.com/NobelTad/test/raw/refs/heads/main/main.exe" // use /raw/ to get the actual file
	destination := ""
	result := DownloadFile(url, destination)
	fmt.Println(result)
	url2 := "https://raw.githubusercontent.com/NobelTad/test/refs/heads/main/main.go" // use /raw/ to get the actual file
	destination2 := ""
	result2 := DownloadFile(url2, destination2)
	fmt.Println(result2)
}
