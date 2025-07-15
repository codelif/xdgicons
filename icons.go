package xdgicons

type IconLookup interface {
  // Finds a specified icon, defaults size to 48 and scale to 1
	Lookup(iconName string) (Icon, error)

  // Finds a specified icon with required size and scale
	FindIcon(iconName string, size int, scale int) (Icon, error)

  // Finds the first available icon in iconList with the required size and scale.
  // Searches in the order of listing.
	FindBestIcon(iconList []string, size int, scale int) (Icon, error)
}

// Found Icon
type Icon struct {
	// Short name of the icon
	Name string

	// Full path of the icon
	Path string

	// Unscaled size of the icon
	//
	// set to 0, if unknown
	Size int

	// Scale of the icon
	//
	// set to 0, if unknown
	Scale int

	// Minimum (unscaled) size that the icon
	// can be scaled to
	//
	// set to 0, if unknown
	MinSize int

	// Maximum (unscaled) size that the icon
	// can be scaled to
	//
	// set to 0, if unknown
	MaxSize int
}

// Theme info extracted from index.theme
type ThemeInfo struct {
	// short name of the icon theme, used in e.g. lists when
	// selecting themes.
	Name string

	// The name of the theme that this theme inherits from.
	// If an icon name is not found in the current theme,
	// it is searched for in the inherited theme
	// (and recursively in all the inherited themes).
	//
	// If no theme is specified, implementations are required
	// to add the "hicolor" theme to the inheritance tree.
	// An implementation may optionally add other default
	// themes in between the last specified theme and the
	// hicolor theme.
	//
	// Themes that are inherited from explicitly must be
	// present on the system.
	Inherits []string

	// list of subdirectories for this theme.
	// For every subdirectory there must be a section in
	// the index.theme file describing that directory.
	Directories []string

	// Additional list of subdirectories for this theme,
	// in addition to the ones in Directories.
	// These directories should only be read by
	// implementations supporting scaled directories
	// and was added to keep compatibility with old
	// implementations that don't support these.
	ScaledDirectories []string

	// map to each info of every subdirectory
	directoryMap map[string]SubDirIconInfo
}

// Common properties of icons listed under a sub-directory
//
// Sub-directory here is the inner-most directory, directly
// under which there are icon files
type SubDirIconInfo struct {
	// Nominal (unscaled) size of the icons in this directory.
	Size int

	// Target scale of of the icons in this directory.
	// Defaults to the value 1 if not present.
	// Any directory with a scale other than 1 should
	// be listed in the ScaledDirectories list rather
	// than Directories for backwards compatibility.
	Scale int

	// The type of icon sizes for the icons in this directory.
	// Valid types are Fixed, Scalable and Threshold.
	// The type decides what other keys in the section are used.
	//
	// If not specified, the default is Threshold.
	Type string

	// Specifies the maximum (unscaled) size that the icons
	// in this directory can be scaled to.
	//
	// Defaults to the value of Size if not present.
	MaxSize int

	// Specifies the minimum (unscaled) size that the icons
	// in this directory can be scaled to.
	//
	// Defaults to the value of Size if not present.
	MinSize int

	// The icons in this directory can be used if the size
	// differ at most this much from the desired (unscaled) size.
	//
	// Defaults to 2 if not present.
	Threshold int
}
