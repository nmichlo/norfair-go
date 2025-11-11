package norfairgodraw

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
	"github.com/nmichlo/norfair-go/pkg/norfairgocolor"
	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Basic Functionality Tests
// =============================================================================

func TestDrawPoints_BasicDefaults(t *testing.T) {
	// Create test frame
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Create test drawables
	points := mat.NewDense(2, 2, []float64{
		100, 100,
		200, 200,
	})
	id := 1
	label := "test"
	drawable, err := NewDrawable(points, &id, &label, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable: %v", err)
	}

	// Draw with default parameters
	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil,   // radius
		nil,   // thickness
		nil,   // color
		true,  // drawLabels
		nil,   // textSize
		true,  // drawIDs
		true,  // drawPoints
		nil,   // textThickness
		nil,   // textColor
		true,  // hideDeadPoints
		false, // drawScores
	)

	if result == nil {
		t.Error("DrawPoints should return the frame")
	}
}

func TestDrawPoints_CustomParameters(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{300, 250})
	id := 42
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	// Custom parameters
	radius := 10
	thickness := 2
	textSize := 1.5
	textThickness := 2

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		&radius,
		&thickness,
		"red",
		false,
		&textSize,
		true,
		true,
		&textThickness,
		nil,
		true,
		false,
	)

	if result == nil {
		t.Error("DrawPoints should return the frame")
	}
}

func TestDrawPoints_NilDrawables(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Should return frame unchanged
	result := DrawPoints(
		&frame,
		nil,
		nil, nil, nil, true, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints should return the frame even for nil drawables")
	}
}

func TestDrawPoints_EmptyDrawables(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Should return frame unchanged
	result := DrawPoints(
		&frame,
		[]interface{}{},
		nil, nil, nil, true, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints should return the frame for empty drawables")
	}
}

// =============================================================================
// Color Strategy Tests
// =============================================================================

func TestDrawPoints_ColorByID(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	id := 5
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, "by_id", false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with by_id color strategy")
	}
}

func TestDrawPoints_ColorByLabel(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	label := "person"
	drawable, _ := NewDrawable(points, nil, &label, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, "by_label", false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with by_label color strategy")
	}
}

func TestDrawPoints_ColorRandom(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, "random", false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with random color strategy")
	}
}

func TestDrawPoints_DirectColorHex(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, "#FF0000", false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with hex color")
	}
}

func TestDrawPoints_DirectColorName(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, "blue", false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with color name")
	}
}

func TestDrawPoints_DirectColorStruct(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, norfairgocolor.Red, false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with Color struct")
	}
}

// =============================================================================
// Text Rendering Tests
// =============================================================================

func TestDrawPoints_OnlyLabels(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	label := "car"
	drawable, _ := NewDrawable(points, nil, &label, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil,
		true, // drawLabels
		nil,
		false, // drawIDs
		true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with only labels")
	}
}

func TestDrawPoints_OnlyIDs(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	id := 123
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil,
		false, // drawLabels
		nil,
		true, // drawIDs
		true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with only IDs")
	}
}

func TestDrawPoints_OnlyScores(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	scores := []float64{0.95}
	drawable, _ := NewDrawable(points, nil, nil, scores, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil,
		false, // drawLabels
		nil,
		false, // drawIDs
		true, nil, nil, true,
		true, // drawScores
	)

	if result == nil {
		t.Error("DrawPoints failed with only scores")
	}
}

func TestDrawPoints_LabelsAndIDs(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	label := "person"
	id := 42
	drawable, _ := NewDrawable(points, &id, &label, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil,
		true, // drawLabels
		nil,
		true, // drawIDs
		true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with labels and IDs")
	}
}

func TestDrawPoints_AllTextFields(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	label := "person"
	id := 42
	scores := []float64{0.95, 0.87, 0.92}
	drawable, _ := NewDrawable(points, &id, &label, scores, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil,
		true, // drawLabels
		nil,
		true, // drawIDs
		true, nil, nil, true,
		true, // drawScores
	)

	if result == nil {
		t.Error("DrawPoints failed with all text fields")
	}
}

func TestDrawPoints_NoText(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil,
		false, // drawLabels
		nil,
		false, // drawIDs
		true, nil, nil, true,
		false, // drawScores
	)

	if result == nil {
		t.Error("DrawPoints failed with no text")
	}
}

func TestDrawPoints_CustomTextColor(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	id := 1
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, "red", true, nil, true, true, nil,
		"blue", // textColor different from point color
		true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with custom text color")
	}
}

// =============================================================================
// Point Filtering Tests
// =============================================================================

func TestDrawPoints_HideDeadPoints_Mixed(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(3, 2, []float64{
		100, 100,
		200, 200,
		300, 300,
	})
	livePoints := []bool{true, false, true}
	drawable, _ := NewDrawable(points, nil, nil, nil, livePoints)

	// Should only draw 2 points (live ones)
	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil, false, nil, false, true, nil, nil,
		true, // hideDeadPoints
		false,
	)

	if result == nil {
		t.Error("DrawPoints failed with mixed live/dead points")
	}
}

func TestDrawPoints_HideDeadPoints_AllDead(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{
		100, 100,
		200, 200,
	})
	livePoints := []bool{false, false}
	drawable, _ := NewDrawable(points, nil, nil, nil, livePoints)

	// Should skip entire object
	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil, false, nil, false, true, nil, nil,
		true, // hideDeadPoints
		false,
	)

	if result == nil {
		t.Error("DrawPoints should still return frame even when skipping all objects")
	}
}

func TestDrawPoints_ShowDeadPoints(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{
		100, 100,
		200, 200,
	})
	livePoints := []bool{false, false}
	drawable, _ := NewDrawable(points, nil, nil, nil, livePoints)

	// Should draw all points even though they're dead
	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil, false, nil, false, true, nil, nil,
		false, // hideDeadPoints
		false,
	)

	if result == nil {
		t.Error("DrawPoints failed when showing dead points")
	}
}

func TestDrawPoints_HideDeadPoints_AllLive(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(2, 2, []float64{
		100, 100,
		200, 200,
	})
	livePoints := []bool{true, true}
	drawable, _ := NewDrawable(points, nil, nil, nil, livePoints)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil, false, nil, false, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with all live points")
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestDrawPoints_SmallFrame(t *testing.T) {
	// Test radius calculation with small frame (100x100)
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{50, 50})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, // Should calculate radius as max(100*0.002, 1) = max(0.2, 1) = 1
		nil, nil, false, nil, false, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with small frame")
	}
}

func TestDrawPoints_LargeFrame(t *testing.T) {
	// Test radius calculation with large frame (1920x1080)
	frame := gocv.NewMatWithSize(1080, 1920, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{960, 540})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, // Should calculate radius as round(1920*0.002) = round(3.84) = 4
		nil, nil, false, nil, false, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with large frame")
	}
}

func TestDrawPoints_BoundaryPoints(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Points at frame boundaries
	points := mat.NewDense(4, 2, []float64{
		0, 0, // Top-left
		639, 0, // Top-right
		0, 479, // Bottom-left
		639, 479, // Bottom-right
	})
	drawable, _ := NewDrawable(points, nil, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil, false, nil, false, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with boundary points")
	}
}

func TestDrawPoints_DrawPointsFalse(t *testing.T) {
	// Test drawing only text, no points
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	points := mat.NewDense(1, 2, []float64{100, 100})
	id := 1
	drawable, _ := NewDrawable(points, &id, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable},
		nil, nil, nil, false, nil, true,
		false, // drawPoints = false
		nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with drawPoints=false")
	}
}

func TestDrawPoints_MultipleObjects(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Create multiple drawables
	points1 := mat.NewDense(1, 2, []float64{100, 100})
	id1 := 1
	drawable1, _ := NewDrawable(points1, &id1, nil, nil, nil)

	points2 := mat.NewDense(1, 2, []float64{300, 300})
	id2 := 2
	drawable2, _ := NewDrawable(points2, &id2, nil, nil, nil)

	result := DrawPoints(
		&frame,
		[]interface{}{drawable1, drawable2},
		nil, nil, "by_id", false, nil, true, true, nil, nil, true, false,
	)

	if result == nil {
		t.Error("DrawPoints failed with multiple objects")
	}
}

// =============================================================================
// Visual Comparison Tests (Golden Image)
// =============================================================================

// Python equivalent: tools/validate_draw_points/main.py::test_4_direct_color()
//
//	from norfair.drawing import draw_points
//	from norfair.drawing.drawer import Drawable
//	import numpy as np
//	import cv2
//
//	def test_4_direct_color():
//	    frame = np.zeros((480, 640, 3), dtype=np.uint8)
//	    obj1 = Drawable(points=np.array([[150, 150], [200, 200]], dtype=np.float64), live_points=np.array([True, True]))
//	    draw_points(frame, drawables=[obj1], color="#ff0000", draw_labels=False, draw_ids=False)
//	    obj2 = Drawable(points=np.array([[400, 300], [450, 350]], dtype=np.float64), live_points=np.array([True, True]))
//	    draw_points(frame, drawables=[obj2], color="blue", draw_labels=False, draw_ids=False)
//	    cv2.imwrite("python_test4_direct_color.png", frame)
func TestDrawPoints_DirectColor_GoldenImage(t *testing.T) {
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// First object with red hex color
	points1 := mat.NewDense(2, 2, []float64{150, 150, 200, 200})
	drawable1, err := NewDrawable(points1, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable1: %v", err)
	}

	DrawPoints(&frame, []interface{}{drawable1}, nil, nil, "#ff0000", false, nil, false, true, nil, nil, true, false)

	// Second object with blue color name
	points2 := mat.NewDense(2, 2, []float64{400, 300, 450, 350})
	drawable2, err := NewDrawable(points2, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable2: %v", err)
	}

	DrawPoints(&frame, []interface{}{drawable2}, nil, nil, "blue", false, nil, false, true, nil, nil, true, false)

	// Compare to golden image
	goldenPath := "../../testdata/drawing/draw_points_direct_color_golden.png"
	testutil.CompareToGoldenImage(t, &frame, goldenPath, 0.95)
}
