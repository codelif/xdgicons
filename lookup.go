package xdgicons

import (
	"fmt"
	"math"
	"path"
	"strings"
	"sync"
	"time"
)

type iconLookup struct {
	theme                   string
	fallbackTheme           string
	extensions              []string
	themeInfoCache          map[string]ThemeInfo
	dirCache                map[string]*baseDirIconCache
	cacheValidCheckInterval time.Duration
	mu                      sync.RWMutex
}

func NewIconLookup() IconLookup {
	return NewIconLookupWithConfig(LookupConfig{})
}

type LookupConfig struct {
	// Icon Theme to use.
	//
	// If unset, then uses default theme
	Theme string

	// Fallback Theme to look into,
	// if no icon was found using the canonical algorithm
	//
	// If unset, does not search for a fallback theme
	FallbackTheme string

	// Icon file extensions to search for.
	//
	// If unset, defaults to ["png", "svg", "xpm"]
	Extensions []string
}

func NewIconLookupWithConfig(cfg LookupConfig) IconLookup {
	il := &iconLookup{
		themeInfoCache:          make(map[string]ThemeInfo),
		dirCache:                make(map[string]*baseDirIconCache),
		cacheValidCheckInterval: 5 * time.Second,
	}

	if cfg.Theme == "" {
		il.theme = DefaultTheme()
	} else {
		il.theme = cfg.Theme
	}

	il.fallbackTheme = cfg.FallbackTheme

	if len(cfg.Extensions) == 0 {
		il.extensions = []string{"png", "svg", "xpm"}
	} else {
		il.extensions = cfg.Extensions
	}

	il.createInitialCache()
	return il
}

func (il *iconLookup) Lookup(iconName string) (Icon, error) {
	icon, err := il.FindIcon(iconName, 48, 1)
	return icon, err
}

func (il *iconLookup) FindIcon(iconName string, size int, scale int) (Icon, error) {
	icon, err := il.findIconHelper(iconName, size, scale, il.theme)
	if err == nil {
		return icon, nil
	}
	// hicolor is always searched in findIconHelper since readIndex adds it to Inherits

	icon, err = il.lookupFallbackIcon(iconName)
	if err == nil {
		return icon, nil
	}

	// searching a fallback theme as well... since some apps (blueman-applet)
	// asks for bluetooth-symbolic which is not in hicolor (so specifying adwaita
	// can be useful)
	if il.fallbackTheme != "" {
		icon, err = il.findIconHelper(iconName, size, scale, il.fallbackTheme)
		if err == nil {
			return icon, nil
		}
	}
	return Icon{}, fmt.Errorf("icon %q not found", iconName)
}

func (il *iconLookup) findIconHelper(iconName string, size int, scale int, theme string) (Icon, error) {
	// fmt.Printf("Searching icon=%q size=%d scale=%d theme=%q\n", iconName, size, scale, theme)
	themeInfo, err := il.getThemeInfo(theme)
	if err != nil {
		return Icon{}, err
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

	return Icon{}, fmt.Errorf("icon %q not found", iconName)
}

func (il *iconLookup) lookupIcon(iconName string, size int, scale int, theme string) (Icon, error) {
	themeInfo, err := il.getThemeInfo(theme)
	if err != nil {
		return Icon{}, err
	}

	for _, subdir := range append(themeInfo.Directories, themeInfo.ScaledDirectories...) {
		for _, directory := range GetBaseDirs() {
			for _, extension := range il.extensions {
				if il.directoryMatchesSize(themeInfo, subdir, size, scale) {
					iconPath := path.Join(directory, theme, subdir, iconName+"."+extension)
					// fmt.Printf("[XDGICONS]: Searching for %q\n", iconPath)
					if il.fileExists(directory, iconPath) {
						iconInfo := themeInfo.directoryMap[subdir]
						return Icon{
							Name:    iconName,
							Path:    iconPath,
							Size:    iconInfo.Size,
							MinSize: iconInfo.MinSize,
							MaxSize: iconInfo.MaxSize,
							Scale:   iconInfo.Scale,
						}, nil
					}
				}
			}
		}
	}

	minimalSize := math.MaxInt
	var closestFilename string
	var closestSubdir string

	for _, subdir := range append(themeInfo.Directories, themeInfo.ScaledDirectories...) {
		for _, directory := range GetBaseDirs() {
			for _, extension := range il.extensions {
				iconPath := path.Join(directory, theme, subdir, iconName+"."+extension)
				if il.fileExists(directory, iconPath) && il.directorySizeDistance(themeInfo, subdir, size, scale) < minimalSize {
					// fmt.Printf("[XDGICONS]: Searching for %q\n", iconPath)
					closestFilename = iconPath
					closestSubdir = subdir
					minimalSize = il.directorySizeDistance(themeInfo, subdir, size, scale)
				}
			}
		}
	}
	if closestFilename != "" {
		iconInfo := themeInfo.directoryMap[closestSubdir]
		return Icon{
			Name:    iconName,
			Path:    closestFilename,
			Size:    iconInfo.Size,
			MinSize: iconInfo.MinSize,
			MaxSize: iconInfo.MaxSize,
			Scale:   iconInfo.Scale,
		}, nil
	}

	return Icon{}, fmt.Errorf("icon %q not found", iconName)
}

func (il *iconLookup) lookupFallbackIcon(iconName string) (Icon, error) {

	for _, directory := range GetBaseDirs() {
		for _, extension := range il.extensions {
			iconPath := path.Join(directory, iconName+"."+extension)

			// fmt.Printf("[XDGICONS]: Searching for %q\n", iconPath)
			if il.fileExists(directory, iconPath) {
				return Icon{
					Name: iconName,
					Path: iconPath,
				}, nil
			}
		}
	}

	return Icon{}, fmt.Errorf("icon %q not found", iconName)
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

func (il *iconLookup) FindBestIcon(iconList []string, size int, scale int) (Icon, error) {

	icon, err := il.findBestIconHelper(iconList, size, scale, il.theme)
	if err == nil {
		return icon, nil
	}

	// hicolor is always searched in findIconHelper since readIndex adds it to Inherits

	// doing a fallback lookup in pixmaps directory
	for _, iconName := range iconList {
		icon, err := il.lookupFallbackIcon(iconName)
		if err == nil {
			return icon, nil
		}
	}

	// searching a fallback theme as well... since some apps (blueman-applet)
	// asks for bluetooth-symbolic which is not in hicolor (so specifying adwaita
	// can be useful)
	if il.fallbackTheme != "" {
		icon, err = il.findBestIconHelper(iconList, size, scale, il.fallbackTheme)
		if err == nil {
			return icon, nil
		}
	}

	return Icon{}, fmt.Errorf("icons \"%s\" not found", strings.Join(iconList, ","))
}

func (il *iconLookup) findBestIconHelper(iconList []string, size int, scale int, theme string) (Icon, error) {
	// fmt.Printf("Searching icon=%q size=%d scale=%d theme=%q\n", iconName, size, scale, theme)
	themeInfo, err := il.getThemeInfo(theme)
	if err != nil {
		return Icon{}, err
	}

	for _, iconName := range iconList {
		icon, err := il.lookupIcon(iconName, size, scale, theme)
		if err == nil {
			return icon, nil
		}
	}

	for _, parent := range themeInfo.Inherits {
		icon, err := il.findBestIconHelper(iconList, size, scale, parent)
		if err == nil {
			return icon, nil
		}
	}

	return Icon{}, fmt.Errorf("icons \"%s\" not found", strings.Join(iconList, ","))
}
