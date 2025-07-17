package xdgicons

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"slices"
	"strings"

	"gopkg.in/ini.v1"
)

func DefaultTheme() (theme string) {
	dconfPath := []string{
		"org",
		"gnome",
		"desktop",
		"interface",
		"icon-theme",
	}

	cmd := exec.Command("dconf", "read", path.Join(dconfPath...))
	outputBytes := new(bytes.Buffer)
	cmd.Stdout = outputBytes
	cmd.Run()
	theme = cleanDconfOutput(outputBytes.String())
	if theme != "" {
		return theme
	}

	basenameIndex := len(dconfPath) - 1
	gsettingsSchema := strings.Join(dconfPath[:basenameIndex], ".")
	gsettingsKey := dconfPath[basenameIndex]

	cmd = exec.Command("gsettings", "get", gsettingsSchema, gsettingsKey)
	outputBytes.Reset()
	cmd.Stdout = outputBytes
	cmd.Run()
	theme = cleanDconfOutput(outputBytes.String())
	if theme != "" {
		return theme
	}

	return "hicolor"
}

func cleanDconfOutput(raw string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.Trim(raw, "\n "), "'"), "'")
}

// returns current theme
func (il *IconLookup) Theme() string {
	return il.theme
}

// returns fallback theme
func (il *IconLookup) FallbackTheme() string {
	return il.fallbackTheme
}

func (il *IconLookup) readThemeIndex(theme, indexPath string) (*ThemeInfo, error) {
	index, err := ini.Load(indexPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	iconThemeSection, err := index.GetSection("Icon Theme")
	if err != nil {
		return nil, fmt.Errorf("error reading required section: %v", err)
	}

	nameKey, err := iconThemeSection.GetKey("Name")
	if err != nil {
		return nil, fmt.Errorf("error reading required key: %v", err)
	}

	directorys, err := iconThemeSection.GetKey("Directories")
	if err != nil {
		return nil, fmt.Errorf("error reading required key: %v", err)
	}

	themeInfo := &ThemeInfo{
		Name:         nameKey.String(),
		Directories:  directorys.Strings(","),
		directoryMap: make(map[string]SubDirIconInfo),
	}

	inheritsKey, err := iconThemeSection.GetKey("Inherits")
	if err == nil {
		themeInfo.Inherits = inheritsKey.Strings(",")
	}

	if theme != "hicolor" && !slices.Contains(themeInfo.Inherits, "hicolor") {
		themeInfo.Inherits = slices.Insert(themeInfo.Inherits, len(themeInfo.Inherits), "hicolor")
	}

	scaledDirectorys, err := iconThemeSection.GetKey("ScaledDirectories")
	if err == nil {
		themeInfo.ScaledDirectories = scaledDirectorys.Strings(",")
	}

	for _, dir := range append(themeInfo.Directories, themeInfo.ScaledDirectories...) {
		dirSection, err := index.GetSection(dir)
		if err != nil {
			return nil, fmt.Errorf("error reading required section: %v", err)
		}

		sizeKey, err := dirSection.GetKey("Size")
		if err != nil {
			return nil, fmt.Errorf("error reading required key: %v", err)
		}

		size, err := sizeKey.Int()
		if err != nil {
			return nil, fmt.Errorf("error parsing value: %v", err)
		}

		subDirIconInfo := SubDirIconInfo{
			Size: size,
		}

		scaleKey, err := dirSection.GetKey("Scale")
		if err != nil {
			subDirIconInfo.Scale = 1
		} else {
			subDirIconInfo.Scale = scaleKey.MustInt(1)
		}

		typeKey, err := dirSection.GetKey("Type")
		if err != nil {
			subDirIconInfo.Type = "Threshold"
		} else {
			subDirIconInfo.Type = typeKey.MustString("Threshold")
		}

		maxSizeKey, err := dirSection.GetKey("MaxSize")
		if err != nil {
			subDirIconInfo.MaxSize = subDirIconInfo.Size
		} else {
			subDirIconInfo.MaxSize = maxSizeKey.MustInt(subDirIconInfo.Size)
		}

		minSizeKey, err := dirSection.GetKey("MinSize")
		if err != nil {
			subDirIconInfo.MinSize = subDirIconInfo.Size
		} else {
			subDirIconInfo.MinSize = minSizeKey.MustInt(subDirIconInfo.Size)
		}

		thresholdKey, err := dirSection.GetKey("Threshold")
		if err != nil {
			subDirIconInfo.Threshold = 2
		} else {
			subDirIconInfo.Threshold = thresholdKey.MustInt(2)
		}

		themeInfo.directoryMap[dir] = subDirIconInfo
	}

	return themeInfo, nil
}
