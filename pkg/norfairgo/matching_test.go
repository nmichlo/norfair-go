package norfairgo

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Helper Functions for Tests
// =============================================================================

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// =============================================================================
// Test Perfect Matches
// =============================================================================

// Python equivalent: tools/validate_matching/main.py::test_perfect_matches()
//
//	def greedy_match(distance_matrix, distance_threshold):
//	    # Greedy matching from norfair tracker
//	    matched_det_indices = []
//	    matched_obj_indices = []
//	    matrix_copy = distance_matrix.copy()
//	    current_min = matrix_copy.min()
//	    while current_min < distance_threshold:
//	        det_idx, obj_idx = np.unravel_index(matrix_copy.argmin(), matrix_copy.shape)
//	        matched_det_indices.append(det_idx)
//	        matched_obj_indices.append(obj_idx)
//	        matrix_copy[det_idx, :] = distance_threshold + 1.0
//	        matrix_copy[:, obj_idx] = distance_threshold + 1.0
//	        current_min = matrix_copy.min()
//	    return matched_det_indices, matched_obj_indices
//
//	def test_perfect_matches():
//	    distance_matrix = np.array([[0.5, 0.9, 0.8], [0.9, 0.3, 0.7], [0.8, 0.7, 0.4]])
//	    threshold = 1.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    # Greedy order: [1,1]=0.3, [2,2]=0.4, [0,0]=0.5
//	    assert cand_indices == [1, 2, 0]
//	    assert obj_indices == [1, 2, 0]
//
// Test cases match tools/validate_matching/main.py::test_perfect_matches()
func TestMatching_PerfectMatches(t *testing.T) {
	// All distances below threshold - all should match
	distanceMatrix := mat.NewDense(3, 3, []float64{
		0.5, 0.9, 0.8,
		0.9, 0.3, 0.7,
		0.8, 0.7, 0.4,
	})

	threshold := 1.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should match all 3 pairs (greedy picks minimums)
	if len(candIndices) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(candIndices))
	}
	if len(objIndices) != 3 {
		t.Errorf("Expected 3 object indices, got %d", len(objIndices))
	}

	// Verify greedy behavior: picks minimum first (0.3 at [1,1])
	// Then next minimum from remaining, etc.
	// Expected order: [1,1]=0.3, [2,2]=0.4, [0,0]=0.5
	expectedCandIndices := []int{1, 2, 0}
	expectedObjIndices := []int{1, 2, 0}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}
}

// =============================================================================
// Test Threshold Filtering
// =============================================================================

// Python equivalent: tools/validate_matching/main.py::test_threshold_filtering()
//
//	def test_threshold_filtering():
//	    distance_matrix = np.array([[0.5, 2.0, 3.0], [2.5, 0.8, 2.0], [3.0, 3.0, 0.3]])
//	    threshold = 1.5
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    # Greedy order: [2,2]=0.3, [0,0]=0.5, [1,1]=0.8
//	    assert cand_indices == [2, 0, 1]
func TestMatching_ThresholdFiltering(t *testing.T) {
	// Some distances above threshold, some below
	distanceMatrix := mat.NewDense(3, 3, []float64{
		0.5, 2.0, 3.0, // Only [0,0]=0.5 is below threshold
		2.5, 0.8, 2.0, // Only [1,1]=0.8 is below threshold
		3.0, 3.0, 0.3, // Only [2,2]=0.3 is below threshold
	})

	threshold := 1.5

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should match only the 3 values below threshold
	// Greedy order: [2,2]=0.3, [0,0]=0.5, [1,1]=0.8
	expectedCandIndices := []int{2, 0, 1}
	expectedObjIndices := []int{2, 0, 1}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}
}

// Python equivalent: tools/validate_matching/main.py::test_all_above_threshold()
//
//	def test_all_above_threshold():
//	    distance_matrix = np.array([[5.0, 6.0], [7.0, 8.0]])
//	    threshold = 3.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    assert len(cand_indices) == 0
func TestMatching_AllAboveThreshold(t *testing.T) {
	// All distances above threshold - no matches
	distanceMatrix := mat.NewDense(2, 2, []float64{
		5.0, 6.0,
		7.0, 8.0,
	})

	threshold := 3.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should have no matches
	if len(candIndices) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(candIndices))
	}
	if len(objIndices) != 0 {
		t.Errorf("Expected 0 object indices, got %d", len(objIndices))
	}
}

// =============================================================================
// Test Empty/Minimal Inputs
// =============================================================================

// Python equivalent: tools/validate_matching/main.py::test_single_element()
//
//	def test_single_element():
//	    # Case 2: Above threshold
//	    distance_matrix = np.array([[5.0]])
//	    threshold = 3.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    assert cand_indices == []
func TestMatching_SingleElementNoMatch(t *testing.T) {
	// 1x1 matrix with distance above threshold - no match
	distanceMatrix := mat.NewDense(1, 1, []float64{5.0})
	threshold := 3.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	if len(candIndices) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(candIndices))
	}
	if len(objIndices) != 0 {
		t.Errorf("Expected 0 object indices, got %d", len(objIndices))
	}
}

// Python equivalent: tools/validate_matching/main.py::test_single_element()
//
//	def test_single_element():
//	    # Case 1: Below threshold
//	    distance_matrix = np.array([[0.5]])
//	    threshold = 1.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    assert cand_indices == [0]
func TestMatching_SingleElementMatch(t *testing.T) {
	// 1x1 matrix with distance below threshold - should match
	distanceMatrix := mat.NewDense(1, 1, []float64{0.5})
	threshold := 1.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	expectedCandIndices := []int{0}
	expectedObjIndices := []int{0}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}
}

// =============================================================================
// Test Greedy vs Optimal Behavior
// =============================================================================

func TestMatching_GreedyBehavior(t *testing.T) {
	// Matrix where greedy != globally optimal assignment
	// Greedy will pick [0,0]=1.0 first, then [1,1]=1.0
	// Optimal Hungarian would also pick these, but this tests greedy behavior
	distanceMatrix := mat.NewDense(2, 2, []float64{
		1.0, 2.0,
		2.0, 1.0,
	})

	threshold := 3.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should match both (greedy picks first minimum at [0,0] or [1,1])
	if len(candIndices) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(candIndices))
	}

	// Verify one-to-one mapping (no duplicates)
	candMap := make(map[int]bool)
	objMap := make(map[int]bool)
	for i := range candIndices {
		if candMap[candIndices[i]] {
			t.Errorf("Duplicate candidate index: %d", candIndices[i])
		}
		if objMap[objIndices[i]] {
			t.Errorf("Duplicate object index: %d", objIndices[i])
		}
		candMap[candIndices[i]] = true
		objMap[objIndices[i]] = true
	}
}

// =============================================================================
// Test One-to-One Constraint
// =============================================================================

// Python equivalent: tools/validate_matching/main.py::test_one_to_one_constraint()
//
//	def test_one_to_one_constraint():
//	    distance_matrix = np.array([[0.5, 3.0], [0.6, 3.5], [0.7, 2.0]])
//	    threshold = 4.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    # Greedy picks: [0,0]=0.5, then [2,1]=2.0
//	    assert cand_indices == [0, 2]
//	    assert obj_indices == [0, 1]
func TestMatching_OneToOneConstraint(t *testing.T) {
	// Multiple candidates close to same object
	// Only one should match
	distanceMatrix := mat.NewDense(3, 2, []float64{
		0.5, 3.0, // Cand 0 closest to Obj 0
		0.6, 3.5, // Cand 1 also close to Obj 0
		0.7, 2.0, // Cand 2 closest to Obj 1
	})

	threshold := 4.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should have 2 matches (one per object)
	if len(candIndices) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(candIndices))
	}

	// Greedy picks: [0,0]=0.5 first, then [2,1]=2.0
	// Cand 1 is left unmatched because Obj 0 is already taken
	expectedCandIndices := []int{0, 2}
	expectedObjIndices := []int{0, 1}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}
}

// =============================================================================
// Test Asymmetric Matrices
// =============================================================================

// Python equivalent: tools/validate_matching/main.py::test_asymmetric_more_detections()
//
//	def test_asymmetric_more_detections():
//	    distance_matrix = np.array([[0.5, 2.0, 3.0], [0.8, 0.4, 2.5], [1.2, 1.5, 0.3], [2.0, 2.5, 1.8], [3.0, 3.5, 2.2]])
//	    threshold = 2.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    # Greedy order: [2,2]=0.3, [1,1]=0.4, [0,0]=0.5
//	    assert cand_indices == [2, 1, 0]
func TestMatching_MoreDetectionsThanObjects(t *testing.T) {
	// 5 detections, 3 objects
	distanceMatrix := mat.NewDense(5, 3, []float64{
		0.5, 2.0, 3.0,
		0.8, 0.4, 2.5,
		1.2, 1.5, 0.3,
		2.0, 2.5, 1.8,
		3.0, 3.5, 2.2,
	})

	threshold := 2.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should match at most 3 (limited by number of objects)
	if len(candIndices) > 3 {
		t.Errorf("Expected at most 3 matches, got %d", len(candIndices))
	}

	// Greedy picks: [2,2]=0.3, [1,1]=0.4, [0,0]=0.5
	expectedCandIndices := []int{2, 1, 0}
	expectedObjIndices := []int{2, 1, 0}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}
}

// Python equivalent: tools/validate_matching/main.py::test_asymmetric_more_objects()
//
//	def test_asymmetric_more_objects():
//	    distance_matrix = np.array([[0.5, 2.0, 1.5, 3.0], [1.8, 0.6, 2.5, 2.2]])
//	    threshold = 2.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    # Greedy order: [0,0]=0.5, [1,1]=0.6
//	    assert cand_indices == [0, 1]
func TestMatching_MoreObjectsThanDetections(t *testing.T) {
	// 2 detections, 4 objects
	distanceMatrix := mat.NewDense(2, 4, []float64{
		0.5, 2.0, 1.5, 3.0,
		1.8, 0.6, 2.5, 2.2,
	})

	threshold := 2.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should match at most 2 (limited by number of detections)
	if len(candIndices) > 2 {
		t.Errorf("Expected at most 2 matches, got %d", len(candIndices))
	}

	// Greedy picks: [0,0]=0.5, [1,1]=0.6
	expectedCandIndices := []int{0, 1}
	expectedObjIndices := []int{0, 1}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}
}

// =============================================================================
// Test NaN Detection
// =============================================================================

func TestMatching_NaNDetection(t *testing.T) {
	// Matrix with NaN value
	distanceMatrix := mat.NewDense(2, 2, []float64{
		0.5, math.NaN(),
		1.0, 0.8,
	})

	// Test hasNaN function
	if !hasNaN(distanceMatrix) {
		t.Error("hasNaN should return true for matrix with NaN")
	}

	// Test ValidateDistanceMatrix
	err := ValidateDistanceMatrix(distanceMatrix)
	if err == nil {
		t.Error("ValidateDistanceMatrix should return error for matrix with NaN")
	}
}

func TestMatching_NoNaN(t *testing.T) {
	// Matrix without NaN
	distanceMatrix := mat.NewDense(2, 2, []float64{
		0.5, 1.0,
		1.5, 0.8,
	})

	// Test hasNaN function
	if hasNaN(distanceMatrix) {
		t.Error("hasNaN should return false for matrix without NaN")
	}

	// Test ValidateDistanceMatrix
	err := ValidateDistanceMatrix(distanceMatrix)
	if err != nil {
		t.Errorf("ValidateDistanceMatrix should not return error for valid matrix: %v", err)
	}
}

// =============================================================================
// Test Inf Handling
// =============================================================================

// Python equivalent: tools/validate_matching/main.py::test_inf_handling()
//
//	def test_inf_handling():
//	    distance_matrix = np.array([[0.5, np.inf, np.inf], [np.inf, 0.8, np.inf], [np.inf, np.inf, 0.3]])
//	    threshold = 1.0
//	    cand_indices, obj_indices = greedy_match(distance_matrix, threshold)
//	    # Greedy order: [2,2]=0.3, [0,0]=0.5, [1,1]=0.8
//	    assert cand_indices == [2, 0, 1]
func TestMatching_InfHandling(t *testing.T) {
	// Matrix with Inf values (from label mismatches in distance functions)
	distanceMatrix := mat.NewDense(3, 3, []float64{
		0.5, math.Inf(1), math.Inf(1),
		math.Inf(1), 0.8, math.Inf(1),
		math.Inf(1), math.Inf(1), 0.3,
	})

	threshold := 1.0

	candIndices, objIndices := MatchDetectionsAndObjects(distanceMatrix, threshold)

	// Should match the 3 finite values below threshold
	// Greedy order: [2,2]=0.3, [0,0]=0.5, [1,1]=0.8
	expectedCandIndices := []int{2, 0, 1}
	expectedObjIndices := []int{2, 0, 1}

	if !slicesEqual(candIndices, expectedCandIndices) {
		t.Errorf("Expected cand indices %v, got %v", expectedCandIndices, candIndices)
	}
	if !slicesEqual(objIndices, expectedObjIndices) {
		t.Errorf("Expected obj indices %v, got %v", expectedObjIndices, objIndices)
	}

	// Verify Inf doesn't cause NaN error
	err := ValidateDistanceMatrix(distanceMatrix)
	if err != nil {
		t.Errorf("Matrix with Inf should not cause validation error: %v", err)
	}
}

// =============================================================================
// Test Helper Functions
// =============================================================================

func TestArgMin(t *testing.T) {
	m := mat.NewDense(3, 3, []float64{
		5.0, 3.0, 7.0,
		2.0, 9.0, 4.0,
		6.0, 1.0, 8.0,
	})

	minIdx := argMin(m)
	expectedIdx := 7 // Row 2, Col 1 (value=1.0), flattened index = 2*3 + 1 = 7

	if minIdx != expectedIdx {
		t.Errorf("argMin expected index %d, got %d", expectedIdx, minIdx)
	}
}

func TestMinMatrix(t *testing.T) {
	m := mat.NewDense(3, 3, []float64{
		5.0, 3.0, 7.0,
		2.0, 9.0, 4.0,
		6.0, 1.0, 8.0,
	})

	minVal := minMatrix(m)
	expectedVal := 1.0

	if minVal != expectedVal {
		t.Errorf("minMatrix expected %f, got %f", expectedVal, minVal)
	}
}
