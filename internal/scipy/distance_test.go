// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause

package scipy

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestCdist_Euclidean(t *testing.T) {
	// Test simple case
	XA := mat.NewDense(2, 2, []float64{
		0, 0,
		1, 0,
	})
	XB := mat.NewDense(2, 2, []float64{
		0, 1,
		1, 1,
	})

	result := Cdist(XA, XB, "euclidean")

	// [0,0] to [0,1]: sqrt(1) = 1
	// [0,0] to [1,1]: sqrt(2)
	// [1,0] to [0,1]: sqrt(2)
	// [1,0] to [1,1]: sqrt(1) = 1
	expected := mat.NewDense(2, 2, []float64{
		1.0, math.Sqrt(2),
		math.Sqrt(2), 1.0,
	})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Euclidean distance incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}

func TestCdist_Manhattan(t *testing.T) {
	XA := mat.NewDense(2, 3, []float64{
		1, 2, 3,
		4, 5, 6,
	})
	XB := mat.NewDense(2, 3, []float64{
		1, 1, 1,
		2, 2, 2,
	})

	// Test both "cityblock" and "manhattan" names
	for _, metric := range []string{"cityblock", "manhattan"} {
		result := Cdist(XA, XB, metric)

		// |1-1| + |2-1| + |3-1| = 0 + 1 + 2 = 3
		// |1-2| + |2-2| + |3-2| = 1 + 0 + 1 = 2
		// |4-1| + |5-1| + |6-1| = 3 + 4 + 5 = 12
		// |4-2| + |5-2| + |6-2| = 2 + 3 + 4 = 9
		expected := mat.NewDense(2, 2, []float64{
			3.0, 2.0,
			12.0, 9.0,
		})

		if !mat.EqualApprox(result, expected, 1e-10) {
			t.Errorf("%s distance incorrect.\nGot:\n%v\nExpected:\n%v", metric, mat.Formatted(result), mat.Formatted(expected))
		}
	}
}

func TestCdist_Cosine(t *testing.T) {
	// Orthogonal vectors: [1, 0] and [0, 1]
	XA := mat.NewDense(1, 2, []float64{1, 0})
	XB := mat.NewDense(1, 2, []float64{0, 1})

	result := Cdist(XA, XB, "cosine")

	// Cosine similarity = 0, so cosine distance = 1
	expected := mat.NewDense(1, 1, []float64{1.0})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Cosine distance incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}

func TestCdist_Cosine_Parallel(t *testing.T) {
	// Parallel vectors: [1, 1] and [2, 2]
	XA := mat.NewDense(1, 2, []float64{1, 1})
	XB := mat.NewDense(1, 2, []float64{2, 2})

	result := Cdist(XA, XB, "cosine")

	// Cosine similarity = 1, so cosine distance = 0
	expected := mat.NewDense(1, 1, []float64{0.0})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Cosine distance for parallel vectors incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}

func TestCdist_Cosine_ZeroVector(t *testing.T) {
	// Zero vector case
	XA := mat.NewDense(1, 2, []float64{0, 0})
	XB := mat.NewDense(1, 2, []float64{1, 1})

	result := Cdist(XA, XB, "cosine")

	// Distance should be 0 for zero vector (as per implementation)
	expected := mat.NewDense(1, 1, []float64{0.0})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Cosine distance with zero vector incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}

func TestCdist_SquaredEuclidean(t *testing.T) {
	XA := mat.NewDense(2, 2, []float64{
		0, 0,
		3, 4,
	})
	XB := mat.NewDense(1, 2, []float64{
		0, 0,
	})

	result := Cdist(XA, XB, "sqeuclidean")

	// (0-0)^2 + (0-0)^2 = 0
	// (3-0)^2 + (4-0)^2 = 9 + 16 = 25
	expected := mat.NewDense(2, 1, []float64{
		0.0,
		25.0,
	})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Squared Euclidean distance incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}

func TestCdist_Chebyshev(t *testing.T) {
	XA := mat.NewDense(2, 3, []float64{
		1, 2, 3,
		0, 0, 0,
	})
	XB := mat.NewDense(2, 3, []float64{
		2, 1, 1,
		5, 5, 5,
	})

	result := Cdist(XA, XB, "chebyshev")

	// max(|1-2|, |2-1|, |3-1|) = max(1, 1, 2) = 2
	// max(|1-5|, |2-5|, |3-5|) = max(4, 3, 2) = 4
	// max(|0-2|, |0-1|, |0-1|) = max(2, 1, 1) = 2
	// max(|0-5|, |0-5|, |0-5|) = max(5, 5, 5) = 5
	expected := mat.NewDense(2, 2, []float64{
		2.0, 4.0,
		2.0, 5.0,
	})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Chebyshev distance incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}

func TestCdist_DifferentDimensions(t *testing.T) {
	// Different number of rows is OK
	XA := mat.NewDense(3, 2, []float64{
		1, 2,
		3, 4,
		5, 6,
	})
	XB := mat.NewDense(2, 2, []float64{
		0, 0,
		1, 1,
	})

	result := Cdist(XA, XB, "euclidean")

	// Should produce 3x2 matrix
	rows, cols := result.Dims()
	if rows != 3 || cols != 2 {
		t.Errorf("Expected 3x2 result, got %dx%d", rows, cols)
	}
}

func TestCdist_PanicOnMismatchedColumns(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for mismatched columns")
		}
	}()

	XA := mat.NewDense(2, 2, []float64{1, 2, 3, 4})
	XB := mat.NewDense(2, 3, []float64{1, 2, 3, 4, 5, 6})

	Cdist(XA, XB, "euclidean")
}

func TestCdist_PanicOnUnsupportedMetric(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for unsupported metric")
		}
	}()

	XA := mat.NewDense(2, 2, []float64{1, 2, 3, 4})
	XB := mat.NewDense(2, 2, []float64{1, 2, 3, 4})

	Cdist(XA, XB, "unsupported_metric")
}

func TestCdist_SingleVectors(t *testing.T) {
	// Test with single vectors
	XA := mat.NewDense(1, 3, []float64{1, 2, 3})
	XB := mat.NewDense(1, 3, []float64{4, 5, 6})

	result := Cdist(XA, XB, "euclidean")

	// sqrt((1-4)^2 + (2-5)^2 + (3-6)^2) = sqrt(9 + 9 + 9) = sqrt(27)
	expected := math.Sqrt(27)

	if math.Abs(result.At(0, 0)-expected) > 1e-10 {
		t.Errorf("Single vector distance incorrect. Got: %v, Expected: %v", result.At(0, 0), expected)
	}
}

func TestCdist_IdenticalVectors(t *testing.T) {
	// Test with identical vectors (distance should be 0)
	XA := mat.NewDense(2, 3, []float64{
		1, 2, 3,
		4, 5, 6,
	})
	XB := mat.NewDense(2, 3, []float64{
		1, 2, 3,
		4, 5, 6,
	})

	metrics := []string{"euclidean", "manhattan", "sqeuclidean", "chebyshev"}
	for _, metric := range metrics {
		result := Cdist(XA, XB, metric)

		// Diagonal should be all zeros
		for i := 0; i < 2; i++ {
			if math.Abs(result.At(i, i)) > 1e-10 {
				t.Errorf("Distance between identical vectors should be 0 for %s, got %v at (%d,%d)",
					metric, result.At(i, i), i, i)
			}
		}
	}
}

func TestCdist_Cosine_AntiParallel(t *testing.T) {
	// Anti-parallel vectors: [1, 0] and [-1, 0]
	XA := mat.NewDense(1, 2, []float64{1, 0})
	XB := mat.NewDense(1, 2, []float64{-1, 0})

	result := Cdist(XA, XB, "cosine")

	// Cosine similarity = -1, so cosine distance = 1 - (-1) = 2
	expected := mat.NewDense(1, 1, []float64{2.0})

	if !mat.EqualApprox(result, expected, 1e-10) {
		t.Errorf("Cosine distance for anti-parallel vectors incorrect.\nGot:\n%v\nExpected:\n%v", mat.Formatted(result), mat.Formatted(expected))
	}
}
