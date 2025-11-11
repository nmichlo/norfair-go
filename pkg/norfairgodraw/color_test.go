package norfairgodraw

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/imaging"
	"github.com/nmichlo/norfair-go/pkg/norfairgocolor"
)

// =============================================================================
// Hex Color Parsing Tests
// =============================================================================

func TestHexToBGR_6CharFormat(t *testing.T) {
	// Test 6-character hex format (#RRGGBB)
	tests := []struct {
		hex      string
		expected Color
	}{
		{"#FF0000", Color{B: 0, G: 0, R: 255}},     // Red
		{"#00FF00", Color{B: 0, G: 255, R: 0}},     // Green
		{"#0000FF", Color{B: 255, G: 0, R: 0}},     // Blue
		{"#FFFFFF", Color{B: 255, G: 255, R: 255}}, // White
		{"#000000", Color{B: 0, G: 0, R: 0}},       // Black
		{"FF00FF", Color{B: 255, G: 0, R: 255}},    // Magenta (without #)
	}

	for _, tt := range tests {
		result, err := HexToBGR(tt.hex)
		if err != nil {
			t.Errorf("HexToBGR(%s) returned error: %v", tt.hex, err)
			continue
		}

		if result != tt.expected {
			t.Errorf("HexToBGR(%s) = %+v, want %+v", tt.hex, result, tt.expected)
		}
	}
}

func TestHexToBGR_3CharFormat(t *testing.T) {
	// Test 3-character hex format (#RGB)
	tests := []struct {
		hex      string
		expected Color
	}{
		{"#F00", Color{B: 0, G: 0, R: 255}},     // Red
		{"#0F0", Color{B: 0, G: 255, R: 0}},     // Green
		{"#00F", Color{B: 255, G: 0, R: 0}},     // Blue
		{"#FFF", Color{B: 255, G: 255, R: 255}}, // White
		{"#000", Color{B: 0, G: 0, R: 0}},       // Black
		{"F0F", Color{B: 255, G: 0, R: 255}},    // Magenta (without #)
	}

	for _, tt := range tests {
		result, err := HexToBGR(tt.hex)
		if err != nil {
			t.Errorf("HexToBGR(%s) returned error: %v", tt.hex, err)
			continue
		}

		if result != tt.expected {
			t.Errorf("HexToBGR(%s) = %+v, want %+v", tt.hex, result, tt.expected)
		}
	}
}

func TestHexToBGR_InvalidFormats(t *testing.T) {
	// Test invalid hex formats
	invalid := []string{
		"#FF",      // Too short
		"#FFFFFFF", // Too long
		"#GGGGGG",  // Invalid hex characters
		"#12345",   // Wrong length
		"",         // Empty
	}

	for _, hex := range invalid {
		_, err := HexToBGR(hex)
		if err == nil {
			t.Errorf("HexToBGR(%s) should return error for invalid format", hex)
		}
	}
}

// =============================================================================
// Color Name Lookup Tests
// =============================================================================

func TestParseColorName_ValidNames(t *testing.T) {
	tests := []struct {
		name     string
		expected Color
	}{
		{"red", norfairgocolor.Red},
		{"green", norfairgocolor.Green},
		{"blue", norfairgocolor.Blue},
		{"white", norfairgocolor.White},
		{"black", norfairgocolor.Black},
		{"Red", norfairgocolor.Red},     // Case insensitive
		{"GREEN", norfairgocolor.Green}, // Case insensitive
		{"BlUe", norfairgocolor.Blue},   // Mixed case
		{"hotpink", norfairgocolor.HotPink},
		{"cornflowerblue", norfairgocolor.CornflowerBlue},
	}

	for _, tt := range tests {
		result, err := ParseColorName(tt.name)
		if err != nil {
			t.Errorf("ParseColorName(%s) returned error: %v", tt.name, err)
			continue
		}

		if result != tt.expected {
			t.Errorf("ParseColorName(%s) = %+v, want %+v", tt.name, result, tt.expected)
		}
	}
}

func TestParseColorName_InvalidNames(t *testing.T) {
	invalid := []string{
		"notacolor",
		"redd",
		"greeen",
		"",
		"color123",
	}

	for _, name := range invalid {
		_, err := ParseColorName(name)
		if err == nil {
			t.Errorf("ParseColorName(%s) should return error for invalid name", name)
		}
	}
}

// =============================================================================
// Palette Tests
// =============================================================================

func TestPalette_NewPalette(t *testing.T) {
	// Test default palette (tab10)
	p := NewPalette(nil)
	if len(p.colors) != len(imaging.Tab10) {
		t.Errorf("Default palette should have %d colors, got %d", len(imaging.Tab10), len(p.colors))
	}

	// Test custom palette
	customColors := []Color{norfairgocolor.Red, norfairgocolor.Green, norfairgocolor.Blue}
	p = NewPalette(customColors)
	if len(p.colors) != 3 {
		t.Errorf("Custom palette should have 3 colors, got %d", len(p.colors))
	}
}

func TestPalette_ChooseColor_Deterministic(t *testing.T) {
	p := NewPalette(nil)

	// Same hashable should always return same color
	color1 := p.ChooseColor(42)
	color2 := p.ChooseColor(42)
	color3 := p.ChooseColor(42)

	if color1 != color2 || color1 != color3 {
		t.Error("ChooseColor should return same color for same hashable")
	}

	// Different hashables should (usually) return different colors
	color4 := p.ChooseColor(43)
	// Note: Can't guarantee different colors due to hash collisions,
	// but with 10 colors, they should differ most of the time
	t.Logf("ChooseColor(42) = %+v", color1)
	t.Logf("ChooseColor(43) = %+v", color4)
}

func TestPalette_ChooseColor_NilHashable(t *testing.T) {
	p := NewPalette(nil)

	// Nil hashable should return default color
	color := p.ChooseColor(nil)
	if color != p.defaultColor {
		t.Errorf("ChooseColor(nil) should return default color %+v, got %+v", p.defaultColor, color)
	}
}

func TestPalette_Set_ValidPalettes(t *testing.T) {
	p := NewPalette(nil)

	// Test tab10
	err := p.Set("tab10")
	if err != nil {
		t.Errorf("Set(tab10) returned error: %v", err)
	}
	if len(p.colors) != len(imaging.Tab10) {
		t.Errorf("tab10 palette should have %d colors, got %d", len(imaging.Tab10), len(p.colors))
	}

	// Test tab20
	err = p.Set("tab20")
	if err != nil {
		t.Errorf("Set(tab20) returned error: %v", err)
	}
	if len(p.colors) != len(imaging.Tab20) {
		t.Errorf("tab20 palette should have %d colors, got %d", len(imaging.Tab20), len(p.colors))
	}

	// Test colorblind
	err = p.Set("colorblind")
	if err != nil {
		t.Errorf("Set(colorblind) returned error: %v", err)
	}
	if len(p.colors) != len(imaging.Colorblind) {
		t.Errorf("colorblind palette should have %d colors, got %d", len(imaging.Colorblind), len(p.colors))
	}

	// Test case insensitivity
	err = p.Set("TAB10")
	if err != nil {
		t.Errorf("Set(TAB10) should be case-insensitive, got error: %v", err)
	}
}

func TestPalette_Set_InvalidPalette(t *testing.T) {
	p := NewPalette(nil)

	err := p.Set("nonexistent")
	if err == nil {
		t.Error("Set should return error for nonexistent palette")
	}
}

func TestPalette_SetDefaultColor(t *testing.T) {
	p := NewPalette(nil)

	// Change default color
	p.SetDefaultColor(norfairgocolor.Red)
	if p.defaultColor != norfairgocolor.Red {
		t.Errorf("SetDefaultColor(Red) failed, default is %+v", p.defaultColor)
	}

	// Verify nil hashable uses new default
	c := p.ChooseColor(nil)
	if c != norfairgocolor.Red {
		t.Errorf("ChooseColor(nil) should return Red, got %+v", c)
	}
}

// =============================================================================
// Color Conversion Tests
// =============================================================================

func TestColor_ToRGBA(t *testing.T) {
	tests := []struct {
		color     Color
		expectedR uint8
		expectedG uint8
		expectedB uint8
	}{
		{norfairgocolor.Red, 255, 0, 0},
		{norfairgocolor.Green, 0, 128, 0},
		{norfairgocolor.Blue, 0, 0, 255},
		{norfairgocolor.White, 255, 255, 255},
		{norfairgocolor.Black, 0, 0, 0},
	}

	for _, tt := range tests {
		rgba := tt.color.ToRGBA()
		if rgba.R != tt.expectedR || rgba.G != tt.expectedG || rgba.B != tt.expectedB {
			t.Errorf("ToRGBA(%+v) = RGBA{%d, %d, %d}, want RGBA{%d, %d, %d}",
				tt.color, rgba.R, rgba.G, rgba.B, tt.expectedR, tt.expectedG, tt.expectedB)
		}
		if rgba.A != 255 {
			t.Errorf("ToRGBA(%+v).A = %d, want 255", tt.color, rgba.A)
		}
	}
}
