package color

import (
	"image/color"
	"testing"
)

// ==============================================================================
// Color Tests
// ==============================================================================

// TestColor_ToRGBA verifies BGR to RGBA conversion
func TestColor_ToRGBA(t *testing.T) {
	testCases := []struct {
		name     string
		color    Color
		expected color.RGBA
	}{
		{
			name:     "Black",
			color:    Black,
			expected: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name:     "White",
			color:    White,
			expected: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name:     "Red",
			color:    Red,
			expected: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name:     "Green",
			color:    Green,
			expected: color.RGBA{R: 0, G: 128, B: 0, A: 255},
		},
		{
			name:     "Blue",
			color:    Blue,
			expected: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
		{
			name:     "Cyan",
			color:    Cyan,
			expected: color.RGBA{R: 0, G: 255, B: 255, A: 255},
		},
		{
			name:     "Magenta",
			color:    Magenta,
			expected: color.RGBA{R: 255, G: 0, B: 255, A: 255},
		},
		{
			name:     "Yellow",
			color:    Yellow,
			expected: color.RGBA{R: 255, G: 255, B: 0, A: 255},
		},
		{
			name:     "HotPink",
			color:    HotPink,
			expected: color.RGBA{R: 255, G: 105, B: 180, A: 255},
		},
		{
			name:     "CornflowerBlue",
			color:    CornflowerBlue,
			expected: color.RGBA{R: 100, G: 149, B: 237, A: 255},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rgba := tc.color.ToRGBA()

			if rgba.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, rgba.R)
			}
			if rgba.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, rgba.G)
			}
			if rgba.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, rgba.B)
			}
			if rgba.A != tc.expected.A {
				t.Errorf("A: expected %d, got %d", tc.expected.A, rgba.A)
			}
		})
	}
}

// TestColor_BGROrdering verifies BGR ordering is correct
func TestColor_BGROrdering(t *testing.T) {
	// Create a color with distinct R, G, B values
	c := Color{B: 10, G: 20, R: 30}

	// Verify the struct fields are in BGR order
	if c.B != 10 {
		t.Errorf("B: expected 10, got %d", c.B)
	}
	if c.G != 20 {
		t.Errorf("G: expected 20, got %d", c.G)
	}
	if c.R != 30 {
		t.Errorf("R: expected 30, got %d", c.R)
	}

	// Verify ToRGBA converts correctly
	rgba := c.ToRGBA()
	if rgba.R != 30 {
		t.Errorf("RGBA.R: expected 30, got %d", rgba.R)
	}
	if rgba.G != 20 {
		t.Errorf("RGBA.G: expected 20, got %d", rgba.G)
	}
	if rgba.B != 10 {
		t.Errorf("RGBA.B: expected 10, got %d", rgba.B)
	}
}

// ==============================================================================
// HexToBGR Tests
// ==============================================================================

// TestHexToBGR_SixChar verifies 6-character hex conversion
func TestHexToBGR_SixChar(t *testing.T) {
	testCases := []struct {
		hex      string
		expected Color
	}{
		{"#FF0000", Color{B: 0, G: 0, R: 255}},     // Red
		{"#00FF00", Color{B: 0, G: 255, R: 0}},     // Green
		{"#0000FF", Color{B: 255, G: 0, R: 0}},     // Blue
		{"#FFFFFF", Color{B: 255, G: 255, R: 255}}, // White
		{"#000000", Color{B: 0, G: 0, R: 0}},       // Black
		{"#FF6969", Color{B: 105, G: 105, R: 255}}, // Light red
		{"#69B4FF", Color{B: 255, G: 180, R: 105}}, // Light blue
	}

	for _, tc := range testCases {
		t.Run(tc.hex, func(t *testing.T) {
			result, err := HexToBGR(tc.hex)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, result.R)
			}
			if result.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, result.G)
			}
			if result.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, result.B)
			}
		})
	}
}

// TestHexToBGR_ThreeChar verifies 3-character hex conversion
func TestHexToBGR_ThreeChar(t *testing.T) {
	testCases := []struct {
		hex      string
		expected Color
	}{
		{"#F00", Color{B: 0, G: 0, R: 255}},     // Red (expanded to #FF0000)
		{"#0F0", Color{B: 0, G: 255, R: 0}},     // Green (expanded to #00FF00)
		{"#00F", Color{B: 255, G: 0, R: 0}},     // Blue (expanded to #0000FF)
		{"#FFF", Color{B: 255, G: 255, R: 255}}, // White (expanded to #FFFFFF)
		{"#000", Color{B: 0, G: 0, R: 0}},       // Black (expanded to #000000)
		{"#ABC", Color{B: 204, G: 187, R: 170}}, // Gray-ish (expanded to #AABBCC)
	}

	for _, tc := range testCases {
		t.Run(tc.hex, func(t *testing.T) {
			result, err := HexToBGR(tc.hex)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, result.R)
			}
			if result.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, result.G)
			}
			if result.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, result.B)
			}
		})
	}
}

// TestHexToBGR_NoHashPrefix verifies hex without # prefix
func TestHexToBGR_NoHashPrefix(t *testing.T) {
	testCases := []struct {
		hex      string
		expected Color
	}{
		{"FF0000", Color{B: 0, G: 0, R: 255}},
		{"00FF00", Color{B: 0, G: 255, R: 0}},
		{"F00", Color{B: 0, G: 0, R: 255}},
		{"0F0", Color{B: 0, G: 255, R: 0}},
	}

	for _, tc := range testCases {
		t.Run(tc.hex, func(t *testing.T) {
			result, err := HexToBGR(tc.hex)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, result.R)
			}
			if result.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, result.G)
			}
			if result.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, result.B)
			}
		})
	}
}

// TestHexToBGR_Lowercase verifies lowercase hex
func TestHexToBGR_Lowercase(t *testing.T) {
	testCases := []struct {
		hex      string
		expected Color
	}{
		{"#ff0000", Color{B: 0, G: 0, R: 255}},
		{"#00ff00", Color{B: 0, G: 255, R: 0}},
		{"#f00", Color{B: 0, G: 0, R: 255}},
	}

	for _, tc := range testCases {
		t.Run(tc.hex, func(t *testing.T) {
			result, err := HexToBGR(tc.hex)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, result.R)
			}
			if result.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, result.G)
			}
			if result.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, result.B)
			}
		})
	}
}

// TestHexToBGR_InvalidLength verifies error on invalid length
func TestHexToBGR_InvalidLength(t *testing.T) {
	invalidHexStrings := []string{
		"#FF",     // Too short (2 chars)
		"#FFFF",   // Invalid length (4 chars)
		"#FFFFF",  // Invalid length (5 chars)
		"#FFFFFFF", // Too long (7 chars)
		"#F",      // Too short (1 char)
		"",        // Empty
	}

	for _, hex := range invalidHexStrings {
		t.Run(hex, func(t *testing.T) {
			_, err := HexToBGR(hex)
			if err == nil {
				t.Errorf("Expected error for invalid hex '%s', got nil", hex)
			}
		})
	}
}

// TestHexToBGR_InvalidCharacters verifies error on invalid characters
func TestHexToBGR_InvalidCharacters(t *testing.T) {
	invalidHexStrings := []string{
		"#GGGGGG", // Invalid hex characters
		"#ZZZZZZ",
		"#12345G",
		"#XYZ",
		"#GGG",
		"#12345Z",
	}

	for _, hex := range invalidHexStrings {
		t.Run(hex, func(t *testing.T) {
			_, err := HexToBGR(hex)
			if err == nil {
				t.Errorf("Expected error for invalid hex '%s', got nil", hex)
			}
		})
	}
}

// TestHexToBGR_EdgeValues verifies edge values (0x00 and 0xFF)
func TestHexToBGR_EdgeValues(t *testing.T) {
	testCases := []struct {
		hex      string
		expected Color
	}{
		{"#000000", Color{B: 0, G: 0, R: 0}},       // All zeros
		{"#FFFFFF", Color{B: 255, G: 255, R: 255}}, // All max
		{"#FF0000", Color{B: 0, G: 0, R: 255}},     // Only R max
		{"#00FF00", Color{B: 0, G: 255, R: 0}},     // Only G max
		{"#0000FF", Color{B: 255, G: 0, R: 0}},     // Only B max
		{"#010101", Color{B: 1, G: 1, R: 1}},       // All ones
		{"#FEFEFE", Color{B: 254, G: 254, R: 254}}, // All max-1
	}

	for _, tc := range testCases {
		t.Run(tc.hex, func(t *testing.T) {
			result, err := HexToBGR(tc.hex)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, result.R)
			}
			if result.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, result.G)
			}
			if result.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, result.B)
			}
		})
	}
}

// ==============================================================================
// Color Constants Tests
// ==============================================================================

// TestColorConstants verifies all color constants are correct
func TestColorConstants(t *testing.T) {
	testCases := []struct {
		name     string
		color    Color
		expected Color
	}{
		{"Black", Black, Color{B: 0, G: 0, R: 0}},
		{"White", White, Color{B: 255, G: 255, R: 255}},
		{"Red", Red, Color{B: 0, G: 0, R: 255}},
		{"Green", Green, Color{B: 0, G: 128, R: 0}},
		{"Blue", Blue, Color{B: 255, G: 0, R: 0}},
		{"Cyan", Cyan, Color{B: 255, G: 255, R: 0}},
		{"Magenta", Magenta, Color{B: 255, G: 0, R: 255}},
		{"Yellow", Yellow, Color{B: 0, G: 255, R: 255}},
		{"HotPink", HotPink, Color{B: 180, G: 105, R: 255}},
		{"CornflowerBlue", CornflowerBlue, Color{B: 237, G: 149, R: 100}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.color.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, tc.color.R)
			}
			if tc.color.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, tc.color.G)
			}
			if tc.color.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, tc.color.B)
			}
		})
	}
}

// TestHexToBGR_RoundTrip verifies consistency with predefined colors
func TestHexToBGR_RoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		hex      string
		expected Color
	}{
		{"Red", "#FF0000", Red},
		{"Blue", "#0000FF", Blue},
		{"White", "#FFFFFF", White},
		{"Black", "#000000", Black},
		{"Cyan", "#00FFFF", Cyan},
		{"Magenta", "#FF00FF", Magenta},
		{"Yellow", "#FFFF00", Yellow},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := HexToBGR(tc.hex)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.R != tc.expected.R || result.G != tc.expected.G || result.B != tc.expected.B {
				t.Errorf("Expected %+v, got %+v", tc.expected, result)
			}
		})
	}
}
