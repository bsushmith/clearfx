//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package terminal

import (
	"os"
	"strconv"
)

func Size(file *os.File) (int, int) {
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
