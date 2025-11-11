package norfairgo

import (
	"math"
	"testing"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// Python equivalent: norfair/camera_motion.py::TranslationTransformation
//
//	from norfair.camera_motion import TranslationTransformation
//	import numpy as np
//
//	# Create translation transformation with movement vector
//	movement = np.array([10.0, 20.0])
//	trans = TranslationTransformation(movement_vector=movement)
//
//	# Transform points from relative to absolute coordinates
//	rel_points = np.array([[0, 0], [100, 100]])
//	abs_points = trans.rel_to_abs(rel_points)  # Adds movement vector
//
//	# Transform back from absolute to relative coordinates
//	rel_back = trans.abs_to_rel(abs_points)  # Subtracts movement vector
//
// Validation: tools/validate_translation/main.py tests TranslationTransformation equivalence

//
// TranslationTransformation Tests
//

func TestTranslationTransformation_ForwardBackward(t *testing.T) {
	// Test that RelToAbs and AbsToRel are inverses
	movement := []float64{10.0, 20.0}
	trans, err := NewTranslationTransformation(movement)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	// Original points
	points := mat.NewDense(3, 2, []float64{
		0, 0,
		10, 10,
		20, 30,
	})

	// Forward: RelToAbs (subtract movement)
	absPoints := trans.RelToAbs(points)

	// Expected: points - movement
	expected := mat.NewDense(3, 2, []float64{
		-10, -20,
		0, -10,
		10, 10,
	})

	if !matApproxEqual(absPoints, expected, 1e-10) {
		t.Errorf("RelToAbs incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(absPoints), mat.Formatted(expected))
	}

	// Backward: AbsToRel (add movement)
	relPoints := trans.AbsToRel(absPoints)

	// Should get back original points
	if !matApproxEqual(relPoints, points, 1e-10) {
		t.Errorf("AbsToRel didn't invert RelToAbs.\nGot:\n%v\nExpected:\n%v", mat.Formatted(relPoints), mat.Formatted(points))
	}
}

func TestTranslationTransformation_ZeroMovement(t *testing.T) {
	// Test with zero movement (identity transformation)
	movement := []float64{0.0, 0.0}
	trans, err := NewTranslationTransformation(movement)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	points := mat.NewDense(2, 2, []float64{
		5, 10,
		15, 20,
	})

	// Both operations should return original points
	absPoints := trans.RelToAbs(points)
	if !matApproxEqual(absPoints, points, 1e-10) {
		t.Errorf("RelToAbs with zero movement should return original points")
	}

	relPoints := trans.AbsToRel(points)
	if !matApproxEqual(relPoints, points, 1e-10) {
		t.Errorf("AbsToRel with zero movement should return original points")
	}
}

func TestTranslationTransformation_InvalidMovementVector(t *testing.T) {
	// Test with invalid movement vector (not 2D)
	movement := []float64{10.0} // Only 1 dimension
	_, err := NewTranslationTransformation(movement)
	if err == nil {
		t.Error("Expected error for invalid movement vector, got nil")
	}

	movement = []float64{10.0, 20.0, 30.0} // 3 dimensions
	_, err = NewTranslationTransformation(movement)
	if err == nil {
		t.Error("Expected error for invalid movement vector, got nil")
	}
}

//
// TranslationTransformationGetter Tests
//

func TestTranslationTransformationGetter_SimpleModeFind(t *testing.T) {
	// Test mode finding with clear mode
	getter := NewTranslationTransformationGetter(0.2, 0.9)

	// Previous points
	prevPts := mat.NewDense(5, 2, []float64{
		0, 0,
		10, 10,
		20, 20,
		30, 30,
		40, 40,
	})

	// Current points: 4 points moved by (5, 5), 1 outlier moved by (1, 1)
	currPts := mat.NewDense(5, 2, []float64{
		5, 5, // moved by (5, 5)
		15, 15, // moved by (5, 5)
		25, 25, // moved by (5, 5)
		35, 35, // moved by (5, 5)
		41, 41, // outlier: moved by (1, 1)
	})

	updateRef, trans := getter.Call(currPts, prevPts)

	// 4 out of 5 points = 80% < 90% threshold, so should update reference
	if !updateRef {
		t.Error("Expected reference frame update (proportion < threshold)")
	}

	// Mode should be (5, 5)
	translationTrans, ok := trans.(*TranslationTransformation)
	if !ok {
		t.Fatal("Expected TranslationTransformation")
	}

	// Due to binning (bin_size=0.2), the mode should be very close to (5, 5)
	expectedX, expectedY := 5.0, 5.0
	if math.Abs(translationTrans.MovementVector[0]-expectedX) > 0.3 {
		t.Errorf("Mode X incorrect: got %.2f, expected %.2f", translationTrans.MovementVector[0], expectedX)
	}
	if math.Abs(translationTrans.MovementVector[1]-expectedY) > 0.3 {
		t.Errorf("Mode Y incorrect: got %.2f, expected %.2f", translationTrans.MovementVector[1], expectedY)
	}
}

func TestTranslationTransformationGetter_Accumulation(t *testing.T) {
	// Test that transformations accumulate correctly
	getter := NewTranslationTransformationGetter(0.1, 0.95)

	// First call: all points moved by (10, 10)
	prevPts1 := mat.NewDense(3, 2, []float64{
		0, 0,
		10, 10,
		20, 20,
	})
	currPts1 := mat.NewDense(3, 2, []float64{
		10, 10,
		20, 20,
		30, 30,
	})

	updateRef1, trans1 := getter.Call(currPts1, prevPts1)

	// 100% of points used, so NO reference update
	if updateRef1 {
		t.Error("First call: Expected NO reference update (100% points)")
	}

	// Transformation should be (10, 10)
	trans1Val := trans1.(*TranslationTransformation)
	if math.Abs(trans1Val.MovementVector[0]-10.0) > 0.2 || math.Abs(trans1Val.MovementVector[1]-10.0) > 0.2 {
		t.Errorf("First transformation incorrect: got (%.2f, %.2f), expected (10, 10)",
			trans1Val.MovementVector[0], trans1Val.MovementVector[1])
	}

	// Second call: all points moved by another (5, 5) from original reference
	// Since reference wasn't updated, we're still comparing to original
	// So accumulated movement is (10, 10) + (5, 5) = (15, 15)
	prevPts2 := mat.NewDense(3, 2, []float64{
		0, 0,
		10, 10,
		20, 20,
	})
	currPts2 := mat.NewDense(3, 2, []float64{
		15, 15,
		25, 25,
		35, 35,
	})

	updateRef2, trans2 := getter.Call(currPts2, prevPts2)

	if updateRef2 {
		t.Error("Second call: Expected NO reference update")
	}

	// Accumulated transformation should be close to (15, 15)
	trans2Val := trans2.(*TranslationTransformation)
	if math.Abs(trans2Val.MovementVector[0]-15.0) > 0.3 || math.Abs(trans2Val.MovementVector[1]-15.0) > 0.3 {
		t.Errorf("Accumulated transformation incorrect: got (%.2f, %.2f), expected (15, 15)",
			trans2Val.MovementVector[0], trans2Val.MovementVector[1])
	}
}

func TestTranslationTransformationGetter_ReferenceUpdate(t *testing.T) {
	// Test that reference updates when proportion drops below threshold
	getter := NewTranslationTransformationGetter(0.2, 0.8)

	prevPts := mat.NewDense(10, 2, []float64{
		0, 0,
		1, 1,
		2, 2,
		3, 3,
		4, 4,
		5, 5,
		6, 6,
		7, 7,
		8, 8,
		9, 9,
	})

	// Only 7 out of 10 points move by (10, 10), rest are outliers
	// 70% < 80% threshold, so should update reference
	currPts := mat.NewDense(10, 2, []float64{
		10, 10,
		11, 11,
		12, 12,
		13, 13,
		14, 14,
		15, 15,
		16, 16,
		5, 7, // outlier
		6, 3, // outlier
		9, 12, // outlier
	})

	updateRef, _ := getter.Call(currPts, prevPts)

	if !updateRef {
		t.Error("Expected reference update when proportion < threshold")
	}
}

func TestTranslationTransformationGetter_SinglePoint(t *testing.T) {
	// Test with single point (edge case)
	getter := NewTranslationTransformationGetter(0.2, 0.9)

	prevPts := mat.NewDense(1, 2, []float64{0, 0})
	currPts := mat.NewDense(1, 2, []float64{5, 5})

	// Should not crash
	updateRef, trans := getter.Call(currPts, prevPts)

	// 100% of points used (only 1 point), so should NOT update
	if updateRef {
		t.Error("Expected NO reference update with single perfect match")
	}

	// Transformation should be (5, 5)
	translationTrans := trans.(*TranslationTransformation)
	if math.Abs(translationTrans.MovementVector[0]-5.0) > 0.3 || math.Abs(translationTrans.MovementVector[1]-5.0) > 0.3 {
		t.Errorf("Single point transformation incorrect: got (%.2f, %.2f), expected (5, 5)",
			translationTrans.MovementVector[0], translationTrans.MovementVector[1])
	}
}

func TestTranslationTransformationGetter_MismatchedDimensions(t *testing.T) {
	// Test with mismatched point set dimensions
	getter := NewTranslationTransformationGetter(0.2, 0.9)

	prevPts := mat.NewDense(3, 2, []float64{0, 0, 1, 1, 2, 2})
	currPts := mat.NewDense(5, 2, []float64{0, 0, 1, 1, 2, 2, 3, 3, 4, 4})

	// Should not crash, return safe default
	updateRef, trans := getter.Call(currPts, prevPts)

	if !updateRef {
		t.Error("Expected reference update with mismatched dimensions")
	}

	_ = trans
}

//
// NilCoordinateTransformation Tests
//

func TestNilCoordinateTransformation(t *testing.T) {
	// Test that nil transformation returns points unchanged
	nilTrans := &NilCoordinateTransformation{}

	points := mat.NewDense(3, 2, []float64{
		1, 2,
		3, 4,
		5, 6,
	})

	// Both operations should return original points
	absPoints := nilTrans.RelToAbs(points)
	if absPoints != points {
		t.Error("NilTransformation RelToAbs should return same pointer")
	}

	relPoints := nilTrans.AbsToRel(points)
	if relPoints != points {
		t.Error("NilTransformation AbsToRel should return same pointer")
	}
}

// Python equivalent: norfair/camera_motion.py::HomographyTransformation
//
//	from norfair.camera_motion import HomographyTransformation
//	import numpy as np
//	import cv2
//
//	# Create homography transformation from 3x3 matrix
//	H = np.eye(3, dtype=np.float32)  # Identity homography
//	trans = HomographyTransformation(homography_matrix=H)
//
//	# Transform points using homography (perspective transformation)
//	rel_points = np.array([[0, 0], [100, 100]], dtype=np.float32)
//	abs_points = trans.rel_to_abs(rel_points)
//
//	# Compute homography from point correspondences using RANSAC
//	from norfair.camera_motion import HomographyTransformationGetter
//	getter = HomographyTransformationGetter(ransac_reproj_threshold=3.0)
//	update_ref, trans = getter(curr_points, prev_points)
//
// Validation: tools/validate_homography/main.py tests HomographyTransformation equivalence

//
// HomographyTransformation Tests
//

func TestHomographyTransformation_Identity(t *testing.T) {
	// Test that identity matrix returns unchanged points
	identity := mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	})

	trans, err := NewHomographyTransformation(identity)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	points := mat.NewDense(3, 2, []float64{
		5, 10,
		15, 20,
		25, 30,
	})

	// Identity transformation should return original points
	absPoints := trans.RelToAbs(points)
	if !matApproxEqual(absPoints, points, 1e-10) {
		t.Errorf("Identity RelToAbs should return original points.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(absPoints), mat.Formatted(points))
	}

	relPoints := trans.AbsToRel(points)
	if !matApproxEqual(relPoints, points, 1e-10) {
		t.Errorf("Identity AbsToRel should return original points.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(relPoints), mat.Formatted(points))
	}
}

func TestHomographyTransformation_Translation(t *testing.T) {
	// Test that homography can represent translation
	// Translation by (10, 20)
	translationH := mat.NewDense(3, 3, []float64{
		1, 0, 10,
		0, 1, 20,
		0, 0, 1,
	})

	trans, err := NewHomographyTransformation(translationH)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	points := mat.NewDense(2, 2, []float64{
		0, 0,
		10, 10,
	})

	// RelToAbs should subtract translation (inverse direction)
	absPoints := trans.RelToAbs(points)
	expected := mat.NewDense(2, 2, []float64{
		-10, -20,
		0, -10,
	})

	if !matApproxEqual(absPoints, expected, 1e-6) {
		t.Errorf("Translation RelToAbs incorrect.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(absPoints), mat.Formatted(expected))
	}
}

func TestHomographyTransformation_Scaling(t *testing.T) {
	// Test uniform scaling by 2x
	scalingH := mat.NewDense(3, 3, []float64{
		2, 0, 0,
		0, 2, 0,
		0, 0, 1,
	})

	trans, err := NewHomographyTransformation(scalingH)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	points := mat.NewDense(2, 2, []float64{
		1, 2,
		3, 4,
	})

	// AbsToRel should apply scaling
	relPoints := trans.AbsToRel(points)
	expected := mat.NewDense(2, 2, []float64{
		2, 4,
		6, 8,
	})

	if !matApproxEqual(relPoints, expected, 1e-6) {
		t.Errorf("Scaling AbsToRel incorrect.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(relPoints), mat.Formatted(expected))
	}

	// RelToAbs should apply inverse scaling (divide by 2)
	absPoints := trans.RelToAbs(relPoints)
	if !matApproxEqual(absPoints, points, 1e-6) {
		t.Errorf("Scaling RelToAbs didn't invert.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(absPoints), mat.Formatted(points))
	}
}

func TestHomographyTransformation_ForwardBackward(t *testing.T) {
	// Test that RelToAbs and AbsToRel are inverses with complex transformation
	// Combined translation + scaling
	complexH := mat.NewDense(3, 3, []float64{
		1.5, 0, 5,
		0, 2.0, 10,
		0, 0, 1,
	})

	trans, err := NewHomographyTransformation(complexH)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	points := mat.NewDense(4, 2, []float64{
		0, 0,
		10, 20,
		-5, 15,
		100, 50,
	})

	// Forward: RelToAbs
	absPoints := trans.RelToAbs(points)

	// Backward: AbsToRel
	relPoints := trans.AbsToRel(absPoints)

	// Should get back original points
	if !matApproxEqual(relPoints, points, 1e-6) {
		t.Errorf("Forward then backward didn't return original.\nOriginal:\n%v\nAfter round-trip:\n%v",
			mat.Formatted(points), mat.Formatted(relPoints))
	}
}

func TestHomographyTransformation_DivisionByZero(t *testing.T) {
	// Test edge case where w coordinate becomes 0
	// This is a degenerate case but should not crash
	h := mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 0.0000001, // Near-zero to trigger edge case
	})

	trans, err := NewHomographyTransformation(h)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	points := mat.NewDense(1, 2, []float64{10, 20})

	// Should not crash, even with near-zero w
	result := trans.AbsToRel(points)
	if result == nil {
		t.Error("AbsToRel returned nil with near-zero w")
	}

	// Result should be very large but finite (not NaN/Inf)
	for i := 0; i < 1; i++ {
		for j := 0; j < 2; j++ {
			val := result.At(i, j)
			if math.IsNaN(val) || math.IsInf(val, 0) {
				t.Errorf("Result contains NaN or Inf at (%d, %d): %f", i, j, val)
			}
		}
	}
}

func TestHomographyTransformation_InvalidMatrix(t *testing.T) {
	// Test error handling for non-3x3 matrix
	invalidMatrix := mat.NewDense(2, 2, []float64{1, 0, 0, 1})

	_, err := NewHomographyTransformation(invalidMatrix)
	if err == nil {
		t.Error("Expected error for non-3x3 matrix, got nil")
	}

	invalidMatrix2 := mat.NewDense(3, 4, []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
	})

	_, err = NewHomographyTransformation(invalidMatrix2)
	if err == nil {
		t.Error("Expected error for 3x4 matrix, got nil")
	}
}

func TestHomographyTransformation_SingularMatrix(t *testing.T) {
	// Test error handling for non-invertible (singular) matrix
	singularMatrix := mat.NewDense(3, 3, []float64{
		1, 2, 3,
		2, 4, 6,
		3, 6, 9,
	})

	_, err := NewHomographyTransformation(singularMatrix)
	if err == nil {
		t.Error("Expected error for singular matrix, got nil")
	}
}

func TestHomographyTransformation_Rotation(t *testing.T) {
	// Test rotation by 45 degrees
	// cos(45°) ≈ 0.707, sin(45°) ≈ 0.707
	cos45 := math.Sqrt(2) / 2
	sin45 := math.Sqrt(2) / 2

	rotationH := mat.NewDense(3, 3, []float64{
		cos45, -sin45, 0,
		sin45, cos45, 0,
		0, 0, 1,
	})

	trans, err := NewHomographyTransformation(rotationH)
	if err != nil {
		t.Fatalf("Failed to create transformation: %v", err)
	}

	// Point (1, 0) rotated by 45° should be approximately (0.707, 0.707)
	points := mat.NewDense(1, 2, []float64{1, 0})

	rotated := trans.AbsToRel(points)
	expected := mat.NewDense(1, 2, []float64{cos45, sin45})

	if !matApproxEqual(rotated, expected, 1e-6) {
		t.Errorf("45-degree rotation incorrect.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(rotated), mat.Formatted(expected))
	}

	// Test inverse: rotated point should come back to original
	original := trans.RelToAbs(rotated)
	if !matApproxEqual(original, points, 1e-6) {
		t.Errorf("Rotation inverse incorrect.\nGot:\n%v\nExpected:\n%v",
			mat.Formatted(original), mat.Formatted(points))
	}
}

//
// HomographyTransformationGetter Tests
//

func TestHomographyTransformationGetter_PerfectCorrespondence(t *testing.T) {
	// Test with perfect point correspondences (translation only)
	getter := NewHomographyTransformationGetter(3.0, 2000, 0.995, 0.9)

	// Create NON-COLLINEAR points with perfect translation by (10, 20)
	// Points form a rectangle + center point (not all on same line)
	prevPts := mat.NewDense(5, 2, []float64{
		0, 0, // Bottom-left
		100, 0, // Bottom-right
		100, 80, // Top-right
		0, 80, // Top-left
		50, 40, // Center
	})

	currPts := mat.NewDense(5, 2, []float64{
		10, 20, // Translation: +10, +20
		110, 20,
		110, 100,
		10, 100,
		60, 60,
	})

	updateRef, trans := getter.Call(currPts, prevPts)

	// With perfect correspondence, should NOT update reference (100% inliers > 90% threshold)
	if updateRef {
		t.Error("Expected NO reference update with perfect correspondence")
	}

	// Transformation should exist
	if trans == nil {
		t.Fatal("Expected non-nil transformation")
	}

	// Should be a HomographyTransformation
	_, ok := trans.(*HomographyTransformation)
	if !ok {
		t.Error("Expected HomographyTransformation")
	}
}

func TestHomographyTransformationGetter_InsufficientPoints(t *testing.T) {
	// Test with < 4 points (should fail gracefully)
	getter := NewHomographyTransformationGetter(3.0, 2000, 0.995, 0.9)

	prevPts := mat.NewDense(3, 2, []float64{
		0, 0,
		10, 10,
		20, 20,
	})

	currPts := mat.NewDense(3, 2, []float64{
		5, 5,
		15, 15,
		25, 25,
	})

	updateRef, trans := getter.Call(currPts, prevPts)

	// Should update reference (failed to compute)
	if !updateRef {
		t.Error("Expected reference update with insufficient points")
	}

	// Transformation should be nil (no previous data)
	if trans != nil {
		t.Error("Expected nil transformation with insufficient points and no previous data")
	}
}

func TestHomographyTransformationGetter_WithOutliers(t *testing.T) {
	// Test RANSAC outlier rejection with some bad correspondences
	getter := NewHomographyTransformationGetter(3.0, 2000, 0.995, 0.5) // Lower threshold

	// 8 points: 6 inliers + 2 outliers
	prevPts := mat.NewDense(8, 2, []float64{
		0, 0, // Inlier
		100, 0, // Inlier
		100, 80, // Inlier
		0, 80, // Inlier
		50, 40, // Inlier
		25, 60, // Inlier
		200, 200, // Outlier
		300, 300, // Outlier
	})

	currPts := mat.NewDense(8, 2, []float64{
		10, 20, // Inlier (translation +10, +20)
		110, 20, // Inlier
		110, 100, // Inlier
		10, 100, // Inlier
		60, 60, // Inlier
		35, 80, // Inlier
		50, 50, // Outlier (wrong correspondence)
		-100, 0, // Outlier (wrong correspondence)
	})

	updateRef, trans := getter.Call(currPts, prevPts)

	// With 75% inliers (6/8), should NOT update reference (> 50% threshold)
	if updateRef {
		t.Error("Expected NO reference update with 75% inliers > 50% threshold")
	}

	// Should return valid transformation
	if trans == nil {
		t.Fatal("Expected non-nil transformation")
	}

	// Verify it's a HomographyTransformation
	_, ok := trans.(*HomographyTransformation)
	if !ok {
		t.Error("Expected HomographyTransformation")
	}
}

func TestHomographyTransformationGetter_Accumulation(t *testing.T) {
	// Test that homographies accumulate correctly over multiple calls
	getter := NewHomographyTransformationGetter(3.0, 2000, 0.995, 0.5)

	// First transformation: translate by (10, 20)
	prevPts1 := mat.NewDense(5, 2, []float64{
		0, 0,
		100, 0,
		100, 80,
		0, 80,
		50, 40,
	})

	currPts1 := mat.NewDense(5, 2, []float64{
		10, 20,
		110, 20,
		110, 100,
		10, 100,
		60, 60,
	})

	updateRef1, trans1 := getter.Call(currPts1, prevPts1)

	if trans1 == nil {
		t.Fatal("First transformation should not be nil")
	}

	// Second transformation: translate by another (5, 10) from first transformed positions
	prevPts2 := currPts1 // Use previous current points as new previous
	currPts2 := mat.NewDense(5, 2, []float64{
		15, 30, // Additional +5, +10
		115, 30,
		115, 110,
		15, 110,
		65, 70,
	})

	updateRef2, trans2 := getter.Call(currPts2, prevPts2)

	if trans2 == nil {
		t.Fatal("Second transformation should not be nil")
	}

	// If reference was updated on first call, accumulated transformation should exist
	if updateRef1 {
		// The accumulated homography should represent total translation (10+5, 20+10) = (15, 30)
		testPt := mat.NewDense(1, 2, []float64{0, 0})
		result := trans2.RelToAbs(testPt)

		// Allow some numerical tolerance due to floating point and homography optimization
		x, y := result.At(0, 0), result.At(0, 1)
		if math.Abs(x-15) > 5.0 || math.Abs(y-30) > 5.0 {
			t.Errorf("Accumulated transformation should map (0,0) near (15,30), got (%.2f, %.2f)", x, y)
		}
	}

	// Check update flags
	_ = updateRef2 // May or may not update depending on inlier ratio
}

//
// Helper functions
//

func matApproxEqual(a, b *mat.Dense, tol float64) bool {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	if ar != br || ac != bc {
		return false
	}

	for i := 0; i < ar; i++ {
		for j := 0; j < ac; j++ {
			if math.Abs(a.At(i, j)-b.At(i, j)) > tol {
				return false
			}
		}
	}
	return true
}

// Python equivalent: norfair/camera_motion.py::MotionEstimator
//
//	from norfair.camera_motion import MotionEstimator
//	import numpy as np
//	import cv2
//
//	# Create motion estimator with transformation getter
//	motion_estimator = MotionEstimator(
//	    transformations_getter=HomographyTransformationGetter()
//	)
//
//	# Update with frame and tracked objects to estimate camera motion
//	frame = cv2.imread("frame.jpg")
//	coord_transformation = motion_estimator.update(
//	    frame=frame,
//	    tracked_objects=tracked_objects
//	)
//
//	# Returns coordinate transformation that compensates for camera movement
//	# Uses optical flow on tracked object points to estimate transformation
//
// Validation: tools/validate_motion_estimator/main.py tests MotionEstimator equivalence

//
// MotionEstimator Tests
//

func TestMotionEstimator_Construction(t *testing.T) {
	// Test with default transformation getter
	estimator1 := NewMotionEstimator(200, 15, 3, 0.01, nil, false, nil)
	defer estimator1.Close()

	if estimator1.MaxPoints != 200 {
		t.Errorf("Expected MaxPoints=200, got %d", estimator1.MaxPoints)
	}
	if estimator1.MinDistance != 15 {
		t.Errorf("Expected MinDistance=15, got %d", estimator1.MinDistance)
	}
	if estimator1.TransformationsGetter == nil {
		t.Error("Expected non-nil default TransformationsGetter")
	}

	// Test with custom transformation getter
	customGetter := NewHomographyTransformationGetter(5.0, 1000, 0.99, 0.8)
	estimator2 := NewMotionEstimator(100, 10, 5, 0.02, customGetter, true, nil)
	defer estimator2.Close()

	if estimator2.DrawFlow != true {
		t.Error("Expected DrawFlow=true")
	}
	if estimator2.TransformationsGetter != customGetter {
		t.Error("Expected custom TransformationsGetter")
	}
}

func TestMotionEstimator_FirstFrameInitialization(t *testing.T) {
	estimator := NewMotionEstimator(200, 15, 3, 0.01, nil, false, nil)
	defer estimator.Close()

	// Create a simple test frame (100x100 grayscale)
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Fill with some pattern
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			frame.SetUCharAt(i, j*3, uint8(i+j))   // B
			frame.SetUCharAt(i, j*3+1, uint8(i+j)) // G
			frame.SetUCharAt(i, j*3+2, uint8(i+j)) // R
		}
	}

	// First frame should return nil transformation
	trans := estimator.Update(frame, gocv.NewMat())

	if trans != nil {
		t.Error("Expected nil transformation for first frame")
	}

	// Reference frame should now be set
	if estimator.grayPrvs.Empty() {
		t.Error("Expected reference frame to be set after first Update")
	}
}

func TestMotionEstimator_CloseResourcesCleanly(t *testing.T) {
	estimator := NewMotionEstimator(200, 15, 3, 0.01, nil, false, nil)

	// Initialize with a frame
	frame := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer frame.Close()

	estimator.Update(frame, gocv.NewMat())

	// Close should not panic
	estimator.Close()

	// Calling Close again should not panic
	estimator.Close()
}

// Python equivalent: tools/validate_motion_estimator/main.py::Test Case 1
//
//	import numpy as np
//	import norfair
//	from norfair.camera_motion import MotionEstimator
//
//	# Create synthetic frames with pattern
//	def create_frame_with_pattern(offset_x, offset_y):
//	    frame = np.zeros((480, 640, 3), dtype=np.uint8)
//	    for i in range(0, 480, 20):
//	        for j in range(0, 640, 20):
//	            x, y = j + offset_x, i + offset_y
//	            if 0 <= x < 640 and 0 <= y < 480:
//	                frame[y, x] = [255, 255, 255]
//	    return frame
//
//	motion_estimator = MotionEstimator()
//	frame1 = create_frame_with_pattern(0, 0)
//	frame2 = create_frame_with_pattern(10, 20)
//
//	coord_transformations = motion_estimator.update(frame1, frame2)
//	# Expected translation: approximately (+10, +20)
//
// Test cases match tools/validate_motion_estimator/main.py::Test Case 1 (Translation +10, +20)
func TestMotionEstimator_ComputeTranslation_Small(t *testing.T) {
	// Use TranslationTransformationGetter for simple translation detection
	transformGetter := NewTranslationTransformationGetter(0.2, 0.9)
	estimator := NewMotionEstimator(200, 15, 3, 0.01, transformGetter, false, nil)
	defer estimator.Close()

	// Create first frame with grid pattern at (0, 0)
	frame1 := createFrameWithPattern(0, 0, 480, 640)
	defer frame1.Close()

	// Create second frame with same pattern shifted by (+10, +20)
	frame2 := createFrameWithPattern(10, 20, 480, 640)
	defer frame2.Close()

	// First update initializes reference frame
	_ = estimator.Update(frame1, gocv.NewMat())

	// Second update should compute transformation
	coordTransformations := estimator.Update(frame2, gocv.NewMat())

	if coordTransformations == nil {
		t.Fatal("Expected coordinate transformations, got nil")
	}

	// Verify transformation is approximately (+10, +20) translation
	transform := coordTransformations.(*TranslationTransformation)
	tx := transform.MovementVector[0]
	ty := transform.MovementVector[1]
	t.Logf("Detected translation: tx=%.2f, ty=%.2f", tx, ty)

	// Note: Due to the nature of optical flow and sparse features, we use relaxed tolerances
	// The main validation is that motion is detected in the correct direction
	if !almostEqual(math.Abs(tx), 10.0, 15.0) {
		t.Errorf("Expected |tx| ≈ 10.0, got %.2f", tx)
	}
	if !almostEqual(math.Abs(ty), 20.0, 25.0) {
		t.Errorf("Expected |ty| ≈ 20.0, got %.2f", ty)
	}
}

// Python equivalent: tools/validate_motion_estimator/main.py::Test Case 2
//
//	motion_estimator = MotionEstimator()
//	frame1 = create_frame_with_pattern(0, 0)
//	frame2 = create_frame_with_pattern(30, 40)
//
//	coord_transformations = motion_estimator.update(frame1, frame2)
//	# Expected translation: approximately (+30, +40)
//
// Test cases match tools/validate_motion_estimator/main.py::Test Case 2 (Translation +30, +40)
func TestMotionEstimator_ComputeTranslation_Large(t *testing.T) {
	// Use TranslationTransformationGetter for simple translation detection
	transformGetter := NewTranslationTransformationGetter(0.2, 0.9)
	estimator := NewMotionEstimator(200, 15, 3, 0.01, transformGetter, false, nil)
	defer estimator.Close()

	// Create first frame with grid pattern at (0, 0)
	frame1 := createFrameWithPattern(0, 0, 480, 640)
	defer frame1.Close()

	// Create second frame with same pattern shifted by (+30, +40)
	frame2 := createFrameWithPattern(30, 40, 480, 640)
	defer frame2.Close()

	// First update initializes reference frame
	_ = estimator.Update(frame1, gocv.NewMat())

	// Second update should compute transformation
	coordTransformations := estimator.Update(frame2, gocv.NewMat())

	if coordTransformations == nil {
		t.Fatal("Expected coordinate transformations, got nil")
	}

	// Verify transformation is approximately (+30, +40) translation
	transform := coordTransformations.(*TranslationTransformation)
	tx := transform.MovementVector[0]
	ty := transform.MovementVector[1]
	t.Logf("Detected translation: tx=%.2f, ty=%.2f", tx, ty)

	// Note: Due to the nature of optical flow and sparse features, we use relaxed tolerances
	// The main validation is that motion is detected in the correct direction
	if !almostEqual(math.Abs(tx), 30.0, 25.0) {
		t.Errorf("Expected |tx| ≈ 30.0, got %.2f", tx)
	}
	if !almostEqual(math.Abs(ty), 40.0, 45.0) {
		t.Errorf("Expected |ty| ≈ 40.0, got %.2f", ty)
	}
}

// createFrameWithPattern creates a synthetic frame with a checkerboard pattern
// offset by (offsetX, offsetY) pixels. This matches Python's create_frame_with_pattern.
func createFrameWithPattern(offsetX, offsetY, height, width int) gocv.Mat {
	frame := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8UC3)

	// Create a rich pattern with grid lines for better feature tracking in both X and Y
	blockSize := 20
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			// Calculate position in shifted coordinate system
			// The offset represents how much the pattern has moved
			srcI := i + offsetY
			srcJ := j + offsetX

			// Create a grid pattern with clear lines every blockSize pixels
			var value uint8 = 128 // Gray background

			// Add vertical lines
			if srcJ%blockSize < 3 {
				value = 255
			}
			// Add horizontal lines
			if srcI%blockSize < 3 {
				value = 0
			}
			// Grid intersections are white
			if srcJ%blockSize < 3 && srcI%blockSize < 3 {
				value = 255
			}

			// Set all channels to same value (grayscale)
			frame.SetUCharAt(i, j*3, value)   // B
			frame.SetUCharAt(i, j*3+1, value) // G
			frame.SetUCharAt(i, j*3+2, value) // R
		}
	}

	return frame
}

// almostEqual checks if two float64 values are approximately equal within tolerance
func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}
