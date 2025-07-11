package xdgicons

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// IconTheme represents an icon theme with its metadata
type IconTheme struct {
	Name        string
	Directories []string
	Parents     []string
	baseDirs    []string
}

// SubdirInfo represents metadata for a theme subdirectory
type SubdirInfo struct {
	Size      int
	Scale     int
	Type      string // "Fixed", "Scaled", "Threshold"
	MinSize   int
	MaxSize   int
	Threshold int
}

// IconLookup is the main struct for icon lookup operations
type IconLookup struct {
	baseDirs     []string
	currentTheme string
	themeCache   map[string]*IconTheme
	subdirCache  map[string]*SubdirInfo
}

// NewIconLookup creates a new IconLookup instance
func NewIconLookup() *IconLookup {
	il := &IconLookup{
		baseDirs:     getDefaultBaseDirs(),
		currentTheme: getCurrentTheme(),
		themeCache:   make(map[string]*IconTheme),
		subdirCache:  make(map[string]*SubdirInfo),
	}
	// DebugThemeInfo shows detailed information about a theme's structure

	return il
}

// TestDirectHicolorLookup directly tests if we can find the icon in hicolor
func (il *IconLookup) TestDirectHicolorLookup(iconName string) {
	fmt.Printf("=== Direct Hicolor Test for %s ===\n", iconName)

	// Test the exact path we know exists
	knownPath := "/usr/share/icons/hicolor/scalable/actions/blueman-send-symbolic.svg"
	if fileExists(knownPath) {
		fmt.Printf("✅ File exists at known location: %s\n", knownPath)
	} else {
		fmt.Printf("❌ File doesn't exist at known location: %s\n", knownPath)
	}

	// Try to look up the icon directly in hicolor theme
	result := il.lookupIcon(iconName, 48, 1, "hicolor")
	if result != "" {
		fmt.Printf("✅ Direct hicolor lookup found: %s\n", result)
	} else {
		fmt.Printf("❌ Direct hicolor lookup failed\n")
	}

	// Test the full lookup process
	result2 := il.findIconHelper(iconName, 48, 1, "hicolor")
	if result2 != "" {
		fmt.Printf("✅ Hicolor helper found: %s\n", result2)
	} else {
		fmt.Printf("❌ Hicolor helper failed\n")
	}
}

// NewIconLookupWithTheme creates a new IconLookup instance with a specific theme
func NewIconLookupWithTheme(theme string) *IconLookup {
	il := NewIconLookup()
	il.currentTheme = theme
	return il
}

// Lookup finds an icon with the given name, using default size (48) and scale (1)
func (il *IconLookup) Lookup(name string) (string, error) {
	return il.FindIcon(name, 48, 1)
}

// LookupWithSize finds an icon with the given name and size, using default scale (1)
func (il *IconLookup) LookupWithSize(name string, size int) (string, error) {
	return il.FindIcon(name, size, 1)
}

// FindIcon implements the main FindIcon algorithm from the spec
func (il *IconLookup) FindIcon(icon string, size, scale int) (string, error) {
	// First try user selected theme
	filename := il.findIconHelper(icon, size, scale, il.currentTheme)
	if filename != "" {
		return filename, nil
	}

	// Then try hicolor theme
	filename = il.findIconHelper(icon, size, scale, "hicolor")
	if filename != "" {
		return filename, nil
	}

	// Finally try fallback
	filename = il.lookupFallbackIcon(icon)
	if filename != "" {
		return filename, nil
	}

	return "", fmt.Errorf("icon '%s' not found", icon)
}

// FindBestIcon finds the first available icon from a list of icon names
func (il *IconLookup) FindBestIcon(iconList []string, size, scale int) (string, error) {
	// First try user selected theme
	filename := il.findBestIconHelper(iconList, size, scale, il.currentTheme)
	if filename != "" {
		return filename, nil
	}

	// Then try hicolor theme
	filename = il.findBestIconHelper(iconList, size, scale, "hicolor")
	if filename != "" {
		return filename, nil
	}

	// Finally try fallback for each icon in the list
	for _, icon := range iconList {
		filename = il.lookupFallbackIcon(icon)
		if filename != "" {
			return filename, nil
		}
	}

	return "", fmt.Errorf("none of the icons %v found", iconList)
}

// findIconHelper implements the FindIconHelper algorithm
func (il *IconLookup) findIconHelper(icon string, size, scale int, theme string) string {
	filename := il.lookupIcon(icon, size, scale, theme)
	if filename != "" {
		return filename
	}

	themeInfo := il.getThemeInfo(theme)
	if themeInfo != nil && len(themeInfo.Parents) > 0 {
		for _, parent := range themeInfo.Parents {
			filename = il.findIconHelper(icon, size, scale, parent)
			if filename != "" {
				return filename
			}
		}
	}

	return ""
}

// findBestIconHelper implements the FindBestIconHelper algorithm
func (il *IconLookup) findBestIconHelper(iconList []string, size, scale int, theme string) string {
	for _, icon := range iconList {
		filename := il.lookupIcon(icon, size, scale, theme)
		if filename != "" {
			return filename
		}
	}

	themeInfo := il.getThemeInfo(theme)
	if themeInfo != nil && len(themeInfo.Parents) > 0 {
		for _, parent := range themeInfo.Parents {
			filename := il.findBestIconHelper(iconList, size, scale, parent)
			if filename != "" {
				return filename
			}
		}
	}

	return ""
}

// lookupIcon implements the LookupIcon algorithm
func (il *IconLookup) lookupIcon(iconname string, size, scale int, theme string) string {
	themeInfo := il.getThemeInfo(theme)
	if themeInfo == nil {
		return ""
	}

	extensions := []string{"png", "svg", "xpm"}
	isSymbolic := strings.HasSuffix(iconname, "-symbolic")

	// First pass: exact matches in regular directories
	for _, subdir := range themeInfo.Directories {
		for _, directory := range il.baseDirs {
			for _, extension := range extensions {
				if il.directoryMatchesSize(subdir, size, scale, theme) {
					filename := filepath.Join(directory, theme, subdir, iconname+"."+extension)
					if fileExists(filename) {
						return filename
					}
				}
			}
		}
	}

	// Special handling for symbolic icons (same logic as debug function)
	if isSymbolic {
		for _, directory := range il.baseDirs {
			// Try /symbolic/ subdirectory
			symbolicPath := filepath.Join(directory, theme, "symbolic")
			if dirExists(symbolicPath) {
				for _, extension := range extensions {
					filename := filepath.Join(symbolicPath, iconname+"."+extension)
					if fileExists(filename) {
						return filename
					}
				}
			}

			// Try context-specific symbolic directories
			symbolicSubdirs := []string{
				"symbolic/actions", "symbolic/apps", "symbolic/devices",
				"symbolic/status", "symbolic/categories", "symbolic/emblems", "symbolic/mimetypes",
			}

			for _, symbolicSubdir := range symbolicSubdirs {
				for _, extension := range extensions {
					filename := filepath.Join(directory, theme, symbolicSubdir, iconname+"."+extension)
					if fileExists(filename) {
						return filename
					}
				}
			}

			// Also try .symbolic.png format in regular size directories
			for _, subdir := range themeInfo.Directories {
				symbolicFilename := filepath.Join(directory, theme, subdir, iconname+".symbolic.png")
				if fileExists(symbolicFilename) {
					return symbolicFilename
				}
			}
		}
	}

	// Second pass: closest match
	minimalSize := math.MaxInt32
	var closestFilename string

	for _, subdir := range themeInfo.Directories {
		for _, directory := range il.baseDirs {
			for _, extension := range extensions {
				filename := filepath.Join(directory, theme, subdir, iconname+"."+extension)
				if fileExists(filename) {
					distance := il.directorySizeDistance(subdir, size, scale, theme)
					if distance < minimalSize {
						closestFilename = filename
						minimalSize = distance
					}
				}
			}
		}
	}

	return closestFilename
}

// lookupFallbackIcon implements the LookupFallbackIcon algorithm
func (il *IconLookup) lookupFallbackIcon(iconname string) string {
	extensions := []string{"png", "svg", "xpm"}
	isSymbolic := strings.HasSuffix(iconname, "-symbolic")

	for _, directory := range il.baseDirs {
		// Standard fallback locations
		for _, extension := range extensions {
			filename := filepath.Join(directory, iconname+"."+extension)
			if fileExists(filename) {
				return filename
			}
		}

		// For symbolic icons, also try some common fallback patterns
		if isSymbolic {
			// Try .symbolic.png format
			symbolicFilename := filepath.Join(directory, iconname+".symbolic.png")
			if fileExists(symbolicFilename) {
				return symbolicFilename
			}

			// Try without -symbolic suffix as ultimate fallback
			baseName := strings.TrimSuffix(iconname, "-symbolic")
			for _, extension := range extensions {
				filename := filepath.Join(directory, baseName+"."+extension)
				if fileExists(filename) {
					return filename
				}
			}
		}
	}

	return ""
}

// directoryMatchesSize implements the DirectoryMatchesSize algorithm
func (il *IconLookup) directoryMatchesSize(subdir string, iconsize, iconscale int, theme string) bool {
	info := il.getSubdirInfo(subdir, theme)
	if info == nil {
		return false
	}

	if info.Scale != iconscale {
		return false
	}

	switch info.Type {
	case "Fixed":
		return info.Size == iconsize
	case "Scalable":
		return info.MinSize <= iconsize && iconsize <= info.MaxSize
	case "Threshold":
		return info.Size-info.Threshold <= iconsize && iconsize <= info.Size+info.Threshold
	}

	return false
}

// directorySizeDistance implements the DirectorySizeDistance algorithm
func (il *IconLookup) directorySizeDistance(subdir string, iconsize, iconscale int, theme string) int {
	info := il.getSubdirInfo(subdir, theme)
	if info == nil {
		return math.MaxInt32
	}

	switch info.Type {
	case "Fixed":
		return abs(info.Size*info.Scale - iconsize*iconscale)
	case "Scaled":
		if iconsize*iconscale < info.MinSize*info.Scale {
			return info.MinSize*info.Scale - iconsize*iconscale
		}
		if iconsize*iconscale > info.MaxSize*info.Scale {
			return iconsize*iconscale - info.MaxSize*info.Scale
		}
		return 0
	case "Threshold":
		if iconsize*iconscale < (info.Size-info.Threshold)*info.Scale {
			return info.MinSize*info.Scale - iconsize*iconscale
		}
		if iconsize*iconscale > (info.Size+info.Threshold)*info.Scale {
			return iconsize*iconscale - info.MaxSize*info.Scale
		}
		return 0
	}

	return math.MaxInt32
}

// getThemeInfo loads and caches theme information
func (il *IconLookup) getThemeInfo(theme string) *IconTheme {
	if cached, ok := il.themeCache[theme]; ok {
		return cached
	}

	for _, baseDir := range il.baseDirs {
		themeDir := filepath.Join(baseDir, theme)
		indexFile := filepath.Join(themeDir, "index.theme")

		if fileExists(indexFile) {
			themeInfo := il.parseThemeIndex(indexFile, theme)
			if themeInfo != nil {
				themeInfo.baseDirs = il.baseDirs
				il.themeCache[theme] = themeInfo
				return themeInfo
			}
		}
	}

	return nil
}

// getSubdirInfo loads and caches subdirectory information
func (il *IconLookup) getSubdirInfo(subdir, theme string) *SubdirInfo {
	key := theme + "/" + subdir
	if cached, ok := il.subdirCache[key]; ok {
		return cached
	}

	for _, baseDir := range il.baseDirs {
		indexFile := filepath.Join(baseDir, theme, "index.theme")
		if fileExists(indexFile) {
			info := il.parseSubdirInfo(indexFile, subdir)
			if info != nil {
				il.subdirCache[key] = info
				return info
			}
		}
	}

	return nil
}

// parseThemeIndex parses the theme index file
func (il *IconLookup) parseThemeIndex(indexFile, theme string) *IconTheme {
	content, err := os.ReadFile(indexFile)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	themeInfo := &IconTheme{Name: theme}

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		if currentSection == "Icon Theme" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Directories":
				themeInfo.Directories = strings.Split(value, ",")
				for i, dir := range themeInfo.Directories {
					themeInfo.Directories[i] = strings.TrimSpace(dir)
				}
			case "ScaledDirectories":
				// Add scaled directories to the main directories list
				scaledDirs := strings.Split(value, ",")
				for _, dir := range scaledDirs {
					themeInfo.Directories = append(themeInfo.Directories, strings.TrimSpace(dir))
				}
			case "Inherits":
				themeInfo.Parents = strings.Split(value, ",")
				for i, parent := range themeInfo.Parents {
					themeInfo.Parents[i] = strings.TrimSpace(parent)
				}
			}
		}
	}

	// Also add common symbolic directories that might not be explicitly listed
	symbolicDirs := []string{"symbolic", "symbolic/actions", "symbolic/apps", "symbolic/devices",
		"symbolic/status", "symbolic/categories", "symbolic/emblems", "symbolic/mimetypes"}

	// Check if these directories actually exist before adding them
	baseDir := filepath.Dir(indexFile)
	for _, symbolicDir := range symbolicDirs {
		fullPath := filepath.Join(baseDir, symbolicDir)
		if dirExists(fullPath) {
			// Only add if not already in the list
			found := false
			for _, existingDir := range themeInfo.Directories {
				if existingDir == symbolicDir {
					found = true
					break
				}
			}
			if !found {
				themeInfo.Directories = append(themeInfo.Directories, symbolicDir)
			}
		}
	}

	// For application-specific icons (like blueman), also check common fallback locations
	// that apps might install their icons to if they're not properly themed
	appSpecificDirs := []string{"scalable/status", "scalable/actions", "scalable/apps"}
	for _, appDir := range appSpecificDirs {
		fullPath := filepath.Join(baseDir, appDir)
		if dirExists(fullPath) {
			found := false
			for _, existingDir := range themeInfo.Directories {
				if existingDir == appDir {
					found = true
					break
				}
			}
			if !found {
				themeInfo.Directories = append(themeInfo.Directories, appDir)
			}
		}
	}

	// CRITICAL FIX: Many themes have scalable directories that aren't listed in index.theme
	// We need to scan for all scalable/* directories that actually exist
	scalableContexts := []string{"actions", "animations", "apps", "categories", "devices",
		"emblems", "emotes", "filesystems", "intl", "mimetypes",
		"places", "status", "stock"}

	for _, context := range scalableContexts {
		scalableDir := "scalable/" + context
		fullPath := filepath.Join(baseDir, scalableDir)
		if dirExists(fullPath) {
			found := false
			for _, existingDir := range themeInfo.Directories {
				if existingDir == scalableDir {
					found = true
					break
				}
			}
			if !found {
				themeInfo.Directories = append(themeInfo.Directories, scalableDir)
			}
		}
	}

	return themeInfo
}

// parseSubdirInfo parses subdirectory information from the theme index
func (il *IconLookup) parseSubdirInfo(indexFile, subdir string) *SubdirInfo {
	content, err := os.ReadFile(indexFile)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	info := &SubdirInfo{
		Scale:     1,
		Type:      "Threshold",
		Threshold: 2,
	}

	// Handle scalable directories that might not be in index.theme
	if strings.HasPrefix(subdir, "scalable/") {
		info.Type = "Scalable"
		info.MinSize = 1
		info.MaxSize = 512
		info.Size = 48 // Default size for scalable
		return info
	}

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		if currentSection == subdir {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Size":
				if size, err := strconv.Atoi(value); err == nil {
					info.Size = size
				}
			case "Scale":
				if scale, err := strconv.Atoi(value); err == nil {
					info.Scale = scale
				}
			case "Type":
				info.Type = value
			case "MinSize":
				if minSize, err := strconv.Atoi(value); err == nil {
					info.MinSize = minSize
				}
			case "MaxSize":
				if maxSize, err := strconv.Atoi(value); err == nil {
					info.MaxSize = maxSize
				}
			case "Threshold":
				if threshold, err := strconv.Atoi(value); err == nil {
					info.Threshold = threshold
				}
			}
		}
	}

	return info
}

// Helper functions

func getDefaultBaseDirs() []string {
	homeDir := os.Getenv("HOME")
	dirs := []string{
		filepath.Join(homeDir, ".icons"),
		filepath.Join(homeDir, ".local/share/icons"),
		"/usr/share/icons",
		"/usr/local/share/icons",
		"/usr/share/pixmaps", // Common fallback for applications
	}

	// Add XDG data directories
	if xdgDataDirs := os.Getenv("XDG_DATA_DIRS"); xdgDataDirs != "" {
		for _, dir := range strings.Split(xdgDataDirs, ":") {
			if dir != "" {
				dirs = append(dirs, filepath.Join(dir, "icons"))
			}
		}
	}

	return dirs
}

func getCurrentTheme() string {
	// Try to get theme from environment or desktop settings
	// For now, default to a common theme
	if theme := os.Getenv("ICON_THEME"); theme != "" {
		return theme
	}
	return "Adwaita" // Default GNOME theme
}

func (il *IconLookup) DebugThemeInfo(theme string) {
	fmt.Printf("=== Debug Info for Theme: %s ===\n", theme)

	themeInfo := il.getThemeInfo(theme)
	if themeInfo == nil {
		fmt.Printf("❌ Theme '%s' not found\n", theme)
		return
	}
	fmt.Printf("✅ Theme found: %s\n", themeInfo.Name)
	fmt.Printf("📁 Directories (%d total):\n", len(themeInfo.Directories))
	for i, dir := range themeInfo.Directories {
		if i < 20 { // Show first 20
			fmt.Printf("  %d: %s\n", i+1, dir)
		} else if i == 20 {
			fmt.Printf("  ... and %d more directories\n", len(themeInfo.Directories)-20)
			break
		}
	}

	fmt.Printf("👨‍👩‍👧‍👦 Parent themes: %v\n", themeInfo.Parents)

	// Check base directories
	fmt.Printf("🗂️ Base directories being searched:\n")
	for i, baseDir := range il.baseDirs {
		fmt.Printf("  %d: %s\n", i+1, baseDir)
	}

	// Specifically check for the problematic file
	targetFile := "blueman-send-symbolic.svg"
	fmt.Printf("\n🔍 Searching for %s in theme %s:\n", targetFile, theme)

	for _, baseDir := range il.baseDirs {
		themeDir := filepath.Join(baseDir, theme)
		if !dirExists(themeDir) {
			fmt.Printf("  ❌ Theme directory doesn't exist: %s\n", themeDir)
			continue
		}

		fmt.Printf("  📂 Checking theme directory: %s\n", themeDir)

		for _, subdir := range themeInfo.Directories {
			fullPath := filepath.Join(themeDir, subdir, targetFile)
			if fileExists(fullPath) {
				fmt.Printf("    ✅ FOUND: %s\n", fullPath)
			}
		}

		// Also check if scalable/actions exists but isn't in the directories list
		scaleableActions := filepath.Join(themeDir, "scalable/actions", targetFile)
		if fileExists(scaleableActions) {
			fmt.Printf("    ⚠️ FOUND BUT NOT IN DIRECTORY LIST: %s\n", scaleableActions)
		}
	}
}

func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && info.IsDir()
}
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DebugLookup provides detailed information about where an icon is searched
func (il *IconLookup) DebugLookup(iconname string, size, scale int) (string, []string, error) {
	var searchPaths []string

	// Helper function to add search path info
	addSearchPath := func(path, reason string) {
		searchPaths = append(searchPaths, fmt.Sprintf("%s (%s)", path, reason))
	}

	// First try user selected theme
	result := il.debugFindIconHelper(iconname, size, scale, il.currentTheme, &searchPaths, addSearchPath)
	if result != "" {
		return result, searchPaths, nil
	}

	// Then try hicolor theme
	result = il.debugFindIconHelper(iconname, size, scale, "hicolor", &searchPaths, addSearchPath)
	if result != "" {
		return result, searchPaths, nil
	}

	// Finally try fallback
	for _, directory := range il.baseDirs {
		extensions := []string{"png", "svg", "xpm"}
		for _, extension := range extensions {
			filename := filepath.Join(directory, iconname+"."+extension)
			addSearchPath(filename, "fallback")
			if fileExists(filename) {
				return filename, searchPaths, nil
			}
		}
	}

	return "", searchPaths, fmt.Errorf("icon '%s' not found", iconname)
}

func (il *IconLookup) debugFindIconHelper(icon string, size, scale int, theme string, paths *[]string, addPath func(string, string)) string {
	// Use the same logic as the main lookup, but with path tracking
	result := il.lookupIconWithPathTracking(icon, size, scale, theme, addPath)

	// Try parent themes if not found
	if result == "" {
		themeInfo := il.getThemeInfo(theme)
		if themeInfo != nil && len(themeInfo.Parents) > 0 {
			for _, parent := range themeInfo.Parents {
				result = il.debugFindIconHelper(icon, size, scale, parent, paths, addPath)
				if result != "" {
					return result
				}
			}
		}
	}

	return result
}

// lookupIconWithPathTracking is the same as lookupIcon but tracks search paths for debugging
func (il *IconLookup) lookupIconWithPathTracking(iconname string, size, scale int, theme string, addPath func(string, string)) string {
	themeInfo := il.getThemeInfo(theme)
	if themeInfo == nil {
		addPath(fmt.Sprintf("Theme '%s' not found", theme), "theme missing")
		return ""
	}

	extensions := []string{"png", "svg", "xpm"}
	isSymbolic := strings.HasSuffix(iconname, "-symbolic")

	// First pass: exact matches in regular directories
	for _, subdir := range themeInfo.Directories {
		for _, directory := range il.baseDirs {
			for _, extension := range extensions {
				if il.directoryMatchesSize(subdir, size, scale, theme) {
					filename := filepath.Join(directory, theme, subdir, iconname+"."+extension)
					addPath(filename, "exact size match")
					if fileExists(filename) {
						return filename
					}
				}
			}
		}
	}

	// Special handling for symbolic icons
	if isSymbolic {
		for _, directory := range il.baseDirs {
			// Try /symbolic/ subdirectory
			symbolicPath := filepath.Join(directory, theme, "symbolic")
			if dirExists(symbolicPath) {
				for _, extension := range extensions {
					filename := filepath.Join(symbolicPath, iconname+"."+extension)
					addPath(filename, "symbolic directory")
					if fileExists(filename) {
						return filename
					}
				}
			}

			// Try context-specific symbolic directories
			symbolicSubdirs := []string{
				"symbolic/actions", "symbolic/apps", "symbolic/devices",
				"symbolic/status", "symbolic/categories", "symbolic/emblems", "symbolic/mimetypes",
			}

			for _, symbolicSubdir := range symbolicSubdirs {
				for _, extension := range extensions {
					filename := filepath.Join(directory, theme, symbolicSubdir, iconname+"."+extension)
					addPath(filename, "symbolic context")
					if fileExists(filename) {
						return filename
					}
				}
			}

			// Also try .symbolic.png format in regular size directories
			for _, subdir := range themeInfo.Directories {
				symbolicFilename := filepath.Join(directory, theme, subdir, iconname+".symbolic.png")
				addPath(symbolicFilename, "symbolic.png format")
				if fileExists(symbolicFilename) {
					return symbolicFilename
				}
			}
		}
	}

	// Second pass: closest match
	minimalSize := math.MaxInt32
	var closestFilename string

	for _, subdir := range themeInfo.Directories {
		for _, directory := range il.baseDirs {
			for _, extension := range extensions {
				filename := filepath.Join(directory, theme, subdir, iconname+"."+extension)
				addPath(filename, "size fallback")
				if fileExists(filename) {
					distance := il.directorySizeDistance(subdir, size, scale, theme)
					if distance < minimalSize {
						closestFilename = filename
						minimalSize = distance
					}
				}
			}
		}
	}

	return closestFilename
}

// LookupSymbolic finds a symbolic icon with the given name, using default size (48) and scale (1)
func (il *IconLookup) LookupSymbolic(name string) (string, error) {
	// Ensure the name has the -symbolic suffix
	if !strings.HasSuffix(name, "-symbolic") {
		name = name + "-symbolic"
	}
	return il.FindIcon(name, 48, 1)
}

// LookupSymbolicWithSize finds a symbolic icon with the given name and size, using default scale (1)
func (il *IconLookup) LookupSymbolicWithSize(name string, size int) (string, error) {
	// Ensure the name has the -symbolic suffix
	if !strings.HasSuffix(name, "-symbolic") {
		name = name + "-symbolic"
	}
	return il.FindIcon(name, size, 1)
}

// GetTheme returns the current theme
func (il *IconLookup) GetTheme() string {
	return il.currentTheme
}

// SetBaseDirs allows setting custom base directories
func (il *IconLookup) SetBaseDirs(dirs []string) {
	il.baseDirs = dirs
}

// SetTheme allows changing the current theme
func (il *IconLookup) SetTheme(theme string) {
	il.currentTheme = theme
}
