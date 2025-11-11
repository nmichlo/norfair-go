package norfairgodraw

import (
	"image"
	"math"
	"testing"

	"gocv.io/x/gocv"
)

// Test 1: Grid generation (equator mode)
func TestGetGrid_EquatorMode(t *testing.T) {
	ClearGridCache()

	size := 20
	w, h := 640, 480
	polar := false

	grid := getGrid(size, w, h, polar)

	rows, cols := grid.Dims()

	// Should have 2 columns (x, y)
	if cols != 2 {
		t.Errorf("Expected 2 columns, got %d", cols)
	}

	// Should have size*size points (approximately)
	step := math.Pi / float64(size)
	start := -math.Pi/2 + step/2
	end := math.Pi / 2
	numTheta := int(math.Ceil((end - start) / step))
	expectedPoints := numTheta * numTheta

	if rows != expectedPoints {
		t.Errorf("Expected %d points, got %d", expectedPoints, rows)
	}

	// Check that points are scaled and centered
	maxDim := float64(maxInt(h, w))
	centerX := float64(w / 2)
	centerY := float64(h / 2)

	// At least some points should be near center
	foundCenterPoint := false
	for i := 0; i < rows; i++ {
		x := grid.At(i, 0)
		y := grid.At(i, 1)

		// Check if point is reasonably close to center
		if math.Abs(x-centerX) < maxDim*0.1 && math.Abs(y-centerY) < maxDim*0.1 {
			foundCenterPoint = true
			break
		}
	}

	if !foundCenterPoint {
		t.Error("Expected at least one point near center")
	}
}

// Test 2: Grid generation (polar mode)
func TestGetGrid_PolarMode(t *testing.T) {
	ClearGridCache()

	size := 20
	w, h := 640, 480
	polar := true

	grid := getGrid(size, w, h, polar)

	rows, cols := grid.Dims()

	// Should have 2 columns
	if cols != 2 {
		t.Errorf("Expected 2 columns, got %d", cols)
	}

	// Should have points
	if rows == 0 {
		t.Error("Expected non-zero points")
	}
}

// Test 3: Grid caching
func TestGetGrid_Caching(t *testing.T) {
	ClearGridCache()

	size := 20
	w, h := 640, 480
	polar := false

	// First call
	grid1 := getGrid(size, w, h, polar)

	// Check cache size
	if GridCacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", GridCacheSize())
	}

	// Second call with same parameters
	grid2 := getGrid(size, w, h, polar)

	// Should return same grid (pointer equality)
	if grid1 != grid2 {
		t.Error("Expected cached grid to be returned")
	}

	// Cache size should still be 1
	if GridCacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", GridCacheSize())
	}

	// Different parameters
	grid3 := getGrid(size, w, h, true) // polar=true

	// Cache size should be 2
	if GridCacheSize() != 2 {
		t.Errorf("Expected cache size 2, got %d", GridCacheSize())
	}

	// Should be different grid
	if grid1 == grid3 {
		t.Error("Expected different grid for different parameters")
	}
}

// Test 4: Cache eviction (maxsize=4)
func TestGetGrid_CacheEviction(t *testing.T) {
	ClearGridCache()

	// Add 5 different grids (maxsize=4)
	for i := 0; i < 5; i++ {
		size := 10 + i*2
		getGrid(size, 640, 480, false)
	}

	// Cache should have been cleared at some point
	cacheSize := GridCacheSize()
	if cacheSize > 4 {
		t.Errorf("Cache size should not exceed 4, got %d", cacheSize)
	}
}

// Test 5: DrawAbsoluteGrid with nil transform
func TestDrawAbsoluteGrid_NilTransform(t *testing.T) {
	ClearGridCache()

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw grid with nil transform (identity)
	DrawAbsoluteGrid(&frame, nil, 10, 2, 1, &Color{B: 0, G: 0, R: 0}, false)

	// Should not crash, frame should still be valid
	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// Test 6: DrawAbsoluteGrid with transform
func TestDrawAbsoluteGrid_WithTransform(t *testing.T) {
	ClearGridCache()

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Mock transform (10 pixel offset)
	transform := &mockFixedCameraTransform{offset: image.Point{X: 10, Y: 10}}

	// Draw grid with transform
	DrawAbsoluteGrid(&frame, transform, 10, 2, 1, &Color{B: 0, G: 0, R: 0}, false)

	// Should not crash
	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// Test 7: DrawAbsoluteGrid with nil frame
func TestDrawAbsoluteGrid_NilFrame(t *testing.T) {
	// Should not crash with nil frame
	DrawAbsoluteGrid(nil, nil, 10, 2, 1, &Color{B: 0, G: 0, R: 0}, false)
}

// Test 8: DrawAbsoluteGrid with defaults
func TestDrawAbsoluteGrid_Defaults(t *testing.T) {
	ClearGridCache()

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw with zero/negative values to test defaults
	DrawAbsoluteGrid(&frame, nil, 0, 0, 0, nil, false)

	// Should use defaults and not crash
	if frame.Empty() {
		t.Error("Frame should not be empty")
	}
}

// Test 9: DrawAbsoluteGrid polar mode
func TestDrawAbsoluteGrid_PolarMode(t *testing.T) {
	ClearGridCache()

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw with polar=true
	DrawAbsoluteGrid(&frame, nil, 10, 2, 1, &Color{B: 255, G: 0, R: 0}, true)

	// Should not crash
	if frame.Empty() {
		t.Error("Frame should not be empty")
	}
}

// Test 10: DrawAbsoluteGridWithDefaults
func TestDrawAbsoluteGridWithDefaults_Function(t *testing.T) {
	ClearGridCache()

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Convenience function
	DrawAbsoluteGridWithDefaults(&frame, nil)

	// Should not crash
	if frame.Empty() {
		t.Error("Frame should not be empty")
	}
}

// Test 11: Grid point visibility filtering
func TestDrawAbsoluteGrid_VisibilityFiltering(t *testing.T) {
	ClearGridCache()

	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Transform that moves points outside frame
	transform := &mockFixedCameraTransform{offset: image.Point{X: 1000, Y: 1000}}

	// Should only draw visible points (none in this case)
	DrawAbsoluteGrid(&frame, transform, 10, 2, 1, &Color{B: 0, G: 0, R: 0}, false)

	// Should not crash
	if frame.Empty() {
		t.Error("Frame should not be empty")
	}
}

// Test 12: Grid cache key generation
func TestGridCacheKey_Format(t *testing.T) {
	key := GridCacheKey(20, 640, 480, false)
	expected := "size=20,w=640,h=480,polar=false"

	if key != expected {
		t.Errorf("Expected key '%s', got '%s'", expected, key)
	}

	key2 := GridCacheKey(20, 640, 480, true)
	expected2 := "size=20,w=640,h=480,polar=true"

	if key2 != expected2 {
		t.Errorf("Expected key '%s', got '%s'", expected2, key2)
	}
}

// Test 13: computeGrid coordinates (equator)
func TestComputeGrid_EquatorCoordinates(t *testing.T) {
	size := 10
	w, h := 100, 100

	grid := computeGrid(size, w, h, false)

	rows, cols := grid.Dims()

	// Check that we have points
	if rows == 0 {
		t.Error("Expected non-zero points")
	}

	// Check dimensions
	if cols != 2 {
		t.Errorf("Expected 2 columns (x, y), got %d", cols)
	}

	// Check that points are valid numbers (not NaN or Inf)
	hasValidPoint := false
	for i := 0; i < rows; i++ {
		x := grid.At(i, 0)
		y := grid.At(i, 1)

		if !math.IsNaN(x) && !math.IsNaN(y) && !math.IsInf(x, 0) && !math.IsInf(y, 0) {
			hasValidPoint = true
			break
		}
	}

	if !hasValidPoint {
		t.Error("Expected at least one valid (non-NaN, non-Inf) point")
	}
}

// Test 14: computeGrid coordinates (polar)
func TestComputeGrid_PolarCoordinates(t *testing.T) {
	size := 10
	w, h := 100, 100

	grid := computeGrid(size, w, h, true)

	rows, _ := grid.Dims()

	// Check that points are within reasonable range (polar mode)
	maxDim := float64(maxInt(h, w))

	pointsInRange := 0
	for i := 0; i < rows; i++ {
		x := grid.At(i, 0)
		y := grid.At(i, 1)

		// Points should be within reasonable distance from center
		if x >= -maxDim && x <= float64(w)+maxDim &&
			y >= -maxDim && y <= float64(h)+maxDim {
			pointsInRange++
		}
	}

	// At least most points should be in range
	if pointsInRange < rows/2 {
		t.Errorf("Expected at least half of points in range, got %d/%d", pointsInRange, rows)
	}
}

// Test 15: Large grid size
func TestGetGrid_LargeSize(t *testing.T) {
	ClearGridCache()

	size := 100
	w, h := 1920, 1080

	grid := getGrid(size, w, h, false)

	rows, cols := grid.Dims()

	if cols != 2 {
		t.Errorf("Expected 2 columns, got %d", cols)
	}

	if rows == 0 {
		t.Error("Expected non-zero points for large grid")
	}

	// Should have many points
	if rows < 100 {
		t.Errorf("Expected at least 100 points for large grid, got %d", rows)
	}
}
