package xdgicons

import (
	"fmt"
	"math"
	"path"
	"sync"
	"time"
)

type iconLookup struct {
	theme                   string
	extensions              []string
	themeInfoCache          map[string]ThemeInfo
	dirCache                map[string]*baseDirIconCache
	cacheValidCheckInterval time.Duration
	mu                      sync.RWMutex
}

func NewIconLookup() IconLookup {
	il := &iconLookup{
		theme:                   DefaultTheme(),
		themeInfoCache:          make(map[string]ThemeInfo),
		dirCache:                make(map[string]*baseDirIconCache),
		extensions:              []string{"png", "svg", "xpm"},
		cacheValidCheckInterval: 5 * time.Second,
	}

	il.createInitialCache()

	return il
}

// testing old API
func (il *iconLookup) Lookup(iconName string) (string, error) {
	icon, err := il.FindIcon(iconName, 48, 1)
	return icon.Path, err
}

func (il *iconLookup) FindIcon(iconName string, size int, scale int) (*Icon, error) {
	icon, err := il.findIconHelper(iconName, size, scale, il.theme)
	if err == nil {
		return icon, nil
	}

	icon, err = il.findIconHelper(iconName, size, scale, "hicolor")
	if err == nil {
		return icon, nil
	}

	return il.lookupFallbackIcon(iconName)
}

func (il *iconLookup) findIconHelper(iconName string, size int, scale int, theme string) (*Icon, error) {
	fmt.Printf("Searching icon=%q size=%d scale=%d theme=%q\n", iconName, size, scale, theme)
	themeInfo, err := il.getThemeInfo(theme)
	if err != nil {
		return nil, err
	}

	icon, err := il.lookupIcon(iconName, size, scale, theme)
	if err == nil {
		return icon, nil
	}

	for _, parent := range themeInfo.Inherits {
		icon, err := il.findIconHelper(iconName, size, scale, parent)
		if err == nil {
			return icon, nil
		}
	}

	return nil, fmt.Errorf("icon %q not found", iconName)
}

func (il *iconLookup) lookupIcon(iconName string, size int, scale int, theme string) (*Icon, error) {
	themeInfo, err := il.getThemeInfo(theme)
	if err != nil {
		return nil, err
	}

	for _, subdir := range append(themeInfo.Directories, themeInfo.ScaledDirectories...) {
		for _, directory := range GetBaseDirs() {
			for _, extension := range il.extensions {
				if il.directoryMatchesSize(themeInfo, subdir, size, scale) {
					iconPath := path.Join(directory, theme, subdir, iconName+"."+extension)
					// fmt.Printf("[XDGICONS]: Searching for %q\n", iconPath)
					if il.fileExists(directory, iconPath) {
						return &Icon{iconPath}, nil
					}
				}
			}
		}
	}

	minimalSize := math.MaxInt
	var closestFilename string

	for _, subdir := range append(themeInfo.Directories, themeInfo.ScaledDirectories...) {
		for _, directory := range GetBaseDirs() {
			for _, extension := range il.extensions {
				iconPath := path.Join(directory, theme, subdir, iconName+"."+extension)
				if il.fileExists(directory, iconPath) && il.directorySizeDistance(themeInfo, subdir, size, scale) < minimalSize {
					// fmt.Printf("[XDGICONS]: Searching for %q\n", iconPath)
					closestFilename = iconPath
					minimalSize = il.directorySizeDistance(themeInfo, subdir, size, scale)
				}
			}
		}
	}
	if closestFilename != "" {
		return &Icon{closestFilename}, nil
	}

	return nil, fmt.Errorf("icon %q not found", iconName)
}

func (il *iconLookup) lookupFallbackIcon(iconName string) (*Icon, error) {

	for _, directory := range GetBaseDirs() {
		for _, extension := range il.extensions {
			iconPath := path.Join(directory, iconName+"."+extension)

			// fmt.Printf("[XDGICONS]: Searching for %q\n", iconPath)
			if il.fileExists(directory, iconPath) {
				return &Icon{iconPath}, nil
			}
		}
	}

	return nil, fmt.Errorf("icon %q not found", iconName)
}

func (il *iconLookup) fileExists(baseDir, iconPath string) bool {
	il.mu.RLock()
	cacheEntry, exists := il.dirCache[baseDir]
	il.mu.RUnlock()

	now := time.Now()

	if !exists || now.Sub(cacheEntry.lastStat) >= il.cacheValidCheckInterval {
		if il.shouldRefreshCache(baseDir, cacheEntry, now) {
			il.mu.Lock()
			il.cacheBaseDirectory(baseDir)
			cacheEntry = il.dirCache[baseDir]
			il.mu.Unlock()
		} else if exists {
			il.mu.Lock()
			cacheEntry.lastStat = now
			il.mu.Unlock()
		}
	}

	if cacheEntry == nil {
		return false
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	return cacheEntry.files[iconPath]
}
func (il *iconLookup) directoryMatchesSize(themeInfo ThemeInfo, subdir string, iconSize int, iconScale int) bool {
	subdirInfo := themeInfo.directoryMap[subdir]

	if subdirInfo.Scale != iconScale {
		return false
	}

	switch subdirInfo.Type {
	case "Fixed":
		return subdirInfo.Size == iconSize
	case "Scalable":
		return subdirInfo.MinSize <= iconSize && iconSize <= subdirInfo.MaxSize
	case "Threshold":
		return subdirInfo.Size-subdirInfo.Threshold <= iconSize && iconSize <= subdirInfo.Size+subdirInfo.Threshold
	}

	return false // this should be unreachable
}

func (il *iconLookup) directorySizeDistance(themeInfo ThemeInfo, subdir string, iconSize int, iconScale int) int {
	subdirInfo := themeInfo.directoryMap[subdir]

	switch subdirInfo.Type {
	case "Fixed":
		return abs(subdirInfo.Size*subdirInfo.Scale - iconSize*iconScale)
	case "Scalable":
		if iconSize*iconScale < subdirInfo.MinSize*subdirInfo.Scale {
			return subdirInfo.MinSize*subdirInfo.Scale - iconSize*iconScale
		}
		if iconSize*iconScale > subdirInfo.MaxSize*subdirInfo.Scale {
			return iconSize*iconScale - subdirInfo.MaxSize*subdirInfo.Scale
		}
		return 0
	case "Threshold":
		if iconSize*iconScale < (subdirInfo.Size-subdirInfo.Threshold)*subdirInfo.Scale {
			return subdirInfo.MinSize*subdirInfo.Scale - iconSize*iconScale
		}
		if iconSize*iconScale > (subdirInfo.Size+subdirInfo.Threshold)*subdirInfo.Scale {
			return iconSize*iconScale - subdirInfo.MaxSize*subdirInfo.Scale
		}
		return 0
	}

	return 0 // this should be unreachable
}

func (il *iconLookup) FindBestIcon(iconList []string, size int, scale int) (*Icon, error) {
	return nil, nil
}
