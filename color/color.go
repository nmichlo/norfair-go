package color

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Color represents an OpenCV color in BGR format.
// Note: OpenCV uses BGR ordering, not RGB!
type Color struct {
	B, G, R uint8
}

// ToRGBA converts Color to color.RGBA format for gocv.
// Note: OpenCV internally uses BGR, but gocv.Circle/etc expect RGBA.
func (c Color) ToRGBA() color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
}

// Common color constants
var (
	Black          = Color{B: 0, G: 0, R: 0}
	White          = Color{B: 255, G: 255, R: 255}
	Red            = Color{B: 0, G: 0, R: 255}
	Green          = Color{B: 0, G: 128, R: 0}  // CSS Green (darker)
	Blue           = Color{B: 255, G: 0, R: 0}
	Cyan           = Color{B: 255, G: 255, R: 0}
	Magenta        = Color{B: 255, G: 0, R: 255}
	Yellow         = Color{B: 0, G: 255, R: 255}
	HotPink        = Color{B: 180, G: 105, R: 255}
	CornflowerBlue = Color{B: 237, G: 149, R: 100}
)

// HexToBGR converts a hex color string to BGR Color.
// Supports both 3-char (#RGB) and 6-char (#RRGGBB) formats.
func HexToBGR(hex string) (Color, error) {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b uint8

	if len(hex) == 3 {
		// 3-char format: #RGB -> #RRGGBB
		rVal, err := strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color: %s", hex)
		}
		gVal, err := strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color: %s", hex)
		}
		bVal, err := strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color: %s", hex)
		}
		r, g, b = uint8(rVal), uint8(gVal), uint8(bVal)
	} else if len(hex) == 6 {
		// 6-char format: #RRGGBB
		rVal, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color: %s", hex)
		}
		gVal, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color: %s", hex)
		}
		bVal, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color: %s", hex)
		}
		r, g, b = uint8(rVal), uint8(gVal), uint8(bVal)
	} else {
		return Color{}, fmt.Errorf("invalid hex color length: %s (expected 3 or 6 chars)", hex)
	}

	// Return in BGR format (OpenCV convention)
	return Color{B: b, G: g, R: r}, nil
}
