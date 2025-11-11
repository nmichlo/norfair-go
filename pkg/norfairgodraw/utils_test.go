package norfairgodraw

import (
	"testing"

	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Centroid Tests
// =============================================================================

func TestCentroid_SimplePoints(t *testing.T) {
	// Test with a square of points
	points := mat.NewDense(4, 2, []float64{
		0, 0,
		10, 0,
		10, 10,
		0, 10,
	})

	x, y := Centroid(points)

	// Center of square should be (5, 5)
	if x != 5 || y != 5 {
		t.Errorf("Centroid = (%d, %d), want (5, 5)", x, y)
	}
}

func TestCentroid_SinglePoint(t *testing.T) {
	// Test with a single point
	points := mat.NewDense(1, 2, []float64{42, 24})

	x, y := Centroid(points)

	// Centroid of single point should be the point itself
	if x != 42 || y != 24 {
		t.Errorf("Centroid = (%d, %d), want (42, 24)", x, y)
	}
}

func TestCentroid_NegativeCoords(t *testing.T) {
	// Test with negative coordinates
	points := mat.NewDense(2, 2, []float64{
		-10, -10,
		10, 10,
	})

	x, y := Centroid(points)

	// Center should be (0, 0)
	if x != 0 || y != 0 {
		t.Errorf("Centroid = (%d, %d), want (0, 0)", x, y)
	}
}

func TestCentroid_ThreePoints(t *testing.T) {
	// Test with triangle
	points := mat.NewDense(3, 2, []float64{
		0, 0,
		3, 0,
		0, 3,
	})

	x, y := Centroid(points)

	// Centroid of right triangle should be (1, 1)
	if x != 1 || y != 1 {
		t.Errorf("Centroid = (%d, %d), want (1, 1)", x, y)
	}
}

// Note: TestCentroid_EmptyPoints removed - mat.NewDense panics on 0-row matrices
// In practice, Centroid is never called with empty point sets

func TestCentroid_FloatRounding(t *testing.T) {
	// Test rounding behavior
	points := mat.NewDense(2, 2, []float64{
		0, 0,
		1, 1,
	})

	x, y := Centroid(points)

	// Average is (0.5, 0.5), should round down to (0, 0) with int conversion
	if x != 0 || y != 0 {
		t.Errorf("Centroid = (%d, %d), want (0, 0)", x, y)
	}
}

// =============================================================================
// BuildText Tests
// =============================================================================

func TestBuildText_AllFields(t *testing.T) {
	// Test with label, ID, and scores
	label := "person"
	id := 42
	drawable := &Drawable{
		Label:  &label,
		ID:     &id,
		Scores: []float64{0.95, 0.87, 0.92},
	}

	text := BuildText(drawable, true, true, true)
	// Mean of [0.95, 0.87, 0.92] = 2.74/3 = 0.9133...
	expected := "person-42-0.9133"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_OnlyLabel(t *testing.T) {
	// Test with only label
	label := "car"
	drawable := &Drawable{
		Label:  &label,
		ID:     nil,
		Scores: nil,
	}

	text := BuildText(drawable, true, false, false)
	expected := "car"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_OnlyID(t *testing.T) {
	// Test with only ID
	id := 123
	drawable := &Drawable{
		Label:  nil,
		ID:     &id,
		Scores: nil,
	}

	text := BuildText(drawable, false, true, false)
	expected := "123"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_OnlyScores(t *testing.T) {
	// Test with only scores
	drawable := &Drawable{
		Label:  nil,
		ID:     nil,
		Scores: []float64{0.99},
	}

	text := BuildText(drawable, false, false, true)
	// Mean of [0.99] = 0.99 (trailing zeros stripped)
	expected := "0.99"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_NoFields(t *testing.T) {
	// Test with all flags disabled
	label := "test"
	id := 1
	drawable := &Drawable{
		Label:  &label,
		ID:     &id,
		Scores: []float64{0.5},
	}

	text := BuildText(drawable, false, false, false)
	expected := ""

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_NilFields(t *testing.T) {
	// Test with nil fields but flags enabled
	drawable := &Drawable{
		Label:  nil,
		ID:     nil,
		Scores: nil,
	}

	text := BuildText(drawable, true, true, true)
	expected := ""

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_MultipleScores(t *testing.T) {
	// Test score formatting with multiple values
	drawable := &Drawable{
		Label:  nil,
		ID:     nil,
		Scores: []float64{0.123, 0.456, 0.789},
	}

	text := BuildText(drawable, false, false, true)
	// Mean of [0.123, 0.456, 0.789] = 1.368/3 = 0.456 (trailing zeros stripped)
	expected := "0.456"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_LabelAndID(t *testing.T) {
	// Test with label and ID, no scores
	label := "bike"
	id := 7
	drawable := &Drawable{
		Label:  &label,
		ID:     &id,
		Scores: nil,
	}

	text := BuildText(drawable, true, true, false)
	expected := "bike-7"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_EmptyLabel(t *testing.T) {
	// Test with empty string label (should be excluded)
	label := ""
	id := 1
	drawable := &Drawable{
		Label:  &label,
		ID:     &id,
		Scores: nil,
	}

	text := BuildText(drawable, true, true, false)
	expected := "1"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}

func TestBuildText_EmptyScores(t *testing.T) {
	// Test with empty scores slice (should be excluded)
	label := "test"
	drawable := &Drawable{
		Label:  &label,
		ID:     nil,
		Scores: []float64{},
	}

	text := BuildText(drawable, true, false, true)
	expected := "test"

	if text != expected {
		t.Errorf("BuildText() = %q, want %q", text, expected)
	}
}
