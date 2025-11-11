package norfairgo

import (
	"math"
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
	"gonum.org/v1/gonum/mat"
)

// Test helper functions are now in internal/testutil/testutil.go

// =============================================================================
// Mock Detection and Tracked Object Creators
// =============================================================================
func newMockDetection(points [][]float64) *Detection {
	rows := len(points)
	cols := len(points[0])
	flat := make([]float64, rows*cols)
	for i := range points {
		for j := range points[i] {
			flat[i*cols+j] = points[i][j]
		}
	}
	return &Detection{
		Points: mat.NewDense(rows, cols, flat),
		Scores: make([]float64, rows), // Default to zeros
	}
}

func newMockDetectionWithScores(points [][]float64, scores interface{}) *Detection {
	rows := len(points)
	cols := len(points[0])
	flat := make([]float64, rows*cols)
	for i := range points {
		for j := range points[i] {
			flat[i*cols+j] = points[i][j]
		}
	}

	// Handle scores - can be a single float or slice
	var scoreSlice []float64
	switch s := scores.(type) {
	case float64:
		scoreSlice = make([]float64, rows)
		for i := range scoreSlice {
			scoreSlice[i] = s
		}
	case []float64:
		scoreSlice = s
	default:
		scoreSlice = make([]float64, rows)
	}

	return &Detection{
		Points: mat.NewDense(rows, cols, flat),
		Scores: scoreSlice,
	}
}

func newMockTrackedObject(points [][]float64) *TrackedObject {
	rows := len(points)
	cols := len(points[0])
	flat := make([]float64, rows*cols)
	for i := range points {
		for j := range points[i] {
			flat[i*cols+j] = points[i][j]
		}
	}
	return &TrackedObject{
		Estimate: mat.NewDense(rows, cols, flat),
	}
}

func newMockTrackedObjectWithScores(points [][]float64, scores interface{}) *TrackedObject {
	rows := len(points)
	cols := len(points[0])
	flat := make([]float64, rows*cols)
	for i := range points {
		for j := range points[i] {
			flat[i*cols+j] = points[i][j]
		}
	}

	// Handle scores - can be a single float or slice
	var scoreSlice []float64
	switch s := scores.(type) {
	case float64:
		scoreSlice = make([]float64, rows)
		for i := range scoreSlice {
			scoreSlice[i] = s
		}
	case []float64:
		scoreSlice = s
	default:
		scoreSlice = make([]float64, rows)
	}

	lastDet := &Detection{
		Points: mat.NewDense(rows, cols, flat),
		Scores: scoreSlice,
	}

	return &TrackedObject{
		Estimate:      mat.NewDense(rows, cols, flat),
		LastDetection: lastDet,
	}
}

// =============================================================================
// Test Frobenius Distance
// =============================================================================

// Python equivalent: norfair/distances.py::frobenius()
//
//	from norfair.distances import frobenius
//
//	def test_frobenius():
//	    det = MockDetection([[1, 2], [3, 4]])
//	    obj = MockTrackedObject([[1, 2], [3, 4]])
//	    result = frobenius(det, obj)
//	    assert abs(result - 0.0) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_frobenius()
func TestFrobenius(t *testing.T) {
	tests := []struct {
		name     string
		det      [][]float64
		obj      [][]float64
		expected float64
	}{
		{
			name:     "perfect match",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{1, 2}, {3, 4}},
			expected: 0.0,
		},
		{
			name:     "perfect match floats",
			det:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			obj:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			expected: 0.0,
		},
		{
			name:     "distance 1 in 1D of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 2}, {3, 4}},
			expected: 1.0,
		},
		{
			name:     "distance 2 in 1D of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 2}, {3, 4}},
			expected: 2.0,
		},
		{
			name:     "distance 1 in all dims",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 3}, {4, 5}},
			expected: 2.0, // sqrt(1+1+1+1) = sqrt(4) = 2.0
		},
		{
			name:     "negative difference",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 2.0,
		},
		{
			name:     "negative equals",
			det:      [][]float64{{-1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := newMockDetection(tt.det)
			obj := newMockTrackedObject(tt.obj)
			result := Frobenius(det, obj)
			testutil.AssertAlmostEqual(t, result, tt.expected, 1e-6, tt.name)
		})
	}
}

// =============================================================================
// Test Mean Manhattan Distance
// =============================================================================

// Python equivalent: norfair/distances.py::mean_manhattan()
//
//	from norfair.distances import mean_manhattan
//
//	def test_mean_manhattan():
//	    det = MockDetection([[1, 2], [3, 4]])
//	    obj = MockTrackedObject([[2, 2], [3, 4]])
//	    result = mean_manhattan(det, obj)
//	    # Returns mean of L1 distances per point
//	    assert abs(result - 0.5) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_mean_manhattan()
func TestMeanManhattan(t *testing.T) {
	tests := []struct {
		name     string
		det      [][]float64
		obj      [][]float64
		expected float64
	}{
		{
			name:     "perfect match",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{1, 2}, {3, 4}},
			expected: 0.0,
		},
		{
			name:     "perfect match floats",
			det:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			obj:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			expected: 0.0,
		},
		{
			name:     "distance 1 in 1D of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 2}, {3, 4}},
			expected: 0.5, // (1+0) / 2 points = 0.5
		},
		{
			name:     "distance 2 in 1D of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 2}, {3, 4}},
			expected: 1.0, // (2+0) / 2 = 1.0
		},
		{
			name:     "distance 1 in all dims",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 3}, {4, 5}},
			expected: 2.0, // (2+2) / 2 = 2.0
		},
		{
			name:     "negative difference",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 1.0, // (2+0) / 2 = 1.0
		},
		{
			name:     "negative equals",
			det:      [][]float64{{-1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := newMockDetection(tt.det)
			obj := newMockTrackedObject(tt.obj)
			result := MeanManhattan(det, obj)
			testutil.AssertAlmostEqual(t, result, tt.expected, 1e-6, tt.name)
		})
	}
}

// =============================================================================
// Test Mean Euclidean Distance
// =============================================================================

// Python equivalent: norfair/distances.py::mean_euclidean()
//
//	from norfair.distances import mean_euclidean
//
//	def test_mean_euclidean():
//	    det = MockDetection([[1, 2], [3, 4]])
//	    obj = MockTrackedObject([[2, 3], [4, 5]])
//	    result = mean_euclidean(det, obj)
//	    # Returns mean of L2 distances per point
//	    # Point 1: sqrt((2-1)^2 + (3-2)^2) = sqrt(2) ≈ 1.41421
//	    # Point 2: sqrt((4-3)^2 + (5-4)^2) = sqrt(2) ≈ 1.41421
//	    # Mean: sqrt(2) ≈ 1.41421
//	    assert abs(result - math.sqrt(2)) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_mean_euclidean()
func TestMeanEuclidean(t *testing.T) {
	tests := []struct {
		name     string
		det      [][]float64
		obj      [][]float64
		expected float64
	}{
		{
			name:     "perfect match",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{1, 2}, {3, 4}},
			expected: 0.0,
		},
		{
			name:     "perfect match floats",
			det:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			obj:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			expected: 0.0,
		},
		{
			name:     "distance 1 in 1D of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 2}, {3, 4}},
			expected: 0.5, // (1+0) / 2 = 0.5
		},
		{
			name:     "distance 2 in 1D of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 2}, {3, 4}},
			expected: 1.0, // (2+0) / 2 = 1.0
		},
		{
			name:     "distance 2 in 1D of all points",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 2}, {5, 4}},
			expected: 2.0, // (2+2) / 2 = 2.0
		},
		{
			name:     "distance 2 in all dims",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 4}, {5, 6}},
			expected: math.Sqrt(8), // (sqrt(8)+sqrt(8)) / 2 = sqrt(8)
		},
		{
			name:     "distance 1 in all dims",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 3}, {4, 5}},
			expected: math.Sqrt(2), // (sqrt(2)+sqrt(2)) / 2 = sqrt(2)
		},
		{
			name:     "negative difference",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 1.0, // (2+0) / 2 = 1.0
		},
		{
			name:     "negative equals",
			det:      [][]float64{{-1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := newMockDetection(tt.det)
			obj := newMockTrackedObject(tt.obj)
			result := MeanEuclidean(det, obj)
			testutil.AssertAlmostEqual(t, result, tt.expected, 1e-6, tt.name)
		})
	}
}

// =============================================================================
// Test IoU Distance
// =============================================================================

// Python equivalent: norfair/distances.py::iou()
//
//	from norfair.distances import iou
//
//	def test_iou():
//	    # Bounding boxes as [[x_min, y_min], [x_max, y_max]]
//	    det = MockDetection([[0, 0], [10, 10]])  # Box: 0,0 to 10,10 (area=100)
//	    obj = MockTrackedObject([[0, 0], [10, 10]])  # Same box
//	    result = iou(det, obj)
//	    # IoU = intersection / union = 100 / 100 = 1.0
//	    assert abs(result - 1.0) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_iou()
func TestIoU(t *testing.T) {
	tests := []struct {
		name     string
		cand     [][]float64
		obj      [][]float64
		expected float64
	}{
		{
			name:     "perfect match",
			cand:     [][]float64{{0, 0, 1, 1}},
			obj:      [][]float64{{0, 0, 1, 1}},
			expected: 0.0, // 1 - 1 = 0
		},
		{
			name:     "perfect match floats",
			cand:     [][]float64{{0, 0, 1.1, 1.1}},
			obj:      [][]float64{{0, 0, 1.1, 1.1}},
			expected: 0.0,
		},
		{
			name:     "detection contained in object",
			cand:     [][]float64{{0, 0, 1, 1}},
			obj:      [][]float64{{0, 0, 2, 2}},
			expected: 0.75, // 1 - (1/4) = 0.75
		},
		{
			name:     "no overlap",
			cand:     [][]float64{{0, 0, 1, 1}},
			obj:      [][]float64{{1, 1, 2, 2}},
			expected: 1.0, // 1 - 0 = 1.0
		},
		{
			name:     "object contained in detection",
			cand:     [][]float64{{0, 0, 4, 4}},
			obj:      [][]float64{{1, 1, 2, 2}},
			expected: 0.9375, // 1 - (1/16) = 0.9375
		},
		{
			name:     "partial overlap",
			cand:     [][]float64{{0, 0, 2, 2}},
			obj:      [][]float64{{1, 1, 3, 3}},
			expected: 1.0 - 1.0/7.0, // intersection=1, union=4+4-1=7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candMat := mat.NewDense(len(tt.cand), 4, nil)
			for i, row := range tt.cand {
				for j, val := range row {
					candMat.Set(i, j, val)
				}
			}

			objMat := mat.NewDense(len(tt.obj), 4, nil)
			for i, row := range tt.obj {
				for j, val := range row {
					objMat.Set(i, j, val)
				}
			}

			result := IoU(candMat, objMat)
			testutil.AssertAlmostEqual(t, result.At(0, 0), tt.expected, 1e-6, tt.name)
		})
	}
}

func TestIoU_InvalidBbox(t *testing.T) {
	// Test invalid bbox shape (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid bbox shape")
		}
	}()

	candMat := mat.NewDense(1, 2, []float64{0, 0}) // Only 2 columns, should be 4
	objMat := mat.NewDense(1, 4, []float64{0, 0, 1, 1})
	IoU(candMat, objMat)
}

// =============================================================================
// Test ScalarDistance Wrapper
// =============================================================================

// Python equivalent: norfair uses scalar distance functions directly with Distance wrapper
//
//	from norfair.distances import frobenius
//	from norfair import Distance
//
//	def test_scalar_distance():
//	    # In Python, scalar distance functions are wrapped by Distance class
//	    distance = Distance(frobenius)
//	    det = MockDetection([[1, 2], [3, 4]])
//	    obj = MockTrackedObject([[1, 2], [3, 4]])
//	    # Distance wrapper creates matrix of distances
//	    matrix = distance.get_distances([obj], [det])
//	    assert matrix.shape == (1, 1)
//	    assert abs(matrix[0, 0] - 0.0) < 1e-6
//
// Test cases validate the ScalarDistance wrapper behavior
func TestScalarDistance(t *testing.T) {
	// Test the ScalarDistance wrapper with frobenius
	distance := NewScalarDistance(Frobenius)

	det := newMockDetection([][]float64{{1, 2}, {3, 4}})
	obj := newMockTrackedObject([][]float64{{1, 2}, {3, 4}})

	detections := []*Detection{det}
	objects := []*TrackedObject{obj}

	matrix := distance.GetDistances(objects, detections)

	rows, cols := matrix.Dims()
	if rows != 1 || cols != 1 {
		t.Errorf("Expected matrix shape (1, 1), got (%d, %d)", rows, cols)
	}

	testutil.AssertAlmostEqual(t, matrix.At(0, 0), 0.0, 1e-6, "frobenius distance should be 0")
}

// =============================================================================
// Test VectorizedDistance Wrapper
// =============================================================================

// Python equivalent: norfair uses vectorized distance functions (like iou) with Distance wrapper
//
//	from norfair.distances import iou
//	from norfair import Distance
//
//	def test_vectorized_distance():
//	    # In Python, vectorized functions like iou are wrapped by Distance class
//	    distance = Distance(iou)
//	    det_bbox = np.array([[0, 0, 1, 1]], dtype=np.float64)
//	    obj_bbox = np.array([[0, 0, 1, 1]], dtype=np.float64)
//	    det = MockDetection(det_bbox)
//	    obj = MockTrackedObject(obj_bbox)
//	    # Distance wrapper creates matrix of distances
//	    matrix = distance.get_distances([obj], [det])
//	    assert matrix.shape == (1, 1)
//	    assert abs(matrix[0, 0] - 0.0) < 1e-6  # Perfect match = distance 0
//
// Test cases validate the VectorizedDistance wrapper behavior
func TestVectorizedDistance(t *testing.T) {
	// Test the VectorizedDistance wrapper with IoU
	distance := NewVectorizedDistance(IoU)

	// Create bboxes for detection and tracked object
	detBbox := mat.NewDense(1, 4, []float64{0, 0, 1, 1})
	objBbox := mat.NewDense(1, 4, []float64{0, 0, 1, 1})

	det := &Detection{Points: detBbox}
	obj := &TrackedObject{Estimate: objBbox}

	detections := []*Detection{det}
	objects := []*TrackedObject{obj}

	matrix := distance.GetDistances(objects, detections)

	rows, cols := matrix.Dims()
	if rows != 1 || cols != 1 {
		t.Errorf("Expected matrix shape (1, 1), got (%d, %d)", rows, cols)
	}

	testutil.AssertAlmostEqual(t, matrix.At(0, 0), 0.0, 1e-6, "IoU distance should be 0 for perfect match")
}

// =============================================================================
// Test ScipyDistance
// =============================================================================

// Python equivalent: norfair/distances.py::ScipyDistance
//
//	from norfair.distances import ScipyDistance
//	import numpy as np
//
//	def test_scipy_distance():
//	    # ScipyDistance uses scipy.spatial.distance.cdist with specified metric
//	    scipy_dist = ScipyDistance("euclidean")
//	    # Points are flattened: [[1, 2], [3, 4]] -> [1, 2, 3, 4]
//	    det = np.array([[1, 2, 3, 4]], dtype=np.float64)
//	    obj = np.array([[1, 2, 4, 4]], dtype=np.float64)
//	    result = scipy_dist.distance_function(det, obj)[0, 0]
//	    # euclidean distance = sqrt((1-1)^2 + (2-2)^2 + (3-4)^2 + (4-4)^2) = sqrt(1) = 1.0
//	    assert abs(result - 1.0) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_scipy_distance()
func TestScipyDistance(t *testing.T) {
	// Test ScipyDistance with euclidean metric
	euc := NewScipyDistance("euclidean")

	// det = [[1, 2], [3, 4]] flattened to [1, 2, 3, 4]
	// obj = [[1, 2], [4, 4]] flattened to [1, 2, 4, 4]
	// euclidean distance = sqrt((1-1)^2 + (2-2)^2 + (3-4)^2 + (4-4)^2) = sqrt(1) = 1.0
	det := newMockDetection([][]float64{{1, 2}, {3, 4}})
	obj := newMockTrackedObject([][]float64{{1, 2}, {4, 4}})

	detections := []*Detection{det}
	objects := []*TrackedObject{obj}

	distMatrix := euc.GetDistances(objects, detections)

	rows, cols := distMatrix.Dims()
	if rows != 1 || cols != 1 {
		t.Errorf("Expected matrix shape (1, 1), got (%d, %d)", rows, cols)
	}

	testutil.AssertAlmostEqual(t, distMatrix.At(0, 0), 1.0, 1e-6, "euclidean distance should be 1.0")
}

// =============================================================================
// Test Keypoint Voting Distance
// =============================================================================

// Python equivalent: norfair/distances.py::create_keypoints_voting_distance()
//
//	from norfair.distances import create_keypoints_voting_distance
//	import numpy as np
//
//	def test_keypoint_voting():
//	    vote_d = create_keypoints_voting_distance(
//	        keypoint_distance_threshold=np.sqrt(8),
//	        detection_threshold=0.5
//	    )
//	    det = MockDetection([[0, 0], [1, 1], [2, 2]], scores=0.6)
//	    obj = MockTrackedObject([[0, 0], [1, 1], [2, 2]], scores=0.6)
//	    result = vote_d(det, obj)
//	    # Returns: 1 - (matching_keypoints / total_keypoints)
//	    # 3 points match -> 3 matching -> (1 - 3/4) = 1/4
//	    assert abs(result - 1.0/4.0) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_keypoint_voting()
func TestKeypointVote(t *testing.T) {
	voteD := CreateKeypointsVotingDistance(math.Sqrt(8), 0.5)

	tests := []struct {
		name      string
		detPoints [][]float64
		detScores interface{}
		objPoints [][]float64
		objScores interface{}
		expected  float64
	}{
		{
			name:      "perfect match",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			detScores: 0.6,
			objPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			objScores: 0.6,
			expected:  1.0 / 4.0, // 3 matches
		},
		{
			name:      "just under distance threshold",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2.0}},
			detScores: 0.6,
			objPoints: [][]float64{{0, 0}, {1, 1}, {4, 3.9}},
			objScores: 0.6,
			expected:  1.0 / 4.0, // 3 matches (dist for last point = sqrt((4-2)^2 + (3.9-2)^2) = sqrt(7.61) < sqrt(8))
		},
		{
			name:      "just above distance threshold",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			detScores: 0.6,
			objPoints: [][]float64{{0, 0}, {1, 1}, {4, 4}},
			objScores: 0.6,
			expected:  1.0 / 3.0, // 2 matches (dist for last point = sqrt((4-2)^2 + (4-2)^2) = sqrt(8) >= sqrt(8))
		},
		{
			name:      "just under score threshold on detection",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			detScores: []float64{0.6, 0.6, 0.5},
			objPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			objScores: []float64{0.6, 0.6, 0.6},
			expected:  1.0 / 3.0, // 2 matches
		},
		{
			name:      "just under score threshold on tracked object",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			detScores: []float64{0.6, 0.6, 0.6},
			objPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			objScores: []float64{0.6, 0.6, 0.5},
			expected:  1.0 / 3.0, // 2 matches
		},
		{
			name:      "no match because of scores",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			detScores: 0.5,
			objPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			objScores: 0.5,
			expected:  1.0, // 0 matches
		},
		{
			name:      "no match because of distances",
			detPoints: [][]float64{{0, 0}, {1, 1}, {2, 2}},
			detScores: 0.6,
			objPoints: [][]float64{{2, 2}, {3, 3}, {4, 4}},
			objScores: 0.6,
			expected:  1.0, // 0 matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := newMockDetectionWithScores(tt.detPoints, tt.detScores)
			obj := newMockTrackedObjectWithScores(tt.objPoints, tt.objScores)

			result := voteD(det, obj)
			testutil.AssertAlmostEqual(t, result, tt.expected, 1e-6, tt.name)
		})
	}
}

// =============================================================================
// Test Normalized Euclidean Distance
// =============================================================================

// Python equivalent: norfair/distances.py::create_normalized_mean_euclidean_distance()
//
//	from norfair.distances import create_normalized_mean_euclidean_distance
//	import numpy as np
//
//	def test_normalized_euclidean():
//	    norm_e = create_normalized_mean_euclidean_distance(height=10, width=10)
//	    det = MockDetection([[1, 2], [3, 4]])
//	    obj = MockTrackedObject([[2, 2], [3, 4]])
//	    result = norm_e(det, obj)
//	    # Normalizes coordinates by dividing by height/width, then computes mean euclidean
//	    # Point 1: sqrt((0.1)^2 + (0)^2) = 0.1
//	    # Point 2: sqrt((0)^2 + (0)^2) = 0
//	    # Mean: (0.1 + 0) / 2 = 0.05
//	    assert abs(result - 0.05) < 1e-6
//
// Test cases match tools/validate_distances/main.py::test_normalized_euclidean()
func TestNormalizedEuclidean(t *testing.T) {
	normE := CreateNormalizedMeanEuclideanDistance(10, 10)

	tests := []struct {
		name     string
		det      [][]float64
		obj      [][]float64
		expected float64
	}{
		{
			name:     "perfect match",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{1, 2}, {3, 4}},
			expected: 0,
		},
		{
			name:     "float type",
			det:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			obj:      [][]float64{{1.1, 2.2}, {3.3, 4.4}},
			expected: 0,
		},
		{
			name:     "distance of 1 in 1 dimension of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 2}, {3, 4}},
			expected: 0.05, // (0.1 + 0) / 2 = 0.05
		},
		{
			name:     "distance of 2 in 1 dimension of 1 point",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 2}, {3, 4}},
			expected: 0.1, // (0.2 + 0) / 2 = 0.1
		},
		{
			name:     "distance of 2 in 1 dimension of all points",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 2}, {5, 4}},
			expected: 0.2, // (0.2 + 0.2) / 2 = 0.2
		},
		{
			name:     "distance of 2 in all dimensions of all points",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{3, 4}, {5, 6}},
			expected: math.Sqrt(8) / 10, // (sqrt(0.04+0.04) + sqrt(0.04+0.04)) / 2 = sqrt(8)/10
		},
		{
			name:     "distance of 1 in all dimensions of all points",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{2, 3}, {4, 5}},
			expected: math.Sqrt(2) / 10, // (sqrt(0.01+0.01) + sqrt(0.01+0.01)) / 2 = sqrt(2)/10
		},
		{
			name:     "negative difference",
			det:      [][]float64{{1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 0.1, // (0.2 + 0) / 2 = 0.1
		},
		{
			name:     "negative equals",
			det:      [][]float64{{-1, 2}, {3, 4}},
			obj:      [][]float64{{-1, 2}, {3, 4}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := newMockDetection(tt.det)
			obj := newMockTrackedObject(tt.obj)

			result := normE(det, obj)
			testutil.AssertAlmostEqual(t, result, tt.expected, 1e-6, tt.name)
		})
	}
}

// =============================================================================
// Test GetDistanceByName
// =============================================================================

// Python equivalent: N/A - This is a Go-specific factory function
//
//	# In Python, distance functions are imported directly:
//	from norfair.distances import frobenius, iou
//	from norfair import Distance
//
//	# No string-based factory in Python, but conceptually similar to:
//	distance_map = {
//	    "frobenius": Distance(frobenius),
//	    "iou": Distance(iou),
//	    "euclidean": ScipyDistance("euclidean"),
//	}
//	distance = distance_map["frobenius"]
//
// Test cases validate the GetDistanceByName factory function
func TestGetDistanceByName(t *testing.T) {
	// Test scalar distance
	t.Run("frobenius", func(t *testing.T) {
		distance := GetDistanceByName("frobenius")
		if distance == nil {
			t.Fatal("Expected non-nil distance")
		}
		if _, ok := distance.(*ScalarDistance); !ok {
			t.Errorf("Expected ScalarDistance, got %T", distance)
		}
	})

	// Test vectorized distance
	t.Run("iou", func(t *testing.T) {
		distance := GetDistanceByName("iou")
		if distance == nil {
			t.Fatal("Expected non-nil distance")
		}
		if _, ok := distance.(*VectorizedDistance); !ok {
			t.Errorf("Expected VectorizedDistance, got %T", distance)
		}
	})

	// Test scipy distance
	t.Run("euclidean", func(t *testing.T) {
		distance := GetDistanceByName("euclidean")
		if distance == nil {
			t.Fatal("Expected non-nil distance")
		}
		if _, ok := distance.(*ScipyDistance); !ok {
			t.Errorf("Expected ScipyDistance, got %T", distance)
		}
	})

	// Test invalid distance
	t.Run("invalid_distance", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for invalid distance name")
			}
		}()
		GetDistanceByName("invalid_distance")
	})
}
