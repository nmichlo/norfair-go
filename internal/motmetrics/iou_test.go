package motmetrics

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
)

// TestIouDistance_PerfectOverlap verifies IoU distance for identical boxes
func TestIouDistance_PerfectOverlap(t *testing.T) {
	box1 := []float64{0, 0, 10, 10}
	box2 := []float64{0, 0, 10, 10}

	distance := IouDistance(box1, box2)
	testutil.AssertAlmostEqual(t, distance, 0.0, 1e-10, "Perfect overlap should have distance 0")
}

// TestIouDistance_NoOverlap verifies IoU distance for non-overlapping boxes
func TestIouDistance_NoOverlap(t *testing.T) {
	box1 := []float64{0, 0, 10, 10}
	box2 := []float64{20, 20, 30, 30}

	distance := IouDistance(box1, box2)
	testutil.AssertAlmostEqual(t, distance, 1.0, 1e-10, "No overlap should have distance 1.0")
}

// TestIouDistance_PartialOverlap verifies IoU distance for partial overlap
func TestIouDistance_PartialOverlap(t *testing.T) {
	// Two 10x10 boxes with 5x10 overlap
	// Area1 = 100, Area2 = 100, Intersection = 50, Union = 150
	// IoU = 50/150 = 1/3, Distance = 1 - 1/3 = 2/3
	box1 := []float64{0, 0, 10, 10}
	box2 := []float64{5, 0, 15, 10}

	distance := IouDistance(box1, box2)
	expected := 1.0 - (1.0 / 3.0) // 2/3
	testutil.AssertAlmostEqual(t, distance, expected, 1e-10, "Partial overlap (50% area) distance")
}

// TestIouDistance_ContainedBox verifies IoU when one box contains another
func TestIouDistance_ContainedBox(t *testing.T) {
	// Small box inside large box
	// Intersection = 25, Union = 100
	// IoU = 25/100 = 0.25, Distance = 0.75
	box1 := []float64{0, 0, 10, 10}   // Area 100
	box2 := []float64{2.5, 2.5, 7.5, 7.5} // Area 25, fully contained

	distance := IouDistance(box1, box2)
	expected := 1.0 - 0.25 // 0.75
	testutil.AssertAlmostEqual(t, distance, expected, 1e-10, "Contained box distance")
}

// TestIouDistance_AdjacentBoxes verifies IoU for adjacent (touching) boxes
func TestIouDistance_AdjacentBoxes(t *testing.T) {
	// Two boxes touching at edge (no overlap)
	box1 := []float64{0, 0, 10, 10}
	box2 := []float64{10, 0, 20, 10} // Touches at x=10

	distance := IouDistance(box1, box2)
	testutil.AssertAlmostEqual(t, distance, 1.0, 1e-10, "Adjacent boxes should have distance 1.0")
}

// TestIouDistance_SmallOverlap verifies IoU for minimal overlap
func TestIouDistance_SmallOverlap(t *testing.T) {
	// Very small overlap region
	box1 := []float64{0, 0, 10, 10}
	box2 := []float64{9, 9, 19, 19}

	// Intersection = 1x1 = 1, Union = 100 + 100 - 1 = 199
	// IoU = 1/199, Distance ≈ 0.995
	distance := IouDistance(box1, box2)
	expected := 1.0 - (1.0 / 199.0)
	testutil.AssertAlmostEqual(t, distance, expected, 1e-10, "Small overlap distance")
}

// TestIouDistance_FloatingPoint verifies IoU works with floating-point coordinates
func TestIouDistance_FloatingPoint(t *testing.T) {
	box1 := []float64{0.5, 0.5, 10.5, 10.5}
	box2 := []float64{5.5, 0.5, 15.5, 10.5}

	// Same as TestIouDistance_PartialOverlap but with 0.5 offset
	distance := IouDistance(box1, box2)
	expected := 1.0 - (1.0 / 3.0)
	testutil.AssertAlmostEqual(t, distance, expected, 1e-10, "Floating-point coordinates")
}

// TestIouDistance_InvalidBox1 verifies panic on invalid box1
func TestIouDistance_InvalidBox1(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid box1, but no panic occurred")
		}
	}()

	box1 := []float64{10, 10, 0, 0} // x_max < x_min
	box2 := []float64{0, 0, 10, 10}
	IouDistance(box1, box2)
}

// TestIouDistance_InvalidBox2 verifies panic on invalid box2
func TestIouDistance_InvalidBox2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid box2, but no panic occurred")
		}
	}()

	box1 := []float64{0, 0, 10, 10}
	box2 := []float64{0, 10, 10, 0} // y_max < y_min
	IouDistance(box1, box2)
}

// TestIouDistance_WrongLength verifies panic on wrong-length box
func TestIouDistance_WrongLength(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for wrong-length box, but no panic occurred")
		}
	}()

	box1 := []float64{0, 0, 10} // Only 3 elements
	box2 := []float64{0, 0, 10, 10}
	IouDistance(box1, box2)
}

// TestComputeIoUMatrix verifies distance matrix computation
func TestComputeIoUMatrix(t *testing.T) {
	gtBBoxes := [][]float64{
		{0, 0, 10, 10},   // GT 0
		{20, 20, 30, 30}, // GT 1
	}

	predBBoxes := [][]float64{
		{0, 0, 10, 10},   // Perfect match with GT 0
		{25, 25, 35, 35}, // Overlaps with GT 1
		{50, 50, 60, 60}, // No overlap with any GT
	}

	matrix := ComputeIoUMatrix(gtBBoxes, predBBoxes)

	// Verify dimensions
	if len(matrix) != 2 {
		t.Errorf("Expected matrix with 2 rows, got %d", len(matrix))
	}
	for i, row := range matrix {
		if len(row) != 3 {
			t.Errorf("Expected row %d with 3 columns, got %d", i, len(row))
		}
	}

	// Verify specific distances
	testutil.AssertAlmostEqual(t, matrix[0][0], 0.0, 1e-10, "GT0-Pred0: perfect match")
	testutil.AssertAlmostEqual(t, matrix[0][2], 1.0, 1e-10, "GT0-Pred2: no overlap")
	testutil.AssertAlmostEqual(t, matrix[1][2], 1.0, 1e-10, "GT1-Pred2: no overlap")

	// GT1-Pred1 should have some overlap
	if matrix[1][1] <= 0.0 || matrix[1][1] >= 1.0 {
		t.Errorf("GT1-Pred1 should have partial overlap distance in (0, 1), got %.3f", matrix[1][1])
	}
}

// TestComputeIoUMatrix_Empty verifies handling of empty inputs
func TestComputeIoUMatrix_Empty(t *testing.T) {
	// Empty GT boxes
	gtBBoxes := [][]float64{}
	predBBoxes := [][]float64{{0, 0, 10, 10}}
	matrix := ComputeIoUMatrix(gtBBoxes, predBBoxes)

	if len(matrix) != 0 {
		t.Errorf("Expected empty matrix for empty GT, got %d rows", len(matrix))
	}

	// Empty pred boxes
	gtBBoxes = [][]float64{{0, 0, 10, 10}}
	predBBoxes = [][]float64{}
	matrix = ComputeIoUMatrix(gtBBoxes, predBBoxes)

	if len(matrix) != 1 {
		t.Errorf("Expected 1 row for single GT, got %d", len(matrix))
	}
	if len(matrix[0]) != 0 {
		t.Errorf("Expected 0 columns for empty pred, got %d", len(matrix[0]))
	}
}

// TestComputeIoUMatrix_LargeSet verifies matrix computation scales
func TestComputeIoUMatrix_LargeSet(t *testing.T) {
	// Generate 10 GT boxes and 15 pred boxes
	gtBBoxes := make([][]float64, 10)
	for i := range gtBBoxes {
		x := float64(i * 15)
		gtBBoxes[i] = []float64{x, 0, x + 10, 10}
	}

	predBBoxes := make([][]float64, 15)
	for i := range predBBoxes {
		x := float64(i * 10)
		predBBoxes[i] = []float64{x, 0, x + 10, 10}
	}

	matrix := ComputeIoUMatrix(gtBBoxes, predBBoxes)

	// Verify dimensions
	if len(matrix) != 10 {
		t.Errorf("Expected 10 rows, got %d", len(matrix))
	}
	for i, row := range matrix {
		if len(row) != 15 {
			t.Errorf("Expected row %d with 15 columns, got %d", i, len(row))
		}
	}

	// Verify all distances are in valid range [0, 1]
	for i, row := range matrix {
		for j, dist := range row {
			if dist < 0.0 || dist > 1.0 {
				t.Errorf("Distance[%d][%d] = %.3f outside valid range [0, 1]", i, j, dist)
			}
		}
	}
}
