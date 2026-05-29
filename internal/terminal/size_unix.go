//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package terminal

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func Size(file *os.File) (int, int) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno == 0 && ws.Col > 0 && ws.Row > 0 {
		return int(ws.Col), int(ws.Row)
	}
	return envSize()
}

func envSize() (int, int) {
	width := envInt("COLUMNS", 80)
	height := envInt("LINES", 24)
	return width, height
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
