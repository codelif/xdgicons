package xdgicons

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

type baseDirIconCache struct {
	files    map[string]bool
	mtime    time.Time
	lastStat time.Time
}

func (il *IconLookup) createInitialCache() {
	il.mu.Lock()
	defer il.mu.Unlock()

	for _, directory := range GetBaseDirs() {
		_ = il.cacheBaseDirectory(directory)
	}
}

func (il *IconLookup) cacheBaseDirectory(dirPath string) error {
	stat, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	files := make(map[string]bool)

	err = filepath.WalkDir(dirPath, func(subPath string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			files[subPath] = true
		}
		return nil
	})
	if err != nil {
		return err
	}

	// fmt.Println("caching Base Dir")
	il.clearThemeInfoCache()
	il.dirCache[dirPath] = &baseDirIconCache{
		files:    files,
		mtime:    stat.ModTime(),
		lastStat: time.Now(),
	}

	return nil
}

func (il *IconLookup) getThemeInfo(theme string) (ThemeInfo, error) {
	il.mu.RLock()
	themeInfo, ok := il.themeInfoCache[theme]
	il.mu.RUnlock()

	if ok {
		return themeInfo, nil
	}

	for _, directory := range GetBaseDirs() {
		indexPath := path.Join(directory, theme, "index.theme")
		_, err := os.Stat(indexPath)
		if err != nil {
			continue
		}
		themeInfo, err := il.readThemeIndex(theme, indexPath)
		if err != nil {
			continue
		}

		il.themeInfoCache[theme] = *themeInfo
		return *themeInfo, nil
	}

	return ThemeInfo{}, fmt.Errorf("theme %q not found", theme)
}

func (il *IconLookup) clearThemeInfoCache() {
	il.themeInfoCache = make(map[string]ThemeInfo)
}

func (il *IconLookup) shouldRefreshCache(baseDir string, cacheEntry *baseDirIconCache, now time.Time) bool {
	if cacheEntry == nil {
		return true
	}

	if now.Sub(cacheEntry.lastStat) < il.cacheValidCheckInterval {
		return false
	}

	stat, err := os.Stat(baseDir)
	if err != nil {
		il.mu.Lock()
		delete(il.dirCache, baseDir)
		il.mu.Unlock()
		return false
	}

	return !stat.ModTime().Equal(cacheEntry.mtime)
}
