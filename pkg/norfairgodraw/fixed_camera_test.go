package norfairgodraw

import (
	"image"
	"testing"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// mockCoordinateTransformation for testing
type mockFixedCameraTransform struct {
	offset image.Point
}

func (m *mockFixedCameraTransform) RelToAbs(points *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	result := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		result.Set(i, 0, points.At(i, 0)-float64(m.offset.X))
		result.Set(i, 1, points.At(i, 1)-float64(m.offset.Y))
	}
	return result
}

func (m *mockFixedCameraTransform) AbsToRel(points *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	result := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		result.Set(i, 0, points.At(i, 0)+float64(m.offset.X))
		result.Set(i, 1, points.At(i, 1)+float64(m.offset.Y))
	}
	return result
}

// Python equivalent: norfair/drawing/fixed_camera.py::FixedCamera
//
//	from norfair.drawing.fixed_camera import FixedCamera
//	from norfair.camera_motion import TranslationTransformation
//	import numpy as np
//	import cv2
//
//	# Create FixedCamera with scale and attenuation
//	fc = FixedCamera(scale=2.0, attenuation=0.05)
//
//	# Apply camera motion compensation to frame
//	frame = np.zeros((480, 640, 3), dtype=np.uint8)
//	transformation = TranslationTransformation(movement_vector=np.array([10.0, 0.0]))
//	result = fc.adjust_frame(frame, transformation)
//
//	# FixedCamera creates a larger background canvas (scale * frame_size)
//	# Applies fade/attenuation to previous background
//	# Centers and composites new frame with camera motion offset
//
// Validation: tools/validate_fixed_camera/main.py tests FixedCamera equivalence

// Test 1: Constructor with defaults
func TestNewFixedCamera_Defaults(t *testing.T) {
	fc := NewFixedCamera(0, -1)

	if fc.scale != 2.0 {
		t.Errorf("Expected default scale 2.0, got %f", fc.scale)
	}
	if fc.attenuationFactor != 1.0-0.05 {
		t.Errorf("Expected default attenuation factor 0.95, got %f", fc.attenuationFactor)
	}
	if fc.background != nil {
		t.Error("Background should be nil (lazy init)")
	}
}

// Test 2: Constructor with custom values
func TestNewFixedCamera_CustomValues(t *testing.T) {
	fc := NewFixedCamera(3.0, 0.1)

	if fc.scale != 3.0 {
		t.Errorf("Expected scale 3.0, got %f", fc.scale)
	}
	if fc.attenuationFactor != 1.0-0.1 {
		t.Errorf("Expected attenuation factor 0.9, got %f", fc.attenuationFactor)
	}
}

// Test 3: Lazy initialization of background
func TestFixedCamera_LazyInit(t *testing.T) {
	fc := NewFixedCamera(2.0, 0.05)
	defer fc.Close()

	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	transform := &mockFixedCameraTransform{offset: image.Point{X: 0, Y: 0}}

	// First call should initialize background
	result := fc.AdjustFrame(&frame, transform)

	if fc.background == nil {
		t.Fatal("Background should be initialized after first AdjustFrame call")
	}

	expectedWidth := int(float64(640) * 2.0)
	expectedHeight := int(float64(480) * 2.0)

	if fc.background.Cols() != expectedWidth {
		t.Errorf("Expected background width %d, got %d", expectedWidth, fc.background.Cols())
	}
	if fc.background.Rows() != expectedHeight {
		t.Errorf("Expected background height %d, got %d", expectedHeight, fc.background.Rows())
	}

	// Result should be same as background
	if result.Cols() != fc.background.Cols() || result.Rows() != fc.background.Rows() {
		t.Error("Result dimensions should match background")
	}
}

// Test 4: Background fade (attenuation)
func TestFixedCamera_Fade(t *testing.T) {
	fc := NewFixedCamera(2.0, 0.5) // High attenuation for testing
	defer fc.Close()

	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()
	frame.SetTo(gocv.NewScalar(255, 255, 255, 0)) // White frame

	transform := &mockFixedCameraTransform{offset: image.Point{X: 0, Y: 0}}

	// First frame
	fc.AdjustFrame(&frame, transform)

	// Get a pixel value from center
	val1 := fc.background.GetVecbAt(100, 100)

	// Second frame with black (fade should occur)
	frame.SetTo(gocv.NewScalar(0, 0, 0, 0))
	fc.AdjustFrame(&frame, transform)

	val2 := fc.background.GetVecbAt(100, 100)

	// Value should decrease due to fade (multiplied by attenuationFactor = 0.5)
	// Note: Exact comparison is tricky due to rounding, just check it decreased
	if val2[0] >= val1[0] {
		t.Errorf("Expected fade to decrease pixel value, got %d >= %d", val2[0], val1[0])
	}
}

// Test 5: Center positioning (no transformation)
func TestFixedCamera_CenterPositioning(t *testing.T) {
	fc := NewFixedCamera(2.0, 0.05)
	defer fc.Close()

	frameWidth := 640
	frameHeight := 480
	frame := gocv.NewMatWithSize(frameHeight, frameWidth, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Identity transformation (no camera motion)
	transform := &mockFixedCameraTransform{offset: image.Point{X: 0, Y: 0}}

	result := fc.AdjustFrame(&frame, transform)

	// Background should be scaled
	bgWidth := int(float64(frameWidth)*2.0 + 0.5)
	bgHeight := int(float64(frameHeight)*2.0 + 0.5)

	if result.Cols() != bgWidth {
		t.Errorf("Expected result width %d, got %d", bgWidth, result.Cols())
	}
	if result.Rows() != bgHeight {
		t.Errorf("Expected result height %d, got %d", bgHeight, result.Rows())
	}
}

// Test 6: Translation transformation
func TestFixedCamera_Translation(t *testing.T) {
	fc := NewFixedCamera(2.0, 0.05)
	defer fc.Close()

	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Camera moved right by 10 pixels
	transform := &mockFixedCameraTransform{offset: image.Point{X: 10, Y: 0}}

	result := fc.AdjustFrame(&frame, transform)

	// Result should be larger canvas
	if result.Cols() <= frame.Cols() {
		t.Error("Result should be larger than input frame")
	}
	if result.Rows() <= frame.Rows() {
		t.Error("Result should be larger than input frame")
	}
}

// Test 7: Boundary cropping (movement exceeds scale)
func TestFixedCamera_BoundaryCropping(t *testing.T) {
	fc := NewFixedCamera(1.5, 0.05) // Small scale
	defer fc.Close()

	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Large camera movement that exceeds scale
	transform := &mockFixedCameraTransform{offset: image.Point{X: 100, Y: 100}}

	// Should not panic, should handle gracefully
	result := fc.AdjustFrame(&frame, transform)

	if result.Empty() {
		t.Error("Result should not be empty even with excessive movement")
	}
}

// Test 8: Close method
func TestFixedCamera_Close(t *testing.T) {
	fc := NewFixedCamera(2.0, 0.05)

	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	transform := &mockFixedCameraTransform{offset: image.Point{X: 0, Y: 0}}

	// Initialize background
	fc.AdjustFrame(&frame, transform)

	if fc.background == nil {
		t.Fatal("Background should be initialized")
	}

	// Close
	fc.Close()

	if fc.background != nil {
		t.Error("Background should be nil after Close()")
	}

	// Should be safe to call Close() multiple times
	fc.Close()
}

// Test 9: Multiple frames with accumulation
func TestFixedCamera_MultipleFrames(t *testing.T) {
	fc := NewFixedCamera(2.0, 0.01) // Low attenuation
	defer fc.Close()

	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	transform := &mockFixedCameraTransform{offset: image.Point{X: 0, Y: 0}}

	// Process multiple frames
	for i := 0; i < 10; i++ {
		result := fc.AdjustFrame(&frame, transform)
		if result.Empty() {
			t.Fatalf("Result should not be empty on frame %d", i)
		}
	}

	// Background should still be valid
	if fc.background == nil || fc.background.Empty() {
		t.Error("Background should be valid after multiple frames")
	}
}

// Test 10: Scale variations
func TestFixedCamera_ScaleVariations(t *testing.T) {
	scales := []float64{1.5, 2.0, 3.0, 5.0}
	frameSize := 100

	for _, scale := range scales {
		t.Run("scale="+string(rune(int(scale))), func(t *testing.T) {
			fc := NewFixedCamera(scale, 0.05)
			defer fc.Close()

			frame := gocv.NewMatWithSize(frameSize, frameSize, gocv.MatTypeCV8UC3)
			defer frame.Close()

			transform := &mockFixedCameraTransform{offset: image.Point{X: 0, Y: 0}}

			result := fc.AdjustFrame(&frame, transform)

			expectedSize := int(float64(frameSize)*scale + 0.5)

			if result.Cols() != expectedSize {
				t.Errorf("Scale %f: expected width %d, got %d", scale, expectedSize, result.Cols())
			}
			if result.Rows() != expectedSize {
				t.Errorf("Scale %f: expected height %d, got %d", scale, expectedSize, result.Rows())
			}
		})
	}
}
