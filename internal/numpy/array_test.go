package numpy

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
)

// TestLinspace_Basic verifies basic linspace functionality
func TestLinspace_Basic(t *testing.T) {
	// Test case: 5 values from 0 to 10
	result := Linspace(0, 10, 5)

	expected := []float64{0, 2.5, 5.0, 7.5, 10}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, val := range result {
		testutil.AssertAlmostEqual(t, val, expected[i], 1e-10, "Linspace value")
	}
}

// TestLinspace_TwoPoints verifies linspace with n=2
func TestLinspace_TwoPoints(t *testing.T) {
	result := Linspace(1, 10, 2)

	expected := []float64{1, 10}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	testutil.AssertAlmostEqual(t, result[0], expected[0], 1e-10, "Start value")
	testutil.AssertAlmostEqual(t, result[1], expected[1], 1e-10, "End value")
}

// TestLinspace_SinglePoint verifies linspace with n=1
func TestLinspace_SinglePoint(t *testing.T) {
	result := Linspace(5, 10, 1)

	if len(result) != 1 {
		t.Fatalf("Expected length 1, got %d", len(result))
	}

	testutil.AssertAlmostEqual(t, result[0], 5.0, 1e-10, "Single point should return start")
}

// TestLinspace_Zero verifies linspace with n=0
func TestLinspace_Zero(t *testing.T) {
	result := Linspace(0, 10, 0)

	if len(result) != 0 {
		t.Fatalf("Expected empty slice, got length %d", len(result))
	}
}

// TestLinspace_Negative verifies linspace with negative values
func TestLinspace_Negative(t *testing.T) {
	result := Linspace(-10, 10, 5)

	expected := []float64{-10, -5, 0, 5, 10}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, val := range result {
		testutil.AssertAlmostEqual(t, val, expected[i], 1e-10, "Negative range value")
	}
}

// TestLinspace_ReverseRange verifies linspace with start > end
func TestLinspace_ReverseRange(t *testing.T) {
	result := Linspace(10, 0, 5)

	expected := []float64{10, 7.5, 5, 2.5, 0}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, val := range result {
		testutil.AssertAlmostEqual(t, val, expected[i], 1e-10, "Reverse range value")
	}
}

// TestLinspace_FloatingPoint verifies linspace with floating-point boundaries
func TestLinspace_FloatingPoint(t *testing.T) {
	result := Linspace(0.1, 0.9, 5)

	expected := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, val := range result {
		testutil.AssertAlmostEqual(t, val, expected[i], 1e-10, "Floating-point value")
	}
}

// TestLinspace_LargeN verifies linspace with large n
func TestLinspace_LargeN(t *testing.T) {
	result := Linspace(0, 100, 101)

	if len(result) != 101 {
		t.Fatalf("Expected length 101, got %d", len(result))
	}

	// Verify first, middle, and last values
	testutil.AssertAlmostEqual(t, result[0], 0.0, 1e-10, "Start value")
	testutil.AssertAlmostEqual(t, result[50], 50.0, 1e-10, "Middle value")
	testutil.AssertAlmostEqual(t, result[100], 100.0, 1e-10, "End value")

	// Verify all values are monotonically increasing
	for i := 1; i < len(result); i++ {
		if result[i] <= result[i-1] {
			t.Errorf("Values not monotonically increasing: result[%d]=%.10f, result[%d]=%.10f",
				i-1, result[i-1], i, result[i])
		}
	}
}

// TestLinspace_EndpointExact verifies endpoint is exact (not drifted by floating point)
func TestLinspace_EndpointExact(t *testing.T) {
	// Test with values that might accumulate floating-point error
	result := Linspace(0, 1, 100)

	if len(result) != 100 {
		t.Fatalf("Expected length 100, got %d", len(result))
	}

	// Endpoint should be exactly 1.0, not 0.9999999...
	if result[99] != 1.0 {
		t.Errorf("Endpoint should be exactly 1.0, got %.20f", result[99])
	}
}

// TestLinspace_ZeroRange verifies linspace with start == end
func TestLinspace_ZeroRange(t *testing.T) {
	result := Linspace(5, 5, 10)

	if len(result) != 10 {
		t.Fatalf("Expected length 10, got %d", len(result))
	}

	// All values should be exactly 5
	for i, val := range result {
		testutil.AssertAlmostEqual(t, val, 5.0, 1e-10, "Zero range value")
		if i > 0 {
			t.Logf("result[%d] = %.20f", i, val)
		}
	}
}

// TestLinspace_SmallInterval verifies linspace with very small intervals
func TestLinspace_SmallInterval(t *testing.T) {
	result := Linspace(0, 1e-6, 5)

	if len(result) != 5 {
		t.Fatalf("Expected length 5, got %d", len(result))
	}

	// Verify start and end
	testutil.AssertAlmostEqual(t, result[0], 0.0, 1e-15, "Start value")
	testutil.AssertAlmostEqual(t, result[4], 1e-6, 1e-15, "End value")

	// Verify spacing
	expectedStep := 0.25e-6
	for i := 1; i < len(result)-1; i++ {
		expected := float64(i) * expectedStep
		testutil.AssertAlmostEqual(t, result[i], expected, 1e-15, "Small interval value")
	}
}

// TestLinspace_LargeInterval verifies linspace with very large intervals
func TestLinspace_LargeInterval(t *testing.T) {
	result := Linspace(0, 1e10, 5)

	if len(result) != 5 {
		t.Fatalf("Expected length 5, got %d", len(result))
	}

	expected := []float64{0, 2.5e9, 5e9, 7.5e9, 1e10}
	for i, val := range result {
		testutil.AssertAlmostEqual(t, val, expected[i], 1e-3, "Large interval value")
	}
}

// TestLinspace_MatchesNumpyBehavior verifies behavior matches numpy.linspace
func TestLinspace_MatchesNumpyBehavior(t *testing.T) {
	// Test cases that match numpy.linspace behavior
	testCases := []struct {
		start    float64
		end      float64
		n        int
		expected []float64
	}{
		{0, 10, 5, []float64{0, 2.5, 5, 7.5, 10}},
		{-5, 5, 11, []float64{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}},
		{0, 1, 3, []float64{0, 0.5, 1}},
		{10, 0, 6, []float64{10, 8, 6, 4, 2, 0}},
		{0, 0, 5, []float64{0, 0, 0, 0, 0}},
	}

	for _, tc := range testCases {
		result := Linspace(tc.start, tc.end, tc.n)

		if len(result) != len(tc.expected) {
			t.Errorf("Linspace(%.1f, %.1f, %d): expected length %d, got %d",
				tc.start, tc.end, tc.n, len(tc.expected), len(result))
			continue
		}

		for i, val := range result {
			testutil.AssertAlmostEqual(t, val, tc.expected[i], 1e-10,
				"Linspace value mismatch")
		}
	}
}

// TestLinspace_Consistency verifies spacing is consistent
func TestLinspace_Consistency(t *testing.T) {
	result := Linspace(0, 100, 11)

	if len(result) != 11 {
		t.Fatalf("Expected length 11, got %d", len(result))
	}

	// Calculate differences between consecutive values
	for i := 1; i < len(result)-1; i++ {
		diff := result[i] - result[i-1]
		testutil.AssertAlmostEqual(t, diff, 10.0, 1e-10, "Consistent spacing")
	}
}
