# xdgicons
Go Lookup functions for https://specifications.freedesktop.org/icon-theme-spec/latest/#icon_lookup

Also contains a basic SVG rendering API

## WARNING: This is AI slop
I wanted to get on with my other projects. And not deal with this spec. So I had this made. Seems to be working p much, though I can't guarantee.

Below are some AI slop examples as well:

### Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/codelif/xdgicons"
)

func main() {
    // Create icon lookup instance
    iconLookup := xdgicons.NewIconLookup()
    
    // Find an icon
    iconPath, err := iconLookup.Lookup("firefox")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Firefox icon: %s\n", iconPath)
    // Output: Firefox icon: /usr/share/icons/hicolor/48x48/apps/firefox.png
}
```

### Basic Icon Lookup

```go
// Simple lookup with default size (48px)
iconPath, err := iconLookup.Lookup("folder")

// Lookup with specific size
iconPath, err := iconLookup.LookupWithSize("folder", 24)

// Full control over size and scale
iconPath, err := iconLookup.FindIcon("text-editor", 32, 2) // 32px at 2x scale

// Symbolic icons
iconPath, err := iconLookup.LookupSymbolic("bluetooth") // Finds bluetooth-symbolic
```

### Smart Fallbacks

```go
// Try multiple icon names with intelligent fallbacks
iconPath, err := iconLookup.FindBestIconWithFallbacks(
    []string{"blueman-send-symbolic"}, 48, 1)
// Will try: blueman-send-symbolic → blueman-send → document-send-symbolic → document-send

// Multiple candidates (useful for MIME types)
iconPath, err := iconLookup.FindBestIcon(
    []string{"text-x-python", "text-x-script", "text-plain"}, 48, 1)
```

### SVG Icon Rendering

```go
import (
    "image/color"

    "github.com/codelif/xdgicons"
)

// Create renderer
renderer := xdgicons.NewIconRenderer(iconLookup)

// Render SVG to PNG
img, err := renderer.RenderIconToPNG("firefox", 48)

// Render symbolic icon with custom color
blueColor := color.RGBA{0, 120, 215, 255}
img, err := renderer.RenderSymbolicSVGToPNG(
    "/usr/share/icons/Adwaita/symbolic/devices/bluetooth-symbolic.svg",
    32, blueColor)

// Smart rendering with fallbacks
img, iconPath, err := renderer.RenderIconWithFallback("bluetooth-symbolic", 24, &blueColor)
```

### Theme Management

```go
// Use specific theme
iconLookup := NewIconLookupWithTheme("Papirus")

// Change theme dynamically
iconLookup.SetTheme("Adwaita")

// Check current theme
fmt.Println("Current theme:", iconLookup.GetTheme())

// Set custom icon directories
iconLookup.SetBaseDirs([]string{
    "/home/user/.local/share/icons",
    "/usr/share/icons",
})
```

## Debugging Missing Icons

```go
// Debug why an icon isn't found
iconPath, searchPaths, err := iconLookup.DebugLookup("missing-icon", 48, 1)
if err != nil {
    fmt.Printf("Icon not found. Searched %d paths:\n", len(searchPaths))
    for i, path := range searchPaths[:5] { // Show first 5
        fmt.Printf("  %d: %s\n", i+1, path)
    }
}

// Get suggestions for alternative names
suggestions := iconLookup.SuggestIconAlternatives("blueman-send-symbolic")
fmt.Printf("Try these instead: %v\n", suggestions)
// Output: [blueman-send document-send mail-send document-send-symbolic ...]

// Debug theme structure
iconLookup.DebugThemeInfo("hicolor")
```

## Installation

```bash
go get github.com/codelif/xdgicons
```

