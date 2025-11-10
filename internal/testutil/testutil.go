package testutil

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// Common test utilities shared across test files

func AlmostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func AssertAlmostEqual(t *testing.T, actual, expected, tolerance float64, msg string) {
	t.Helper()
	if !AlmostEqual(actual, expected, tolerance) {
		t.Errorf("%s: expected %.15f, got %.15f (diff: %.15e)", msg, expected, actual, math.Abs(actual-expected))
	}
}

func AssertMatrixAlmostEqual(t *testing.T, actual, expected *mat.Dense, tolerance float64, msg string) {
	t.Helper()
	r1, c1 := actual.Dims()
	r2, c2 := expected.Dims()
	if r1 != r2 || c1 != c2 {
		t.Fatalf("%s: dimension mismatch - actual (%d,%d) vs expected (%d,%d)", msg, r1, c1, r2, c2)
	}
	for i := 0; i < r1; i++ {
		for j := 0; j < c1; j++ {
			AssertAlmostEqual(t, actual.At(i, j), expected.At(i, j), tolerance,
				msg+" at ["+string(rune('0'+i))+","+string(rune('0'+j))+"]")
		}
	}
}
