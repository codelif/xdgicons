package xdgicons

import (
	"os"
	"path"
	"strings"
)

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}

func GetBaseDirs() (baseDirs []string) {
	homeDir := os.Getenv("HOME")
	dataDirs := strings.Split(os.Getenv("XDG_DATA_DIRS"), ":")
	pixmapDir := "/usr/share/pixmaps"

	if homeDir != "" {
		baseDirs = append(baseDirs, path.Join(homeDir, ".icons"))
	}

	for _, dataDir := range dataDirs {
		baseDirs = append(baseDirs, path.Join(dataDir, "icons"))
	}

	baseDirs = append(baseDirs, pixmapDir)
	return baseDirs
}
