package main

import (
	"syscall"
	"unsafe"
)

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

func main() {
	showInfoMsg()
}
