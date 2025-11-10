package drawing

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Basic Functionality Tests
// =============================================================================

func TestDrawBoxes_BasicDefaults(t *testing.T) {
	// Create test frame
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Create test box (2 points: top-left and bottom-right)
	points := mat.NewDense(2, 2, []float64{
		100, 100, // top-left
		200, 200, // bottom-right
	})
	id := 1
	label := "person"
	drawable, err := NewDrawable(points, &id, &label, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable: %v", err)
	}

	// Draw with default parameters
	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil,   // color
		nil,   // thickness
		false, // drawLabels (default false for boxes!)
		nil,   // textSize
		true,  // drawIDs
		nil,   // textColor
		nil,   // textThickness
		true,  // drawBox
		false, // drawScores
	)

	if result == nil {
		t.Error("DrawBoxes should return the frame")
	}
}

func TestDrawBoxes_CustomParameters(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{
		100, 100,
		300, 250,
	})
	id := 42
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	// Custom parameters
	thickness := 3
	textSize := 1.2
	textThickness := 2

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"blue",
		&thickness,
		false,
		&textSize,
		true,
		nil,
		&textThickness,
		true,
		false,
	)

	if result == nil {
		t.Error("DrawBoxes should return the frame")
	}
}

func TestDrawBoxes_NilDrawables(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Should return frame unchanged
	result := DrawBoxes(
		&frame,
		nil,
		nil, nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes should return the frame even for nil drawables")
	}
}

// =============================================================================
// Color Strategy Tests
// =============================================================================

func TestDrawBoxes_ColorByID(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	id := 5
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"by_id", nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with by_id color strategy")
	}
}

func TestDrawBoxes_ColorByLabel(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	label := "car"
	drawable, _ := NewDrawable(points, nil, &label, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"by_label", nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with by_label color strategy")
	}
}

func TestDrawBoxes_ColorRandom(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"random", nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with random color strategy")
	}
}

func TestDrawBoxes_DirectColorHex(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"#00FF00", nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with hex color")
	}
}

func TestDrawBoxes_DirectColorName(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"yellow", nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with color name")
	}
}

// =============================================================================
// Text Rendering Tests
// =============================================================================

func TestDrawBoxes_DrawLabelsTrue(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	label := "person"
	drawable, _ := NewDrawable(points, nil, &label, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil,
		true, // drawLabels
		nil, false, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with drawLabels=true")
	}
}

func TestDrawBoxes_DrawIDsOnly(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	id := 123
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil,
		false, // drawLabels
		nil,
		true, // drawIDs
		nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with ID only")
	}
}

func TestDrawBoxes_DrawScores(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	scores := []float64{0.95, 0.87}
	drawable, _ := NewDrawable(points, nil, nil, scores, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil, false, nil, false, nil, nil, true,
		true, // drawScores
	)

	if result == nil {
		t.Error("DrawBoxes failed with scores")
	}
}

func TestDrawBoxes_AllTextFields(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	label := "car"
	id := 42
	scores := []float64{0.91, 0.88}
	drawable, _ := NewDrawable(points, &id, &label, scores, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil,
		true, // drawLabels
		nil,
		true, // drawIDs
		nil, nil, true,
		true, // drawScores
	)

	if result == nil {
		t.Error("DrawBoxes failed with all text fields")
	}
}

func TestDrawBoxes_CustomTextColor(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	id := 1
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		"red", nil, false, nil, true,
		"white", // textColor different from box color
		nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with custom text color")
	}
}

// =============================================================================
// Draw Box Flag Tests
// =============================================================================

func TestDrawBoxes_DrawBoxFalse(t *testing.T) {
	// Test drawing only text, no box
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	id := 1
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil, false, nil, true, nil, nil,
		false, // drawBox = false
		false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with drawBox=false")
	}
}

func TestDrawBoxes_NoTextOrBox(t *testing.T) {
	// Test with both drawBox=false and no text
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 200, 200})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil,
		false, // no labels
		nil,
		false, // no IDs
		nil, nil,
		false, // no box
		false, // no scores
	)

	if result == nil {
		t.Error("DrawBoxes should return frame even with nothing to draw")
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestDrawBoxes_SmallFrame(t *testing.T) {
	// Test thickness calculation with small frame (100x100)
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{10, 10, 90, 90})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil,
		nil, // Should calculate thickness as 100/500 = 0 (integer division)
		false, nil, false, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with small frame")
	}
}

func TestDrawBoxes_LargeFrame(t *testing.T) {
	// Test thickness calculation with large frame
	frame := gocv.NewMatWithSize(1080, 1920, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{100, 100, 500, 400})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil,
		nil, // Should calculate thickness as 1920/500 = 3
		false, nil, false, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with large frame")
	}
}

func TestDrawBoxes_InvertedBox(t *testing.T) {
	// Test box with inverted coordinates (x1 < x0, y1 < y0)
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{
		200, 200, // "top-left" is actually bottom-right
		100, 100, // "bottom-right" is actually top-left
	})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil, false, nil, false, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with inverted box")
	}
}

func TestDrawBoxes_MultipleBoxes(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Create multiple boxes
	points1 := mat.NewDense(2, 2, []float64{50, 50, 150, 150})
	id1 := 1
	drawable1, _ := NewDrawable(points1, &id1, nil, nil, nil)

	points2 := mat.NewDense(2, 2, []float64{200, 200, 350, 350})
	id2 := 2
	drawable2, _ := NewDrawable(points2, &id2, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable1, drawable2},
		"by_id", nil, false, nil, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with multiple boxes")
	}
}

func TestDrawBoxes_InvalidPointCount(t *testing.T) {
	// Test with wrong number of points (not a bbox)
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Single point instead of 2
	points := mat.NewDense(1, 2, []float64{100, 100})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil, false, nil, false, nil, nil, true, false,
	)

	// Should skip the invalid drawable but return the frame
	if result == nil {
		t.Error("DrawBoxes should return frame even when skipping invalid drawables")
	}
}

func TestDrawBoxes_BoundaryBox(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Box at frame boundaries
	points := mat.NewDense(2, 2, []float64{
		0, 0, // top-left corner
		639, 479, // bottom-right corner
	})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawBoxes(
		&frame,
		[]interface{}{drawable},
		nil, nil, false, nil, false, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawBoxes failed with boundary box")
	}
}

// =============================================================================
// Visual Comparison Tests (Golden Image)
// =============================================================================

// Python equivalent: tools/validate_draw_boxes/main.py::test_4_direct_color()
//
//	from norfair.drawing import draw_boxes
//	from norfair.drawing.drawer import Drawable
//	import numpy as np
//	import cv2
//
//	def test_4_direct_color():
//	    frame = np.zeros((480, 640, 3), dtype=np.uint8)
//	    # Red box
//	    obj1 = Drawable(points=np.array([[100, 100], [250, 200]], dtype=np.float64), live_points=np.array([True, True]))
//	    draw_boxes(frame, drawables=[obj1], color="#ff0000", draw_labels=False, draw_ids=False)
//	    # Blue box
//	    obj2 = Drawable(points=np.array([[350, 250], [550, 400]], dtype=np.float64), live_points=np.array([True, True]))
//	    draw_boxes(frame, drawables=[obj2], color="blue", draw_labels=False, draw_ids=False)
//	    cv2.imwrite("python_test4_direct_color.png", frame)
func TestDrawBoxes_DirectColor_GoldenImage(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Red box (hex color)
	points1 := mat.NewDense(2, 2, []float64{100, 100, 250, 200})
	drawable1, err := NewDrawable(points1, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable1: %v", err)
	}

	DrawBoxes(&frame, []interface{}{drawable1}, "#ff0000", nil, false, nil, false, nil, nil, true, false)

	// Blue box (color name)
	points2 := mat.NewDense(2, 2, []float64{350, 250, 550, 400})
	drawable2, err := NewDrawable(points2, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable2: %v", err)
	}

	DrawBoxes(&frame, []interface{}{drawable2}, "blue", nil, false, nil, false, nil, nil, true, false)

	// Compare to golden image
	goldenPath := "../testdata/drawing/draw_boxes_direct_color_golden.png"
	testutil.CompareToGoldenImage(t, &frame, goldenPath, 0.95)
}
