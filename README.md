# xdgicons
Go Lookup functions for https://specifications.freedesktop.org/icon-theme-spec/latest/#icon_lookup

## Features
- Cached lookups
- Full spec coverage
- Contains a basic "missing icon" icon generation API (xdgicons/missing)
- Contains a basic SVG rendering API (xdgicons/renderer)


## Installation

```bash
go get github.com/codelif/xdgicons
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/codelif/xdgicons"
)

func main() {
    // Create icon lookup instance with default icon theme (taken from gsettings)
    // This also sets up directory cache, this call may take a few hundred milliseconds
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
// Simple lookup with default size (48px@1)
iconPath, err := iconLookup.Lookup("folder")

// Full control over size and scale
icon, err := iconLookup.FindIcon("text-editor", 32, 2) // 32px at 2x scale
if err == nil {
    fmt.Println(icon.Path)
}

```


### SVG Icon Rendering

```go
import (
    "image/color"

    "github.com/codelif/xdgicons/renderer"
)

// Create renderer
renderer := renderer.NewIconRenderer(iconLookup)

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

### "Missing Icon" Icon Generation
```go
import (
    "image/color"

    "github.com/codelif/xdgicons/missing"
)

// Creates a 48x48px image with white foreground and transparent background
// design is a square box with a X in between. Returns an image.Image
missingImg := GenerateMissingIcon(48, color.WHITE)

// Alternate broken glass/window design.
missingImg2 := GenerateMissingIconBroken(48, color.RED)

// calls to these functions are cached, so they can be called
// multiple times without any performance implications
```

## Note
This used to be AI slop, but every line of code has been rewritten by hand (atleast in the `xdgicons` package). This has been done because the AI code was hot garbage.
