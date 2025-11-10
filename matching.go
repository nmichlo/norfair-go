package norfairgo

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Matching Algorithm
// =============================================================================

// MatchDetectionsAndObjects performs greedy minimum-distance matching between
// candidates (detections or tracked objects) and existing tracked objects.
//
// This implements a greedy matching algorithm that repeatedly finds the global
// minimum distance and creates a match, then invalidates that row and column
// to prevent re-matching. This ensures one-to-one correspondence.
//
// Note: This is NOT the optimal assignment algorithm (Hungarian). It's a simpler
// greedy approach that works well in practice for object tracking.
//
// Parameters:
//   - distanceMatrix: NxM matrix of distances where N=candidates, M=objects
//   - distanceThreshold: Maximum distance to consider a valid match
//
// Returns:
//   - matchedCandIndices: Indices of matched candidates (detections)
//   - matchedObjIndices: Indices of matched objects (tracked objects)
//
// The returned slices have the same length and matchedCandIndices[i] corresponds
// to matchedObjIndices[i].
func MatchDetectionsAndObjects(
	distanceMatrix *mat.Dense,
	distanceThreshold float64,
) (matchedCandIndices, matchedObjIndices []int) {
	// Handle empty matrix
	rows, cols := distanceMatrix.Dims()
	if rows == 0 || cols == 0 {
		return []int{}, []int{}
	}

	// Make a copy to avoid modifying the original matrix
	matrixCopy := mat.DenseCopyOf(distanceMatrix)

	// Initialize result lists
	candIndices := []int{}
	objIndices := []int{}

	// Greedy matching loop
	currentMin := minMatrix(matrixCopy)

	for currentMin < distanceThreshold {
		// Find the position of the minimum value
		flatIdx := argMin(matrixCopy)
		detIdx := flatIdx / cols  // Row index
		objIdx := flatIdx % cols   // Column index

		// Record the match
		candIndices = append(candIndices, detIdx)
		objIndices = append(objIndices, objIdx)

		// Invalidate the matched row and column by setting to high value
		// This prevents re-matching the same detection or object
		invalidValue := distanceThreshold + 1.0

		// Set entire row to invalid
		for c := 0; c < cols; c++ {
			matrixCopy.Set(detIdx, c, invalidValue)
		}

		// Set entire column to invalid
		for r := 0; r < rows; r++ {
			matrixCopy.Set(r, objIdx, invalidValue)
		}

		// Update current minimum for next iteration
		currentMin = minMatrix(matrixCopy)
	}

	return candIndices, objIndices
}

// =============================================================================
// Helper Functions
// =============================================================================

// argMin returns the flattened index of the minimum value in the matrix.
// If there are multiple minimum values, it returns the first one encountered.
func argMin(m *mat.Dense) int {
	rows, cols := m.Dims()
	minVal := math.Inf(1)
	minIdx := 0

	for i := 0; i < rows*cols; i++ {
		r := i / cols
		c := i % cols
		val := m.At(r, c)
		if val < minVal {
			minVal = val
			minIdx = i
		}
	}

	return minIdx
}

// minMatrix returns the minimum value in the matrix.
func minMatrix(m *mat.Dense) float64 {
	rows, cols := m.Dims()
	minVal := math.Inf(1)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			val := m.At(i, j)
			if val < minVal {
				minVal = val
			}
		}
	}

	return minVal
}

// hasNaN checks if the matrix contains any NaN values.
// NaN values in the distance matrix indicate errors in the distance function
// and must be detected early to prevent silent failures.
func hasNaN(m *mat.Dense) bool {
	rows, cols := m.Dims()

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if math.IsNaN(m.At(i, j)) {
				return true
			}
		}
	}

	return false
}

// ValidateDistanceMatrix checks a distance matrix for NaN values and returns
// an error if any are found. This should be called after computing distances
// and before performing matching.
func ValidateDistanceMatrix(m *mat.Dense) error {
	if hasNaN(m) {
		return fmt.Errorf(
			"received NaN values from distance function, please check your distance function for errors",
		)
	}
	return nil
}
