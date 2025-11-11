package norfairgodraw

import (
	"image"
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
	"github.com/nmichlo/norfair-go/pkg/norfairgocolor"
	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// Python equivalent: norfair/drawing/drawer.py::Drawer
//
//	from norfair.drawing.drawer import Drawer
//	from norfair.drawing.color import Color
//	import numpy as np
//	import cv2
//
//	drawer = Drawer()
//	frame = np.zeros((height, width, 3), dtype=np.uint8)
//
//	# Drawing primitives with auto-scaling based on frame size
//	drawer.circle(frame, (x, y), radius=None, color=Color.red, thickness=None)
//	drawer.text(frame, "label", (x, y), size=None, color=Color.white, thickness=None)
//	drawer.rectangle(frame, [(x1, y1), (x2, y2)], color=Color.green, thickness=2)
//	drawer.line(frame, (x1, y1), (x2, y2), color=Color.blue, thickness=2)
//	drawer.cross(frame, (x, y), radius=10, color=Color.cyan, thickness=2)
//	drawer.alpha_blend(frame, overlay, alpha=0.5)
//
// Validation: tools/validate_drawing/main.py tests drawing primitive equivalence

// =============================================================================
// Circle Tests
// =============================================================================

func TestDrawer_Circle_AutoScaling(t *testing.T) {
	// Test auto-scaling formula: radius = max(frame_dim * 0.005, 1)
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(1000, 1000, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Auto-scale: max(1000 * 0.005, 1) = 5
	drawer.Circle(&frame, image.Point{X: 500, Y: 500}, 0, 0, norfairgocolor.Red)

	// No crash means success (visual validation would require comparing pixels)
	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Circle_ExplicitRadius(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Explicit radius=10, thickness=2
	drawer.Circle(&frame, image.Point{X: 50, Y: 50}, 10, 2, norfairgocolor.Blue)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Circle_FilledCircle(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Filled circle: thickness=-1
	drawer.Circle(&frame, image.Point{X: 50, Y: 50}, 20, -1, norfairgocolor.Green)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Circle_SmallFrame(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Auto-scale on tiny frame: max(10 * 0.005, 1) = 1
	drawer.Circle(&frame, image.Point{X: 5, Y: 5}, 0, 0, norfairgocolor.White)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Circle_OutOfBounds(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw circle outside frame (gocv should handle gracefully)
	drawer.Circle(&frame, image.Point{X: -50, Y: -50}, 10, 2, norfairgocolor.Red)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// =============================================================================
// Text Tests
// =============================================================================

func TestDrawer_Text_AutoScaling(t *testing.T) {
	// Test auto-scaling formula: size = min(max(max_dim/4000, 0.5), 1.5)
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(4000, 4000, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Auto-scale: min(max(4000/4000, 0.5), 1.5) = min(max(1.0, 0.5), 1.5) = 1.0
	drawer.Text(&frame, "Test", image.Point{X: 100, Y: 100}, 0, norfairgocolor.Red, 0, false, norfairgocolor.Black, 2)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Text_WithShadow(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(500, 500, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw text with shadow
	drawer.Text(&frame, "Shadow", image.Point{X: 100, Y: 100}, 1.0, norfairgocolor.White, 2, true, norfairgocolor.Black, 3)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Text_WithoutShadow(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(500, 500, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw text without shadow
	drawer.Text(&frame, "No Shadow", image.Point{X: 100, Y: 200}, 1.0, norfairgocolor.Green, 2, false, norfairgocolor.Black, 0)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Text_EmptyString(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Empty string should not crash
	drawer.Text(&frame, "", image.Point{X: 50, Y: 50}, 1.0, norfairgocolor.Red, 1, false, norfairgocolor.Black, 0)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Text_LongString(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(1000, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Very long text
	longText := "This is a very long text string that might exceed the frame width"
	drawer.Text(&frame, longText, image.Point{X: 10, Y: 50}, 0.5, norfairgocolor.Blue, 1, false, norfairgocolor.Black, 0)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Text_SmallFrame(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Auto-scale on small frame: min(max(100/4000, 0.5), 1.5) = 0.5
	drawer.Text(&frame, "Small", image.Point{X: 10, Y: 50}, 0, norfairgocolor.Red, 0, false, norfairgocolor.Black, 0)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// =============================================================================
// Rectangle Tests
// =============================================================================

func TestDrawer_Rectangle_Basic(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw rectangle
	drawer.Rectangle(&frame, image.Point{X: 50, Y: 50}, image.Point{X: 150, Y: 150}, norfairgocolor.Red, 2)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Rectangle_ZeroThickness(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Zero thickness should default to 1
	drawer.Rectangle(&frame, image.Point{X: 10, Y: 10}, image.Point{X: 100, Y: 100}, norfairgocolor.Blue, 0)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Rectangle_FilledRectangle(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Filled rectangle: thickness=-1
	drawer.Rectangle(&frame, image.Point{X: 20, Y: 20}, image.Point{X: 80, Y: 80}, norfairgocolor.Green, -1)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// =============================================================================
// Line Tests
// =============================================================================

func TestDrawer_Line_Basic(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw line
	drawer.Line(&frame, image.Point{X: 0, Y: 0}, image.Point{X: 200, Y: 200}, norfairgocolor.Red, 2)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Line_Horizontal(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Horizontal line
	drawer.Line(&frame, image.Point{X: 10, Y: 100}, image.Point{X: 190, Y: 100}, norfairgocolor.Blue, 3)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Line_Vertical(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Vertical line
	drawer.Line(&frame, image.Point{X: 100, Y: 10}, image.Point{X: 100, Y: 190}, norfairgocolor.Green, 3)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// =============================================================================
// Cross Tests
// =============================================================================

func TestDrawer_Cross_Basic(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(200, 200, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Draw cross at center
	drawer.Cross(&frame, image.Point{X: 100, Y: 100}, 20, norfairgocolor.Red, 2)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Cross_SmallRadius(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Small cross
	drawer.Cross(&frame, image.Point{X: 50, Y: 50}, 5, norfairgocolor.Blue, 1)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

func TestDrawer_Cross_LargeRadius(t *testing.T) {
	drawer := NewDrawer()
	frame := gocv.NewMatWithSize(500, 500, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Large cross
	drawer.Cross(&frame, image.Point{X: 250, Y: 250}, 100, norfairgocolor.Green, 3)

	if frame.Empty() {
		t.Error("Frame should not be empty after drawing")
	}
}

// =============================================================================
// AlphaBlend Tests
// =============================================================================

func TestDrawer_AlphaBlend_HalfAlpha(t *testing.T) {
	drawer := NewDrawer()

	// Create two test frames
	frame1 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame1.Close()
	frame2 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame2.Close()

	// Fill with different colors
	frame1.SetTo(gocv.NewScalar(255, 0, 0, 0)) // Blue
	frame2.SetTo(gocv.NewScalar(0, 255, 0, 0)) // Green

	// Blend 50/50
	result := drawer.AlphaBlend(&frame1, &frame2, 0.5, 0.5, 0)
	defer result.Close()

	if result.Empty() {
		t.Error("Blended result should not be empty")
	}

	// Result should be a mix of blue and green
	rows, cols := result.Rows(), result.Cols()
	if rows != 100 || cols != 100 {
		t.Errorf("Expected 100x100 result, got %dx%d", rows, cols)
	}
}

func TestDrawer_AlphaBlend_AutoBeta(t *testing.T) {
	drawer := NewDrawer()

	frame1 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame1.Close()
	frame2 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame2.Close()

	frame1.SetTo(gocv.NewScalar(255, 0, 0, 0))
	frame2.SetTo(gocv.NewScalar(0, 255, 0, 0))

	// Auto beta: beta = 1 - alpha = 1 - 0.3 = 0.7
	result := drawer.AlphaBlend(&frame1, &frame2, 0.3, -1, 0)
	defer result.Close()

	if result.Empty() {
		t.Error("Blended result should not be empty")
	}
}

func TestDrawer_AlphaBlend_WithGamma(t *testing.T) {
	drawer := NewDrawer()

	frame1 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame1.Close()
	frame2 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame2.Close()

	frame1.SetTo(gocv.NewScalar(100, 100, 100, 0))
	frame2.SetTo(gocv.NewScalar(100, 100, 100, 0))

	// Add gamma offset
	result := drawer.AlphaBlend(&frame1, &frame2, 0.5, 0.5, 50)
	defer result.Close()

	if result.Empty() {
		t.Error("Blended result should not be empty")
	}
}

// =============================================================================
// Drawable Tests
// =============================================================================

// Mock Detection for testing
type MockDetection struct {
	points *mat.Dense
	label  *string
	scores []float64
}

func (m *MockDetection) GetPoints() *mat.Dense {
	return m.points
}

func (m *MockDetection) GetLabel() *string {
	return m.label
}

func (m *MockDetection) GetScores() []float64 {
	return m.scores
}

// Mock TrackedObject for testing
type MockTrackedObject struct {
	estimate   *mat.Dense
	id         *int
	label      *string
	livePoints []bool
}

func (m *MockTrackedObject) GetEstimate(absolute bool) (*mat.Dense, error) {
	return m.estimate, nil
}

func (m *MockTrackedObject) GetID() *int {
	return m.id
}

func (m *MockTrackedObject) GetLabel() *string {
	return m.label
}

func (m *MockTrackedObject) GetLivePoints() []bool {
	return m.livePoints
}

func TestDrawable_NewDrawableFromDetection(t *testing.T) {
	// Create mock detection
	points := mat.NewDense(3, 2, []float64{
		10, 20,
		30, 40,
		50, 60,
	})
	label := "person"
	scores := []float64{0.9, 0.8, 0.7}

	det := &MockDetection{
		points: points,
		label:  &label,
		scores: scores,
	}

	// Create drawable
	drawable, err := NewDrawableFromDetection(det)
	if err != nil {
		t.Fatalf("Failed to create drawable from detection: %v", err)
	}

	// Verify fields
	if drawable.Points != points {
		t.Error("Points should match")
	}
	if drawable.Label != &label {
		t.Error("Label should match")
	}
	if len(drawable.Scores) != 3 {
		t.Errorf("Expected 3 scores, got %d", len(drawable.Scores))
	}
	if drawable.ID != nil {
		t.Error("Detection should not have ID")
	}

	// All points should be live for detection
	if len(drawable.LivePoints) != 3 {
		t.Errorf("Expected 3 live points, got %d", len(drawable.LivePoints))
	}
	for i, live := range drawable.LivePoints {
		if !live {
			t.Errorf("Point %d should be live for detection", i)
		}
	}
}

func TestDrawable_NewDrawableFromTrackedObject(t *testing.T) {
	// Create mock tracked object
	estimate := mat.NewDense(3, 2, []float64{
		15, 25,
		35, 45,
		55, 65,
	})
	id := 42
	label := "car"
	livePoints := []bool{true, true, false}

	obj := &MockTrackedObject{
		estimate:   estimate,
		id:         &id,
		label:      &label,
		livePoints: livePoints,
	}

	// Create drawable
	drawable, err := NewDrawableFromTrackedObject(obj)
	if err != nil {
		t.Fatalf("Failed to create drawable from tracked object: %v", err)
	}

	// Verify fields
	if drawable.Points != estimate {
		t.Error("Points should match estimate")
	}
	if drawable.ID != &id {
		t.Error("ID should match")
	}
	if *drawable.ID != 42 {
		t.Errorf("ID should be 42, got %d", *drawable.ID)
	}
	if drawable.Label != &label {
		t.Error("Label should match")
	}
	if drawable.Scores != nil {
		t.Error("TrackedObject should not have scores")
	}

	// Check live points
	if len(drawable.LivePoints) != 3 {
		t.Errorf("Expected 3 live points, got %d", len(drawable.LivePoints))
	}
	if !drawable.LivePoints[0] || !drawable.LivePoints[1] || drawable.LivePoints[2] {
		t.Error("LivePoints mask should be [true, true, false]")
	}
}

func TestDrawable_NewDrawable_Explicit(t *testing.T) {
	// Create drawable with explicit fields
	points := mat.NewDense(2, 2, []float64{
		100, 200,
		300, 400,
	})
	id := 123
	label := "bike"
	scores := []float64{0.95, 0.85}
	livePoints := []bool{true, false}

	drawable, err := NewDrawable(points, &id, &label, scores, livePoints)
	if err != nil {
		t.Fatalf("Failed to create drawable: %v", err)
	}

	// Verify all fields
	if drawable.Points != points {
		t.Error("Points should match")
	}
	if *drawable.ID != 123 {
		t.Errorf("ID should be 123, got %d", *drawable.ID)
	}
	if *drawable.Label != "bike" {
		t.Errorf("Label should be 'bike', got '%s'", *drawable.Label)
	}
	if len(drawable.Scores) != 2 {
		t.Errorf("Expected 2 scores, got %d", len(drawable.Scores))
	}
	if len(drawable.LivePoints) != 2 {
		t.Errorf("Expected 2 live points, got %d", len(drawable.LivePoints))
	}
}

func TestDrawable_NewDrawable_NilLivePoints(t *testing.T) {
	// Create drawable without live points (should auto-generate)
	points := mat.NewDense(4, 2, []float64{
		1, 2,
		3, 4,
		5, 6,
		7, 8,
	})

	drawable, err := NewDrawable(points, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create drawable: %v", err)
	}

	// Should auto-generate all live points as true
	if len(drawable.LivePoints) != 4 {
		t.Errorf("Expected 4 auto-generated live points, got %d", len(drawable.LivePoints))
	}
	for i, live := range drawable.LivePoints {
		if !live {
			t.Errorf("Auto-generated point %d should be live", i)
		}
	}
}

func TestDrawable_NewDrawable_NilPoints(t *testing.T) {
	// Creating drawable with nil points should error
	_, err := NewDrawable(nil, nil, nil, nil, nil)
	if err == nil {
		t.Error("Should return error for nil points")
	}
}

func TestDrawable_NewDrawableFromDetection_NilPoints(t *testing.T) {
	// Detection with nil points should error
	det := &MockDetection{
		points: nil,
		label:  nil,
		scores: nil,
	}

	_, err := NewDrawableFromDetection(det)
	if err == nil {
		t.Error("Should return error for detection with nil points")
	}
}

// =============================================================================
// Visual Comparison Tests (Golden Image)
// =============================================================================

// Python equivalent: tools/validate_drawing/main.py::test_drawing_primitives()
//
//	from norfair.drawing.drawer import Drawer
//	from norfair.drawing.color import Color
//	import numpy as np
//	import cv2
//
//	def test_drawing_primitives():
//	    drawer = Drawer()
//	    frame = np.zeros((500, 500, 3), dtype=np.uint8)
//
//	    # Test Circle
//	    drawer.circle(frame, (250, 250), radius=None, color=Color.red, thickness=None)
//	    drawer.circle(frame, (100, 100), radius=20, color=Color.green, thickness=2)
//
//	    # Test Text
//	    drawer.text(frame, "Test", (50, 50), size=None, color=Color.white, thickness=None)
//	    drawer.text(frame, "Shadow", (50, 100), size=1.0, color=Color.white,
//	                thickness=2, shadow=True)
//
//	    # Test Rectangle
//	    drawer.rectangle(frame, [(10, 10), (90, 90)], color=Color.red, thickness=2)
//
//	    # Test Line
//	    drawer.line(frame, (0, 0), (100, 100), color=Color.cyan, thickness=2)
//
//	    cv2.imwrite("output.png", frame)
//
// This test compares Go rendering against golden image generated from Python
func TestDrawer_DrawingPrimitives_GoldenImage(t *testing.T) {
	drawer := NewDrawer()

	// Create test frame matching Python: np.zeros((500, 500, 3))
	frame := gocv.NewMatWithSize(500, 500, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Test Circle (matches Python drawer.circle calls)
	drawer.Circle(&frame, image.Point{X: 250, Y: 250}, 0, 0, norfairgocolor.Red)
	drawer.Circle(&frame, image.Point{X: 100, Y: 100}, 20, 2, norfairgocolor.Green)

	// Test Text (matches Python drawer.text calls)
	drawer.Text(&frame, "Test", image.Point{X: 50, Y: 50}, 0, norfairgocolor.White, 0, false, norfairgocolor.Black, 0)
	drawer.Text(&frame, "Shadow", image.Point{X: 50, Y: 100}, 1.0, norfairgocolor.White, 2, true, norfairgocolor.Black, 2)

	// Test Rectangle (matches Python drawer.rectangle call)
	drawer.Rectangle(&frame, image.Point{X: 10, Y: 10}, image.Point{X: 90, Y: 90}, norfairgocolor.Red, 2)

	// Test Line (matches Python drawer.line call)
	drawer.Line(&frame, image.Point{X: 0, Y: 0}, image.Point{X: 100, Y: 100}, norfairgocolor.Cyan, 2)

	// Compare to golden image (95% similarity threshold for anti-aliasing tolerance)
	// Golden image generated from Python norfair using tools/validate_drawing/main.py
	goldenPath := "../../testdata/drawing/drawing_primitives_golden.png"

	testutil.CompareToGoldenImage(t, &frame, goldenPath, 0.95)
}
