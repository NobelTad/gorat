package main

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procMessageBoxW   = user32.NewProc("MessageBoxW")
	procCreateWindow  = user32.NewProc("CreateWindowExW")
	procDestroyWindow = user32.NewProc("DestroyWindow")
	procShowWindow    = user32.NewProc("ShowWindow")
	procUpdateWindow  = user32.NewProc("UpdateWindow")
)

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

func main() {
	showErrorMsg()
}
