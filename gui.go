package main

import (
	"syscall"
	"unsafe"
)

func showErrorMsg() {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")

	title := "Error"
	message := "Error Your system does't meet the minimum requirement of directx12"

	// HWND = 0, Text, Title, MB_ICONERROR = 0x10
	msgBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0x10, // MB_ICONERROR
	)
}

func main() {
	showErrorMsg()
}
