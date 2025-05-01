package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFile(fileURL, destination string) {
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
		fmt.Println("Error downloading file:", err)
		return
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
		// Just a folder, append fileName
		filePath = filepath.Join(dir, fileName)
	} else if strings.HasSuffix(destination, fileName) {
		// Full path with filename given
		filePath = destination
	} else {
		// Assume destination is directory
		filePath = filepath.Join(destination, fileName)
	}

	// Create file
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Println("File saved as:", filePath)
}

func main() {
	// Example usage
	url := "https://github.com/NobelTad/test/blob/main/main.exe"
	destination := "C:\\Users\\nobel\\Desktop" // try "" or "C:/Users/Nobel/Desktop/" or full path like "C:/Users/Nobel/Desktop/file.jpg"
	DownloadFile(url, destination)
}
