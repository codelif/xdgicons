package xdgicons

import (
	"image"
	"image/color"
)

// GenerateMissingIcon creates an n×n image.Image representing a missing icon
func GenerateMissingIcon(size int, foregroundColor color.Color) image.Image {
	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Convert foreground color to RGBA for manipulation
	r, g, b, a := foregroundColor.RGBA()
	fgColor := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}

	// Create a lighter version for border (add transparency)
	borderColor := color.RGBA{fgColor.R, fgColor.G, fgColor.B, uint8(float64(fgColor.A) * 0.6)}

	// Background is transparent (no need to fill)

	// Draw border (2px thick)
	borderWidth := max(1, size/32) // Responsive border width
	for i := 0; i < borderWidth; i++ {
		// Top and bottom borders
		for x := 0; x < size; x++ {
			img.Set(x, i, borderColor)
			img.Set(x, size-1-i, borderColor)
		}
		// Left and right borders
		for y := 0; y < size; y++ {
			img.Set(i, y, borderColor)
			img.Set(size-1-i, y, borderColor)
		}
	}

	// Draw X (cross) in the center
	crossWidth := max(2, size/16) // Responsive cross width
	center := size / 2
	crossSize := size / 3 // Size of the cross arms

	// Draw diagonal lines forming an X
	for i := -crossSize; i <= crossSize; i++ {
		for j := -crossWidth / 2; j <= crossWidth/2; j++ {
			// Main diagonal (\)
			if center+i+j >= 0 && center+i+j < size && center+i >= 0 && center+i < size {
				img.Set(center+i+j, center+i, fgColor)
			}
			// Anti-diagonal (/)
			if center-i+j >= 0 && center-i+j < size && center+i >= 0 && center+i < size {
				img.Set(center-i+j, center+i, fgColor)
			}
		}
	}

	return img
}

// Alternative version with a broken image icon style
func GenerateMissingIconBroken(size int, foregroundColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Convert foreground color to RGBA for manipulation
	r, g, b, a := foregroundColor.RGBA()
	fgColor := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}

	// Create a lighter version for border
	borderColor := color.RGBA{fgColor.R, fgColor.G, fgColor.B, uint8(float64(fgColor.A) * 0.5)}

	// Background is transparent (no need to fill)

	// Draw dashed border
	borderWidth := max(1, size/32)
	dashSize := max(3, size/16)

	for i := 0; i < borderWidth; i++ {
		// Top and bottom dashed borders
		for x := 0; x < size; x += dashSize * 2 {
			for dx := 0; dx < dashSize && x+dx < size; dx++ {
				img.Set(x+dx, i, borderColor)
				img.Set(x+dx, size-1-i, borderColor)
			}
		}
		// Left and right dashed borders
		for y := 0; y < size; y += dashSize * 2 {
			for dy := 0; dy < dashSize && y+dy < size; dy++ {
				img.Set(i, y+dy, borderColor)
				img.Set(size-1-i, y+dy, borderColor)
			}
		}
	}

	// Draw broken image icon in center
	iconSize := size / 3
	startX := (size - iconSize) / 2
	startY := (size - iconSize) / 2

	// Draw rectangle outline
	lineWidth := max(1, size/64)
	for i := 0; i < lineWidth; i++ {
		// Horizontal lines
		for x := startX; x < startX+iconSize; x++ {
			img.Set(x, startY+i, fgColor)
			img.Set(x, startY+iconSize-1-i, fgColor)
		}
		// Vertical lines
		for y := startY; y < startY+iconSize; y++ {
			img.Set(startX+i, y, fgColor)
			img.Set(startX+iconSize-1-i, y, fgColor)
		}
	}

	// Draw diagonal crack
	for i := 0; i < iconSize; i++ {
		for j := 0; j < lineWidth; j++ {
			if startX+i+j < size && startY+i < size {
				img.Set(startX+i+j, startY+i, fgColor)
			}
		}
	}

	return img
}

// Helper function for Go versions that don't have built-in max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
