// simple svg renderer
// still AI slop, needs improvement
package renderer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/codelif/xdgicons"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// IconRenderer handles rendering icons to different formats
type IconRenderer struct {
	iconLookup *xdgicons.IconLookup
}

// NewIconRenderer creates a new icon renderer
func NewIconRenderer(iconLookup *xdgicons.IconLookup) *IconRenderer {
	return &IconRenderer{
		iconLookup: iconLookup,
	}
}

// RenderIconToPNG renders an icon to PNG format at the specified size
func (ir *IconRenderer) RenderIconToPNG(iconName string, size int) (image.Image, error) {
	// First find the icon
	iconPath, err := ir.iconLookup.FindIcon(iconName, size, 1)
	if err != nil {
		return nil, fmt.Errorf("icon not found: %w", err)
	}

	return ir.RenderFileToPNG(iconPath.Path, size)
}

// RenderFileToPNG renders an icon file to PNG format at the specified size
func (ir *IconRenderer) RenderFileToPNG(iconPath string, size int) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(iconPath))
	
	switch ext {
	case ".svg":
		return ir.renderSVGToPNG(iconPath, size)
	case ".png":
		return ir.loadAndResizePNG(iconPath, size)
	case ".xpm":
		return ir.loadAndResizeXPM(iconPath, size)
	default:
		return nil, fmt.Errorf("unsupported icon format: %s", ext)
	}
}

// renderSVGToPNG renders an SVG file to PNG
func (ir *IconRenderer) renderSVGToPNG(svgPath string, size int) (image.Image, error) {
	// Read SVG file
	svgFile, err := os.Open(svgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SVG file: %w", err)
	}
	defer svgFile.Close()

	// Parse SVG
	icon, err := oksvg.ReadIconStream(svgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG: %w", err)
	}

	// Set the target size
	icon.SetTarget(0, 0, float64(size), float64(size))

	// Create image canvas
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Create scanner for rasterization
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	
	// Create rasterizer
	raster := rasterx.NewDasher(size, size, scanner)

	// Render SVG to the image
	icon.Draw(raster, 1.0)

	return img, nil
}

// RenderSymbolicSVGToPNG renders a symbolic SVG with custom colors
func (ir *IconRenderer) RenderSymbolicSVGToPNG(svgPath string, size int, foregroundColor color.Color) (image.Image, error) {
	// Read and modify SVG content for symbolic icons
	svgContent, err := os.ReadFile(svgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SVG file: %w", err)
	}

	// For symbolic icons, replace the default color with the desired foreground color
	modifiedSVG := ir.replaceSymbolicColors(string(svgContent), foregroundColor)

	// Create a temporary reader
	svgReader := strings.NewReader(modifiedSVG)

	// Parse modified SVG
	icon, err := oksvg.ReadIconStream(svgReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modified SVG: %w", err)
	}

	// Set the target size
	icon.SetTarget(0, 0, float64(size), float64(size))

	// Create image canvas with transparent background
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Create scanner for rasterization
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	
	// Create rasterizer
	raster := rasterx.NewDasher(size, size, scanner)

	// Render SVG to the image
	icon.Draw(raster, 1.0)

	return img, nil
}

// replaceSymbolicColors replaces symbolic icon colors with the desired foreground color
func (ir *IconRenderer) replaceSymbolicColors(svgContent string, foregroundColor color.Color) string {
	r, g, b, _ := foregroundColor.RGBA()
	// Convert to 8-bit values
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	hexColor := fmt.Sprintf("#%02x%02x%02x", r8, g8, b8)

	// Replace common symbolic icon colors
	replacements := map[string]string{
		"#bebebe": hexColor, // Default symbolic color
		"#2e3436": hexColor, // Dark symbolic color
		"#000000": hexColor, // Black
		"#ffffff": hexColor, // White
		"currentColor": hexColor, // CSS currentColor
	}
  
	modifiedSVG := svgContent
	for oldColor, newColor := range replacements {
		modifiedSVG = strings.ReplaceAll(modifiedSVG, oldColor, newColor)
	}

	return modifiedSVG
}

// loadAndResizePNG loads a PNG file and optionally resizes it
func (ir *IconRenderer) loadAndResizePNG(pngPath string, targetSize int) (image.Image, error) {
	file, err := os.Open(pngPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PNG file: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// If the image is already the right size, return it as-is
	bounds := img.Bounds()
	if bounds.Dx() == targetSize && bounds.Dy() == targetSize {
		return img, nil
	}

	// Resize the image (simple nearest neighbor for now)
	return ir.resizeImage(img, targetSize), nil
}

// loadAndResizeXPM loads an XPM file (basic implementation)
func (ir *IconRenderer) loadAndResizeXPM(xpmPath string, targetSize int) (image.Image, error) {
	// XPM support is limited - for now, create a placeholder
	// You'd need a proper XPM decoder library for full support
	img := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	
	// Fill with a placeholder color
	gray := color.RGBA{128, 128, 128, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{gray}, image.Point{}, draw.Src)
	
	return img, nil
}

// resizeImage performs simple image resizing using nearest neighbor
func (ir *IconRenderer) resizeImage(src image.Image, targetSize int) image.Image {
	srcBounds := src.Bounds()
	srcW, srcH := srcBounds.Dx(), srcBounds.Dy()
	
	dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	
	for y := 0; y < targetSize; y++ {
		for x := 0; x < targetSize; x++ {
			srcX := x * srcW / targetSize
			srcY := y * srcH / targetSize
			dst.Set(x, y, src.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
		}
	}
	
	return dst
}

// SaveImageToPNG saves an image to a PNG file
func (ir *IconRenderer) SaveImageToPNG(img image.Image, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	return png.Encode(file, img)
}

// RenderIconWithFallback renders an icon with smart fallbacks and returns the image
func (ir *IconRenderer) RenderIconWithFallback(iconName string, size int, symbolicColor *color.Color) (image.Image, string, error) {
	// Try to find the icon
	icon, err := ir.iconLookup.FindBestIcon([]string{iconName}, size, 1)
	if err != nil {
		return nil, "", fmt.Errorf("icon not found: %w", err)
	}
  iconPath := icon.Path

	// Check if it's a symbolic icon and we have a color specified
	if strings.HasSuffix(iconName, "-symbolic") && symbolicColor != nil && strings.HasSuffix(iconPath, ".svg") {
		img, err := ir.RenderSymbolicSVGToPNG(iconPath, size, *symbolicColor)
		return img, iconPath, err
	}

	// Regular rendering
	img, err := ir.RenderFileToPNG(iconPath, size)
	return img, iconPath, err
}

