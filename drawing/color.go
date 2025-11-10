package drawing

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/nmichlo/norfair-go/color"
	"github.com/nmichlo/norfair-go/internal/imaging"
)

// Color is an alias for color.Color (BGR format for OpenCV).
type Color = color.Color

// =============================================================================
// Palette - Color selection via hashing
// =============================================================================

// Palette manages color selection for drawing tracked objects.
// Uses deterministic hashing to assign colors based on object IDs.
type Palette struct {
	colors       []Color
	defaultColor Color
}

// Note: Palette colors (Tab10, Tab20, Colorblind) moved to internal/imaging package

// NewPalette creates a new Palette with the given colors.
// If colors is nil or empty, uses tab10 palette by default.
func NewPalette(colors []Color) *Palette {
	if len(colors) == 0 {
		colors = imaging.Tab10
	}

	return &Palette{
		colors:       colors,
		defaultColor: color.White,
	}
}

// ChooseColor selects a color based on a hashable value (typically object ID).
// Uses FNV-1a hash for deterministic color assignment.
func (p *Palette) ChooseColor(hashable interface{}) Color {
	if hashable == nil {
		return p.defaultColor
	}

	// Hash the value
	h := fnv.New32a()
	fmt.Fprintf(h, "%v", hashable)
	hash := h.Sum32()

	// Select color based on hash
	idx := int(hash) % len(p.colors)
	return p.colors[idx]
}

// Set changes the palette to a named palette.
// Supported names: "tab10", "tab20", "colorblind".
func (p *Palette) Set(paletteName string) error {
	switch strings.ToLower(paletteName) {
	case "tab10":
		p.colors = imaging.Tab10
	case "tab20":
		p.colors = imaging.Tab20
	case "colorblind":
		p.colors = imaging.Colorblind
	default:
		return fmt.Errorf("unknown palette: %s (supported: tab10, tab20, colorblind)", paletteName)
	}
	return nil
}

// SetDefaultColor sets the default color (used when hashable is nil).
func (p *Palette) SetDefaultColor(color Color) {
	p.defaultColor = color
}

// =============================================================================
// Helper Functions
// =============================================================================

// HexToBGR converts a hex color string to BGR Color.
// Supports both 3-char (#RGB) and 6-char (#RRGGBB) formats.
func HexToBGR(hex string) (Color, error) {
	return color.HexToBGR(hex)
}

// ParseColorName looks up a color by name (case-insensitive).
func ParseColorName(name string) (Color, error) {
	imgColor, ok := imaging.ColorMap[strings.ToLower(name)]
	if !ok {
		return Color{}, fmt.Errorf("unknown color name: %s", name)
	}
	return Color(imgColor), nil
}
