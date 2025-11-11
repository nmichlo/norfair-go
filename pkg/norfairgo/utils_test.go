package norfairgo

import (
	"testing"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// ValidatePoints Tests
// =============================================================================

func TestValidatePoints_Valid2D(t *testing.T) {
	// Test valid 2D points array (n_points, 2)
	points := mat.NewDense(3, 2, []float64{
		1.0, 2.0,
		3.0, 4.0,
		5.0, 6.0,
	})

	validated, err := ValidatePoints(points)
	if err != nil {
		t.Fatalf("Expected no error for valid 2D points, got: %v", err)
	}

	rows, cols := validated.Dims()
	if rows != 3 || cols != 2 {
		t.Errorf("Expected shape (3, 2), got (%d, %d)", rows, cols)
	}
}

func TestValidatePoints_Valid3D(t *testing.T) {
	// Test valid 3D points array (n_points, 3)
	points := mat.NewDense(2, 3, []float64{
		1.0, 2.0, 3.0,
		4.0, 5.0, 6.0,
	})

	validated, err := ValidatePoints(points)
	if err != nil {
		t.Fatalf("Expected no error for valid 3D points, got: %v", err)
	}

	rows, cols := validated.Dims()
	if rows != 2 || cols != 3 {
		t.Errorf("Expected shape (2, 3), got (%d, %d)", rows, cols)
	}
}

func TestValidatePoints_Single2DPoint(t *testing.T) {
	// Test single 2D point (1, 2) - edge case
	points := mat.NewDense(1, 2, []float64{10.0, 20.0})

	validated, err := ValidatePoints(points)
	if err != nil {
		t.Fatalf("Expected no error for single 2D point, got: %v", err)
	}

	rows, cols := validated.Dims()
	if rows != 1 || cols != 2 {
		t.Errorf("Expected shape (1, 2), got (%d, %d)", rows, cols)
	}

	// Verify values preserved
	if validated.At(0, 0) != 10.0 || validated.At(0, 1) != 20.0 {
		t.Errorf("Expected values (10, 20), got (%.1f, %.1f)",
			validated.At(0, 0), validated.At(0, 1))
	}
}

func TestValidatePoints_Single3DPoint(t *testing.T) {
	// Test single 3D point (1, 3) - edge case
	points := mat.NewDense(1, 3, []float64{10.0, 20.0, 30.0})

	validated, err := ValidatePoints(points)
	if err != nil {
		t.Fatalf("Expected no error for single 3D point, got: %v", err)
	}

	rows, cols := validated.Dims()
	if rows != 1 || cols != 3 {
		t.Errorf("Expected shape (1, 3), got (%d, %d)", rows, cols)
	}

	// Verify values preserved
	if validated.At(0, 0) != 10.0 || validated.At(0, 1) != 20.0 || validated.At(0, 2) != 30.0 {
		t.Errorf("Expected values (10, 20, 30), got (%.1f, %.1f, %.1f)",
			validated.At(0, 0), validated.At(0, 1), validated.At(0, 2))
	}
}

func TestValidatePoints_InvalidDimensions4D(t *testing.T) {
	// Test invalid dimensions (n, 4) - should error
	points := mat.NewDense(2, 4, []float64{
		1.0, 2.0, 3.0, 4.0,
		5.0, 6.0, 7.0, 8.0,
	})

	_, err := ValidatePoints(points)
	if err == nil {
		t.Fatal("Expected error for 4D points, got nil")
	}

	// Verify error message mentions invalid shape
	expectedMsg := "invalid points shape"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidatePoints_InvalidDimensions1D(t *testing.T) {
	// Test invalid dimensions (n, 1) - should error
	points := mat.NewDense(3, 1, []float64{1.0, 2.0, 3.0})

	_, err := ValidatePoints(points)
	if err == nil {
		t.Fatal("Expected error for 1D points (n, 1), got nil")
	}

	// Verify error message mentions invalid shape
	expectedMsg := "invalid points shape"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidatePoints_InvalidSingleValue(t *testing.T) {
	// Test single value (1, 1) - should error (neither 2D nor 3D)
	points := mat.NewDense(1, 1, []float64{10.0})

	_, err := ValidatePoints(points)
	if err == nil {
		t.Fatal("Expected error for single value (1, 1), got nil")
	}

	// Verify error message mentions invalid shape
	expectedMsg := "invalid points shape"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

// =============================================================================
// GetTerminalSize Tests
// =============================================================================

func TestGetTerminalSize_ReturnsValues(t *testing.T) {
	// Test that GetTerminalSize returns values (either detected or defaults)
	cols, lines := GetTerminalSize(80, 24)

	// Should return positive values
	if cols <= 0 {
		t.Errorf("Expected positive cols, got %d", cols)
	}

	if lines <= 0 {
		t.Errorf("Expected positive lines, got %d", lines)
	}

	// In most test environments, this will return the defaults since
	// tests don't run in a real terminal
	// But the function should at least return valid values
	t.Logf("Terminal size: %d cols x %d lines", cols, lines)
}

func TestGetTerminalSize_CustomDefaults(t *testing.T) {
	// Test with custom default values
	cols, lines := GetTerminalSize(100, 50)

	// Should return positive values
	if cols <= 0 {
		t.Errorf("Expected positive cols, got %d", cols)
	}

	if lines <= 0 {
		t.Errorf("Expected positive lines, got %d", lines)
	}

	t.Logf("Terminal size with custom defaults: %d cols x %d lines", cols, lines)
}

func TestGetTerminalSize_StandardDefaults(t *testing.T) {
	// Test that standard defaults work (80x24 is conventional)
	cols, lines := GetTerminalSize(80, 24)

	// Verify values are reasonable
	if cols < 80 && lines < 24 {
		t.Errorf("Terminal size seems too small: %dx%d", cols, lines)
	}

	t.Logf("Terminal size: %d cols x %d lines", cols, lines)
}

// =============================================================================
// GetCutout Tests
// =============================================================================

func TestGetCutout_CenterRegion(t *testing.T) {
	// Create a 100x100 test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Set entire image to black
	img.SetTo(gocv.NewScalar(0, 0, 0, 0))

	// Define points in center region (20,20) to (80,80)
	points := mat.NewDense(4, 2, []float64{
		20.0, 20.0,
		80.0, 20.0,
		80.0, 80.0,
		20.0, 80.0,
	})

	// Extract cutout
	cutout := GetCutout(points, img)
	defer cutout.Close()

	// Verify cutout dimensions
	// Region should be from (20,20) to (81,81) = 61x61
	if cutout.Rows() != 61 || cutout.Cols() != 61 {
		t.Errorf("Expected cutout size 61x61, got %dx%d", cutout.Rows(), cutout.Cols())
	}
}

func TestGetCutout_CornerRegion(t *testing.T) {
	// Create a 100x100 test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Define points in top-left corner (0,0) to (30,30)
	points := mat.NewDense(2, 2, []float64{
		0.0, 0.0,
		30.0, 30.0,
	})

	// Extract cutout
	cutout := GetCutout(points, img)
	defer cutout.Close()

	// Verify cutout dimensions
	// Region should be from (0,0) to (31,31) = 31x31
	if cutout.Rows() != 31 || cutout.Cols() != 31 {
		t.Errorf("Expected cutout size 31x31, got %dx%d", cutout.Rows(), cutout.Cols())
	}
}

func TestGetCutout_SinglePoint(t *testing.T) {
	// Create a 100x100 test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Define single point at (50,50)
	points := mat.NewDense(1, 2, []float64{50.0, 50.0})

	// Extract cutout
	cutout := GetCutout(points, img)
	defer cutout.Close()

	// Single point should create 1x1 region from (50,50) to (51,51)
	if cutout.Rows() != 1 || cutout.Cols() != 1 {
		t.Errorf("Expected cutout size 1x1 for single point, got %dx%d", cutout.Rows(), cutout.Cols())
	}
}

func TestGetCutout_OutOfBounds(t *testing.T) {
	// Create a 100x100 test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Define points that extend beyond image bounds
	points := mat.NewDense(4, 2, []float64{
		-10.0, -10.0, // Out of bounds (negative)
		50.0, 50.0, // In bounds
		150.0, 50.0, // Out of bounds (too large)
		50.0, 150.0, // Out of bounds (too large)
	})

	// Extract cutout
	cutout := GetCutout(points, img)
	defer cutout.Close()

	// Region should be clamped to (0,0) to (100,100) = full image
	if cutout.Rows() != 100 || cutout.Cols() != 100 {
		t.Errorf("Expected cutout size 100x100 (clamped), got %dx%d", cutout.Rows(), cutout.Cols())
	}
}

func TestGetCutout_LargeRegion(t *testing.T) {
	// Create a 200x200 test image
	img := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Define points covering most of the image
	points := mat.NewDense(4, 2, []float64{
		10.0, 10.0,
		190.0, 10.0,
		190.0, 190.0,
		10.0, 190.0,
	})

	// Extract cutout
	cutout := GetCutout(points, img)
	defer cutout.Close()

	// Verify large cutout dimensions
	// Region should be from (10,10) to (191,191) = 181x181
	if cutout.Rows() != 181 || cutout.Cols() != 181 {
		t.Errorf("Expected cutout size 181x181, got %dx%d", cutout.Rows(), cutout.Cols())
	}
}

func TestGetCutout_InvalidPoints(t *testing.T) {
	// Create a 100x100 test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Points with only 1 column (invalid)
	points := mat.NewDense(3, 1, []float64{10.0, 20.0, 30.0})

	// Extract cutout
	cutout := GetCutout(points, img)
	defer cutout.Close()

	// Should return empty mat
	if !cutout.Empty() {
		t.Error("Expected empty mat for invalid points")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
