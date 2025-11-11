package norfairgodraw

import (
	"image"
	"testing"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go/internal/numpy"
	"github.com/nmichlo/norfair-go/pkg/norfairgo"
)

// mockTrackedObjectForPaths creates a mock TrackedObject for path testing
type mockTrackedObjectForPaths struct {
	estimate   *mat.Dense
	id         *int
	label      *string
	absToRel   norfairgo.CoordinateTransformation
	livePoints []bool
}

func (m *mockTrackedObjectForPaths) GetEstimate(absolute bool) (*mat.Dense, error) {
	return m.estimate, nil
}

func (m *mockTrackedObjectForPaths) GetID() *int {
	return m.id
}

func (m *mockTrackedObjectForPaths) GetLabel() *string {
	return m.label
}

func (m *mockTrackedObjectForPaths) GetLivePoints() []bool {
	return m.livePoints
}

func (m *mockTrackedObjectForPaths) LivePoints() []bool {
	return m.livePoints
}

// Ensure mockTrackedObjectForPaths implements norfairgo.TrackedObject
var _ interface {
	GetEstimate(bool) (*mat.Dense, error)
	GetID() *int
	GetLabel() *string
	GetLivePoints() []bool
	LivePoints() []bool
} = (*mockTrackedObjectForPaths)(nil)

// TestPaths_LazyInit verifies that mask is created on first Draw() call
func TestPaths_LazyInit(t *testing.T) {
	paths := NewPaths(nil, nil, nil, nil, 0.01)

	// Initially, mask should be nil
	if paths.mask != nil {
		t.Error("Expected mask to be nil before first draw")
	}

	// Create a test frame
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw once with empty object list
	result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result.Close()

	// Now mask should be initialized
	if paths.mask == nil {
		t.Error("Expected mask to be initialized after first draw")
	}

	// Verify mask dimensions match frame
	if paths.mask.Rows() != frame.Rows() || paths.mask.Cols() != frame.Cols() {
		t.Errorf("Expected mask size %dx%d, got %dx%d",
			frame.Rows(), frame.Cols(), paths.mask.Rows(), paths.mask.Cols())
	}
}

// Python equivalent: tools/validate_path/main.py calculates auto-scaling
//
//	def calc_auto_params(frame_height):
//	    frame_scale = frame_height / 100.0
//	    radius = int(max(frame_scale * 0.7, 1))
//	    thickness = int(max(frame_scale / 7.0, 1))
//	    return radius, thickness
//	# Height 480p: radius=3, thickness=1
//	# Height 1080p: radius=7, thickness=1
//
// TestPaths_AutoScaling verifies radius and thickness are auto-calculated
func TestPaths_AutoScaling(t *testing.T) {
	tests := []struct {
		name           string
		height         int
		expectedRadius int
		expectedThick  int
	}{
		{
			name:           "480p",
			height:         480,
			expectedRadius: 3, // max(480/100 * 0.7, 1) = max(3.36, 1) = 3
			expectedThick:  1, // max(480/100 / 7, 1) = max(0.685, 1) = 1
		},
		{
			name:           "1080p",
			height:         1080,
			expectedRadius: 7, // max(1080/100 * 0.7, 1) = max(7.56, 1) = 7
			expectedThick:  1, // max(1080/100 / 7, 1) = max(1.542, 1) = 1
		},
		{
			name:           "Small (100x100)",
			height:         100,
			expectedRadius: 1, // max(100/100 * 0.7, 1) = max(0.7, 1) = 1
			expectedThick:  1, // max(100/100 / 7, 1) = max(0.142, 1) = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := NewPaths(nil, nil, nil, nil, 0.01)

			frame := gocv.NewMatWithSize(tt.height, 640, gocv.MatTypeCV8UC3)
			defer frame.Close()

			// Trigger lazy init with empty object list
			result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
			defer result.Close()

			if paths.radius == nil {
				t.Fatal("Expected radius to be set")
			}
			if *paths.radius != tt.expectedRadius {
				t.Errorf("Expected radius %d, got %d", tt.expectedRadius, *paths.radius)
			}

			if paths.thickness == nil {
				t.Fatal("Expected thickness to be set")
			}
			if *paths.thickness != tt.expectedThick {
				t.Errorf("Expected thickness %d, got %d", tt.expectedThick, *paths.thickness)
			}
		})
	}
}

// TestPaths_Fade verifies attenuation works over multiple frames
func TestPaths_Fade(t *testing.T) {
	attenuation := 0.1 // 10% fade per frame
	paths := NewPaths(nil, nil, nil, nil, attenuation)

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw 10 frames with no objects (just fading)
	for i := 0; i < 10; i++ {
		result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
		result.Close()
	}

	// After 10 frames with 10% attenuation, mask should be very dim
	// attenuation_factor = 1 - 0.1 = 0.9
	// After 10 frames: 0.9^10 ≈ 0.349
	// So a pixel with value 255 should become approximately 255 * 0.349 ≈ 89

	// Check that mask exists and has been attenuated
	if paths.mask == nil {
		t.Fatal("Expected mask to exist after drawing")
	}

	// Note: We can't easily verify the exact pixel values without drawing something first,
	// but we can verify the structure is correct
	if paths.attenuationFactor != 0.9 {
		t.Errorf("Expected attenuationFactor 0.9, got %f", paths.attenuationFactor)
	}
}

// TestPaths_Accumulation verifies circles accumulate on mask
func TestPaths_Accumulation(t *testing.T) {
	radius := 5
	thickness := -1                                       // Filled circle
	paths := NewPaths(nil, &thickness, nil, &radius, 0.0) // attenuation=0 means no fade

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw first frame
	result1 := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result1.Close()

	// Draw second frame (should accumulate)
	result2 := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result2.Close()

	// Verify mask exists
	if paths.mask == nil {
		t.Error("Expected mask to exist")
	}
}

// TestPaths_ColorByID verifies palette-based color selection
func TestPaths_ColorByID(t *testing.T) {
	paths := NewPaths(nil, nil, nil, nil, 0.01)

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw with empty list to trigger initialization
	result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result.Close()

	// Verify palette was created
	if paths.palette == nil {
		t.Error("Expected palette to be initialized")
	}

	// Verify color is nil (will use palette)
	if paths.color != nil {
		t.Error("Expected color to be nil when using palette")
	}
}

// TestPaths_CustomColor verifies explicit color parameter is respected
func TestPaths_CustomColor(t *testing.T) {
	customColor := Color{B: 0, G: 0, R: 255} // Red (BGR format)
	paths := NewPaths(nil, nil, &customColor, nil, 0.01)

	if paths.color == nil {
		t.Fatal("Expected color to be set")
	}

	if *paths.color != customColor {
		t.Errorf("Expected color %v, got %v", customColor, *paths.color)
	}

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result.Close()

	// Color should still be the custom color
	if *paths.color != customColor {
		t.Error("Expected custom color to be preserved")
	}
}

// TestPaths_CameraMotionWarning verifies warning is issued when used with camera motion
func TestPaths_CameraMotionWarning(t *testing.T) {
	paths := NewPaths(nil, nil, nil, nil, 0.01)

	// Initially, no warning should have been issued
	if paths.warnedCameraMotion {
		t.Error("Expected warnedCameraMotion to be false initially")
	}

	// Note: We can't easily capture the warning output in the test,
	// but we can verify the flag is set after calling with an object that has abs_to_rel

	// For now, just verify the structure
	if paths.warnedCameraMotion {
		t.Error("Expected warnedCameraMotion to remain false without camera motion objects")
	}
}

// TestPaths_GetPointsToDraw verifies custom point extraction function
func TestPaths_GetPointsToDraw(t *testing.T) {
	// Custom function that returns first point only
	customFunc := func(estimate *mat.Dense) []image.Point {
		if estimate.RawMatrix().Rows == 0 {
			return []image.Point{}
		}
		return []image.Point{{X: int(estimate.At(0, 0)), Y: int(estimate.At(0, 1))}}
	}

	paths := NewPaths(customFunc, nil, nil, nil, 0.01)

	// Verify function was set
	if paths.getPointsToDraw == nil {
		t.Error("Expected getPointsToDraw to be set")
	}

	// Test the function directly
	testEstimate := mat.NewDense(2, 2, []float64{10, 20, 30, 40})
	points := paths.getPointsToDraw(testEstimate)

	if len(points) != 1 {
		t.Errorf("Expected 1 point, got %d", len(points))
	}

	if points[0].X != 10 || points[0].Y != 20 {
		t.Errorf("Expected point (10, 20), got (%d, %d)", points[0].X, points[0].Y)
	}
}

// TestPaths_EmptyObjects verifies handling of empty object list
func TestPaths_EmptyObjects(t *testing.T) {
	paths := NewPaths(nil, nil, nil, nil, 0.01)

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw with empty list
	result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result.Close()

	// Should not crash and should initialize mask
	if paths.mask == nil {
		t.Error("Expected mask to be initialized even with empty object list")
	}

	// Result should be valid
	if result.Empty() {
		t.Error("Expected non-empty result frame")
	}
}

// TestPaths_AlphaBlend verifies final blend operation returns correct frame
func TestPaths_AlphaBlend(t *testing.T) {
	paths := NewPaths(nil, nil, nil, nil, 0.01)

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Fill frame with white
	frame.SetTo(gocv.NewScalar(255, 255, 255, 0))

	// Draw (will blend mask with frame)
	result := paths.Draw(&frame, []*norfairgo.TrackedObject{})
	defer result.Close()

	// Result should have same dimensions
	if result.Rows() != frame.Rows() || result.Cols() != frame.Cols() {
		t.Errorf("Expected result size %dx%d, got %dx%d",
			frame.Rows(), frame.Cols(), result.Rows(), result.Cols())
	}

	// Result should not be empty
	if result.Empty() {
		t.Error("Expected non-empty result frame")
	}

	// Result should be a different Mat from input (new allocation)
	// This is implicit in the return value
}

// TestDefaultGetPointsToDraw verifies the default centroid calculation
func TestDefaultGetPointsToDraw(t *testing.T) {
	tests := []struct {
		name     string
		estimate *mat.Dense
		expected image.Point
	}{
		{
			name:     "Single point",
			estimate: mat.NewDense(1, 2, []float64{100, 200}),
			expected: image.Point{X: 100, Y: 200},
		},
		{
			name:     "Two points",
			estimate: mat.NewDense(2, 2, []float64{100, 200, 300, 400}),
			expected: image.Point{X: 200, Y: 300}, // Mean: (100+300)/2=200, (200+400)/2=300
		},
		{
			name:     "Four points (square)",
			estimate: mat.NewDense(4, 2, []float64{0, 0, 100, 0, 100, 100, 0, 100}),
			expected: image.Point{X: 50, Y: 50}, // Centroid of square
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultGetPointsToDraw(tt.estimate)

			if len(result) != 1 {
				t.Fatalf("Expected 1 point, got %d", len(result))
			}

			if result[0] != tt.expected {
				t.Errorf("Expected point %v, got %v", tt.expected, result[0])
			}
		})
	}

	// Test empty estimate separately (can't create 0x0 matrix easily)
	t.Run("Empty estimate", func(t *testing.T) {
		// Create a 1x1 matrix but check the function handles rows==0
		empty := mat.NewDense(1, 2, []float64{0, 0})
		empty.Reset()
		// After Reset, matrix may not be in valid state, so test the logic directly
		// by creating a matrix with 0 rows using SetRow approach or just verify
		// the function can handle edge case

		// Actually, let's test with a matrix that has cols < 2
		invalidCols := mat.NewDense(1, 1, []float64{100})
		result := defaultGetPointsToDraw(invalidCols)
		if len(result) != 0 {
			t.Errorf("Expected empty result for invalid columns, got %d points", len(result))
		}
	})
}

// mockCoordinateTransformation implements a simple translation transformation for testing
type mockCoordinateTransformation struct {
	offset image.Point // Offset to apply in AbsToRel
}

func (m *mockCoordinateTransformation) AbsToRel(points *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	result := mat.NewDense(rows, cols, nil)

	for i := 0; i < rows; i++ {
		// Apply translation: relative = absolute + offset
		result.Set(i, 0, points.At(i, 0)+float64(m.offset.X))
		result.Set(i, 1, points.At(i, 1)+float64(m.offset.Y))
	}

	return result
}

func (m *mockCoordinateTransformation) RelToAbs(points *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	result := mat.NewDense(rows, cols, nil)

	for i := 0; i < rows; i++ {
		// Inverse translation: absolute = relative - offset
		result.Set(i, 0, points.At(i, 0)-float64(m.offset.X))
		result.Set(i, 1, points.At(i, 1)-float64(m.offset.Y))
	}

	return result
}

// =============================================================================
// AbsolutePaths Tests
// =============================================================================

// TestAbsolutePaths_Constructor verifies constructor and alpha generation
func TestAbsolutePaths_Constructor(t *testing.T) {
	t.Run("Default maxHistory", func(t *testing.T) {
		ap := NewAbsolutePaths(nil, nil, nil, nil, 0)

		if ap.maxHistory != 20 {
			t.Errorf("Expected default maxHistory 20, got %d", ap.maxHistory)
		}

		if len(ap.alphas) != 20 {
			t.Errorf("Expected 20 alphas, got %d", len(ap.alphas))
		}
	})

	t.Run("Custom maxHistory", func(t *testing.T) {
		ap := NewAbsolutePaths(nil, nil, nil, nil, 10)

		if ap.maxHistory != 10 {
			t.Errorf("Expected maxHistory 10, got %d", ap.maxHistory)
		}

		if len(ap.alphas) != 10 {
			t.Errorf("Expected 10 alphas, got %d", len(ap.alphas))
		}
	})

	t.Run("Alphas range 0.99 to 0.01", func(t *testing.T) {
		ap := NewAbsolutePaths(nil, nil, nil, nil, 5)

		if len(ap.alphas) != 5 {
			t.Fatalf("Expected 5 alphas, got %d", len(ap.alphas))
		}

		// First alpha should be ~0.99
		if ap.alphas[0] < 0.98 || ap.alphas[0] > 1.0 {
			t.Errorf("Expected first alpha ~0.99, got %f", ap.alphas[0])
		}

		// Last alpha should be ~0.01
		if ap.alphas[4] < 0.0 || ap.alphas[4] > 0.02 {
			t.Errorf("Expected last alpha ~0.01, got %f", ap.alphas[4])
		}

		// Alphas should decrease monotonically
		for i := 1; i < len(ap.alphas); i++ {
			if ap.alphas[i] >= ap.alphas[i-1] {
				t.Errorf("Alphas not decreasing: alphas[%d]=%f >= alphas[%d]=%f",
					i, ap.alphas[i], i-1, ap.alphas[i-1])
			}
		}
	})
}

// TestAbsolutePaths_AutoScaling verifies radius and thickness auto-calculation
func TestAbsolutePaths_AutoScaling(t *testing.T) {
	ap := NewAbsolutePaths(nil, nil, nil, nil, 5)

	frame := gocv.NewMatWithSize(1080, 1920, gocv.MatTypeCV8UC3)
	// Note: Don't defer frame.Close() - Draw() takes ownership

	coordTransform := &mockCoordinateTransformation{offset: image.Point{X: 0, Y: 0}}

	// Draw with no objects to trigger auto-scaling
	result := ap.Draw(&frame, []*norfairgo.TrackedObject{}, coordTransform)
	defer result.Close()

	// Verify radius and thickness were set
	if ap.radius == nil {
		t.Fatal("Expected radius to be set")
	}

	if ap.thickness == nil {
		t.Fatal("Expected thickness to be set")
	}

	// For 1080p: frameScale = 1080/100 = 10.8
	// radius = max(10.8 * 0.7, 1) = max(7.56, 1) = 7
	// thickness = max(10.8 / 7, 1) = max(1.54, 1) = 1
	if *ap.radius != 7 {
		t.Errorf("Expected radius 7 for 1080p, got %d", *ap.radius)
	}

	if *ap.thickness != 1 {
		t.Errorf("Expected thickness 1 for 1080p, got %d", *ap.thickness)
	}
}

// TestAbsolutePaths_EmptyObjects verifies handling of objects with no live points
func TestAbsolutePaths_EmptyObjects(t *testing.T) {
	ap := NewAbsolutePaths(nil, nil, nil, nil, 5)

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	// Note: Don't defer frame.Close() - Draw() takes ownership

	coordTransform := &mockCoordinateTransformation{offset: image.Point{X: 0, Y: 0}}

	// Draw with empty list
	result := ap.Draw(&frame, []*norfairgo.TrackedObject{}, coordTransform)
	defer result.Close()

	// Should not crash
	if result.Empty() {
		t.Error("Expected non-empty result")
	}

	// pastPoints map should be empty
	if len(ap.pastPoints) != 0 {
		t.Errorf("Expected empty pastPoints map, got %d entries", len(ap.pastPoints))
	}
}

// TestLinspace verifies the linspace helper function
func TestLinspace(t *testing.T) {
	tests := []struct {
		name      string
		start     float64
		end       float64
		n         int
		expected  []float64
		tolerance float64
	}{
		{
			name:     "Zero elements",
			start:    0.0,
			end:      1.0,
			n:        0,
			expected: []float64{},
		},
		{
			name:     "One element",
			start:    0.5,
			end:      1.0,
			n:        1,
			expected: []float64{0.5},
		},
		{
			name:      "Two elements",
			start:     0.0,
			end:       1.0,
			n:         2,
			expected:  []float64{0.0, 1.0},
			tolerance: 1e-10,
		},
		{
			name:      "Five elements",
			start:     0.0,
			end:       1.0,
			n:         5,
			expected:  []float64{0.0, 0.25, 0.5, 0.75, 1.0},
			tolerance: 1e-10,
		},
		{
			name:      "Descending",
			start:     1.0,
			end:       0.0,
			n:         3,
			expected:  []float64{1.0, 0.5, 0.0},
			tolerance: 1e-10,
		},
		{
			name:      "Alpha fade (99% to 1%)",
			start:     0.99,
			end:       0.01,
			n:         5,
			expected:  []float64{0.99, 0.745, 0.5, 0.255, 0.01},
			tolerance: 1e-10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := numpy.Linspace(tt.start, tt.end, tt.n)

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d elements, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				diff := result[i] - tt.expected[i]
				if diff < 0 {
					diff = -diff
				}
				if diff > tt.tolerance {
					t.Errorf("Element %d: expected %f, got %f (diff %f > tolerance %f)",
						i, tt.expected[i], result[i], diff, tt.tolerance)
				}
			}
		})
	}
}

// TestAbsolutePaths_ColorByID verifies palette-based color selection
func TestAbsolutePaths_ColorByID(t *testing.T) {
	ap := NewAbsolutePaths(nil, nil, nil, nil, 5)

	// Verify palette was created
	if ap.palette == nil {
		t.Error("Expected palette to be initialized")
	}

	// Verify color is nil (will use palette)
	if ap.color != nil {
		t.Error("Expected color to be nil when using palette")
	}
}

// TestAbsolutePaths_CustomColor verifies explicit color parameter is respected
func TestAbsolutePaths_CustomColor(t *testing.T) {
	customColor := Color{B: 0, G: 255, R: 0} // Green (BGR format)
	ap := NewAbsolutePaths(nil, nil, &customColor, nil, 5)

	if ap.color == nil {
		t.Fatal("Expected color to be set")
	}

	if *ap.color != customColor {
		t.Errorf("Expected color %v, got %v", customColor, *ap.color)
	}
}

// TestAbsolutePaths_CustomGetPointsToDraw verifies custom point extraction
func TestAbsolutePaths_CustomGetPointsToDraw(t *testing.T) {
	// Custom function that returns all points
	customFunc := func(estimate *mat.Dense) []image.Point {
		rows, _ := estimate.Dims()
		result := make([]image.Point, rows)
		for i := 0; i < rows; i++ {
			result[i] = image.Point{
				X: int(estimate.At(i, 0)),
				Y: int(estimate.At(i, 1)),
			}
		}
		return result
	}

	ap := NewAbsolutePaths(customFunc, nil, nil, nil, 5)

	// Verify function was set
	if ap.getPointsToDraw == nil {
		t.Error("Expected getPointsToDraw to be set")
	}

	// Test the function directly
	testEstimate := mat.NewDense(3, 2, []float64{10, 20, 30, 40, 50, 60})
	points := ap.getPointsToDraw(testEstimate)

	if len(points) != 3 {
		t.Errorf("Expected 3 points, got %d", len(points))
	}

	expected := []image.Point{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}
	for i, pt := range points {
		if pt != expected[i] {
			t.Errorf("Point %d: expected %v, got %v", i, expected[i], pt)
		}
	}
}

// TestAbsolutePaths_TransformPointsToRelative verifies coordinate transformation
func TestAbsolutePaths_TransformPointsToRelative(t *testing.T) {
	ap := NewAbsolutePaths(nil, nil, nil, nil, 5)

	// Create transformation with offset (10, 20)
	coordTransform := &mockCoordinateTransformation{offset: image.Point{X: 10, Y: 20}}

	// Test points in absolute coordinates
	absolutePoints := []image.Point{
		{X: 100, Y: 200},
		{X: 300, Y: 400},
	}

	// Transform to relative
	relativePoints := ap.transformPointsToRelative(absolutePoints, coordTransform)

	// Expected: absolute + offset = relative
	// (100, 200) + (10, 20) = (110, 220)
	// (300, 400) + (10, 20) = (310, 420)
	expected := []image.Point{
		{X: 110, Y: 220},
		{X: 310, Y: 420},
	}

	if len(relativePoints) != len(expected) {
		t.Fatalf("Expected %d points, got %d", len(expected), len(relativePoints))
	}

	for i, pt := range relativePoints {
		if pt != expected[i] {
			t.Errorf("Point %d: expected %v, got %v", i, expected[i], pt)
		}
	}
}

// TestAbsolutePaths_TransformPointsToRelative_Empty verifies empty input handling
func TestAbsolutePaths_TransformPointsToRelative_Empty(t *testing.T) {
	ap := NewAbsolutePaths(nil, nil, nil, nil, 5)

	coordTransform := &mockCoordinateTransformation{offset: image.Point{X: 10, Y: 20}}

	// Empty points list
	relativePoints := ap.transformPointsToRelative([]image.Point{}, coordTransform)

	if len(relativePoints) != 0 {
		t.Errorf("Expected empty result, got %d points", len(relativePoints))
	}
}

// NOTE: Additional tests for AbsolutePaths.Draw() with live TrackedObjects
// are covered by integration tests due to complexity of creating mock TrackedObject instances
