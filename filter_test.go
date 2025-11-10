package norfairgo

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
	"gonum.org/v1/gonum/mat"
)

// Test helper functions are now in internal/testutil/testutil.go

// =============================================================================
// FilterPyKalmanFilter Tests
// =============================================================================

func TestFilterPyKalmanFilterFactory_Create(t *testing.T) {
	factory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)

	// Test with a simple 1-point, 2D detection
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	kf, ok := filter.(*FilterPyKalmanFilter)
	if !ok {
		t.Fatalf("Expected FilterPyKalmanFilter, got %T", filter)
	}

	// Check dimensions
	if kf.GetDimZ() != 2 {
		t.Errorf("Expected dimZ=2, got %d", kf.GetDimZ())
	}
	if kf.GetDimX() != 4 {
		t.Errorf("Expected dimX=4, got %d", kf.GetDimX())
	}

	// Check initial state
	x := kf.GetX()
	testutil.AssertAlmostEqual(t, x.At(0, 0), 1.0, 1e-10, "initial x[0]")
	testutil.AssertAlmostEqual(t, x.At(1, 0), 1.0, 1e-10, "initial x[1]")
	testutil.AssertAlmostEqual(t, x.At(2, 0), 0.0, 1e-10, "initial x[2] (velocity)")
	testutil.AssertAlmostEqual(t, x.At(3, 0), 0.0, 1e-10, "initial x[3] (velocity)")

	// Check F matrix
	expectedF := mat.NewDense(4, 4, []float64{
		1, 0, 1, 0,
		0, 1, 0, 1,
		0, 0, 1, 0,
		0, 0, 0, 1,
	})
	testutil.AssertMatrixAlmostEqual(t, kf.GetF(), expectedF, 1e-10, "F matrix")

	// Check H matrix
	expectedH := mat.NewDense(2, 4, []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
	})
	testutil.AssertMatrixAlmostEqual(t, kf.GetH(), expectedH, 1e-10, "H matrix")

	// Check R matrix (should be 4.0 * I)
	R := kf.GetR()
	testutil.AssertAlmostEqual(t, R.At(0, 0), 4.0, 1e-10, "R[0,0]")
	testutil.AssertAlmostEqual(t, R.At(1, 1), 4.0, 1e-10, "R[1,1]")

	// Check P matrix
	P := kf.GetP()
	testutil.AssertAlmostEqual(t, P.At(0, 0), 10.0, 1e-10, "P[0,0] position variance")
	testutil.AssertAlmostEqual(t, P.At(1, 1), 10.0, 1e-10, "P[1,1] position variance")
	testutil.AssertAlmostEqual(t, P.At(2, 2), 1.0, 1e-10, "P[2,2] velocity variance")
	testutil.AssertAlmostEqual(t, P.At(3, 3), 1.0, 1e-10, "P[3,3] velocity variance")
}

func TestFilterPyKalmanFilter_StaticObject(t *testing.T) {
	factory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Test predict-update cycle with static object
	for i := 0; i < 5; i++ {
		filter.Predict()
		measurement := mat.NewDense(2, 1, []float64{1.0, 1.0})
		filter.Update(measurement, nil, nil)

		state := filter.GetState()
		// Position should stay close to 1.0
		testutil.AssertAlmostEqual(t, state.At(0, 0), 1.0, 0.1, "position x after iteration")
		testutil.AssertAlmostEqual(t, state.At(1, 0), 1.0, 0.1, "position y after iteration")
		// Velocity should stay close to 0
		testutil.AssertAlmostEqual(t, state.At(2, 0), 0.0, 0.1, "velocity x after iteration")
		testutil.AssertAlmostEqual(t, state.At(3, 0), 0.0, 0.1, "velocity y after iteration")
	}
}

func TestFilterPyKalmanFilter_MovingObject(t *testing.T) {
	factory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Simulate a moving object (y increases by 1 each step)
	positions := [][]float64{
		{1.0, 1.0},
		{1.0, 2.0},
		{1.0, 3.0},
		{1.0, 4.0},
	}

	for i, pos := range positions {
		measurement := mat.NewDense(2, 1, []float64{pos[0], pos[1]})
		filter.Update(measurement, nil, nil)
		filter.Predict()

		state := filter.GetState()
		if i == len(positions)-1 {
			// After last update, check that estimate is reasonable
			// Position should be between 3 and 4
			testutil.AssertAlmostEqual(t, state.At(0, 0), 1.0, 0.1, "final position x")
			if state.At(1, 0) < 3.0 || state.At(1, 0) > 4.5 {
				t.Errorf("Expected position y between 3 and 4.5, got %.2f", state.At(1, 0))
			}
		}
	}
}

// =============================================================================
// OptimizedKalmanFilter Tests
// =============================================================================

func TestOptimizedKalmanFilterFactory_Create(t *testing.T) {
	factory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)

	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	okf, ok := filter.(*OptimizedKalmanFilter)
	if !ok {
		t.Fatalf("Expected OptimizedKalmanFilter, got %T", filter)
	}

	// Check dimensions
	if okf.dimZ != 2 {
		t.Errorf("Expected dimZ=2, got %d", okf.dimZ)
	}
	if okf.dimX != 4 {
		t.Errorf("Expected dimX=4, got %d", okf.dimX)
	}

	// Check initial state
	testutil.AssertAlmostEqual(t, okf.x.At(0, 0), 1.0, 1e-10, "initial x[0]")
	testutil.AssertAlmostEqual(t, okf.x.At(1, 0), 1.0, 1e-10, "initial x[1]")
	testutil.AssertAlmostEqual(t, okf.x.At(2, 0), 0.0, 1e-10, "initial x[2] (velocity)")
	testutil.AssertAlmostEqual(t, okf.x.At(3, 0), 0.0, 1e-10, "initial x[3] (velocity)")

	// Check variance vectors
	testutil.AssertAlmostEqual(t, okf.PosVariance[0], 10.0, 1e-10, "PosVariance[0]")
	testutil.AssertAlmostEqual(t, okf.PosVariance[1], 10.0, 1e-10, "PosVariance[1]")
	testutil.AssertAlmostEqual(t, okf.VelVariance[0], 1.0, 1e-10, "VelVariance[0]")
	testutil.AssertAlmostEqual(t, okf.VelVariance[1], 1.0, 1e-10, "VelVariance[1]")
	testutil.AssertAlmostEqual(t, okf.PosVelCovariance[0], 0.0, 1e-10, "PosVelCovariance[0]")
}

func TestOptimizedKalmanFilter_StaticObject(t *testing.T) {
	factory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Test predict-update cycle with static object
	for i := 0; i < 5; i++ {
		filter.Predict()
		measurement := mat.NewDense(2, 1, []float64{1.0, 1.0})
		filter.Update(measurement, nil, nil)

		state := filter.GetState()
		// Position should stay close to 1.0
		testutil.AssertAlmostEqual(t, state.At(0, 0), 1.0, 0.1, "position x after iteration")
		testutil.AssertAlmostEqual(t, state.At(1, 0), 1.0, 0.1, "position y after iteration")
		// Velocity should stay close to 0
		testutil.AssertAlmostEqual(t, state.At(2, 0), 0.0, 0.1, "velocity x after iteration")
		testutil.AssertAlmostEqual(t, state.At(3, 0), 0.0, 0.1, "velocity y after iteration")
	}
}

func TestOptimizedKalmanFilter_MovingObject(t *testing.T) {
	factory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Simulate a moving object (y increases by 1 each step)
	positions := [][]float64{
		{1.0, 1.0},
		{1.0, 2.0},
		{1.0, 3.0},
		{1.0, 4.0},
	}

	for i, pos := range positions {
		measurement := mat.NewDense(2, 1, []float64{pos[0], pos[1]})
		filter.Update(measurement, nil, nil)
		filter.Predict()

		state := filter.GetState()
		if i == len(positions)-1 {
			// After last update, check that estimate is reasonable
			testutil.AssertAlmostEqual(t, state.At(0, 0), 1.0, 0.1, "final position x")
			if state.At(1, 0) < 3.0 || state.At(1, 0) > 4.5 {
				t.Errorf("Expected position y between 3 and 4.5, got %.2f", state.At(1, 0))
			}
		}
	}
}

// =============================================================================
// NoFilter Tests
// =============================================================================

func TestNoFilterFactory_Create(t *testing.T) {
	factory := NewNoFilterFactory()

	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	nf, ok := filter.(*NoFilter)
	if !ok {
		t.Fatalf("Expected NoFilter, got %T", filter)
	}

	// Check dimensions
	if nf.dimZ != 2 {
		t.Errorf("Expected dimZ=2, got %d", nf.dimZ)
	}
	if nf.dimX != 4 {
		t.Errorf("Expected dimX=4, got %d", nf.dimX)
	}

	// Check initial state
	testutil.AssertAlmostEqual(t, nf.x.At(0, 0), 1.0, 1e-10, "initial x[0]")
	testutil.AssertAlmostEqual(t, nf.x.At(1, 0), 1.0, 1e-10, "initial x[1]")
}

func TestNoFilter_Predict(t *testing.T) {
	factory := NewNoFilterFactory()
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Predict should be a no-op
	stateBefore := mat.DenseCopyOf(filter.GetState())
	filter.Predict()
	stateAfter := filter.GetState()

	testutil.AssertMatrixAlmostEqual(t, stateAfter, stateBefore, 1e-10, "state unchanged after predict")
}

func TestNoFilter_Update(t *testing.T) {
	factory := NewNoFilterFactory()
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Update with new measurement
	measurement := mat.NewDense(2, 1, []float64{2.0, 3.0})
	filter.Update(measurement, nil, nil)

	state := filter.GetState()
	testutil.AssertAlmostEqual(t, state.At(0, 0), 2.0, 1e-10, "position x updated")
	testutil.AssertAlmostEqual(t, state.At(1, 0), 3.0, 1e-10, "position y updated")
}

// =============================================================================
// Comparison Tests - FilterPy vs Optimized
// =============================================================================

func TestFilterComparison_StaticObject(t *testing.T) {
	// Both filters should produce similar results for a static object
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})

	filterPyFactory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)
	optimizedFactory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)

	filterPy := filterPyFactory.CreateFilter(initialDetection)
	optimized := optimizedFactory.CreateFilter(initialDetection)

	// Run same sequence on both
	for i := 0; i < 10; i++ {
		filterPy.Predict()
		optimized.Predict()

		measurement := mat.NewDense(2, 1, []float64{1.0, 1.0})
		filterPy.Update(measurement, nil, nil)
		optimized.Update(measurement, nil, nil)
	}

	statePy := filterPy.GetState()
	stateOpt := optimized.GetState()

	// States should be very close (allowing for some numerical differences)
	testutil.AssertAlmostEqual(t, stateOpt.At(0, 0), statePy.At(0, 0), 0.01, "position x comparison")
	testutil.AssertAlmostEqual(t, stateOpt.At(1, 0), statePy.At(1, 0), 0.01, "position y comparison")
	testutil.AssertAlmostEqual(t, stateOpt.At(2, 0), statePy.At(2, 0), 0.01, "velocity x comparison")
	testutil.AssertAlmostEqual(t, stateOpt.At(3, 0), statePy.At(3, 0), 0.01, "velocity y comparison")
}

func TestFilterComparison_MovingObject(t *testing.T) {
	// Both filters should produce similar results for a moving object
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})

	filterPyFactory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)
	optimizedFactory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)

	filterPy := filterPyFactory.CreateFilter(initialDetection)
	optimized := optimizedFactory.CreateFilter(initialDetection)

	// Simulate a moving object
	positions := [][]float64{
		{1.0, 1.0},
		{1.0, 2.0},
		{1.0, 3.0},
		{1.0, 4.0},
		{1.0, 5.0},
	}

	for _, pos := range positions {
		measurement := mat.NewDense(2, 1, []float64{pos[0], pos[1]})

		filterPy.Update(measurement, nil, nil)
		optimized.Update(measurement, nil, nil)

		filterPy.Predict()
		optimized.Predict()
	}

	statePy := filterPy.GetState()
	stateOpt := optimized.GetState()

	// States should be reasonably close
	testutil.AssertAlmostEqual(t, stateOpt.At(0, 0), statePy.At(0, 0), 0.1, "final position x comparison")
	testutil.AssertAlmostEqual(t, stateOpt.At(1, 0), statePy.At(1, 0), 0.1, "final position y comparison")
	testutil.AssertAlmostEqual(t, stateOpt.At(2, 0), statePy.At(2, 0), 0.1, "final velocity x comparison")
	testutil.AssertAlmostEqual(t, stateOpt.At(3, 0), statePy.At(3, 0), 0.1, "final velocity y comparison")
}

// =============================================================================
// Partial Measurement Tests (H matrix)
// =============================================================================

func TestFilterPyKalmanFilter_PartialMeasurement(t *testing.T) {
	factory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Create H matrix that only observes the first dimension
	H := mat.NewDense(2, 4, []float64{
		1, 0, 0, 0,
		0, 0, 0, 0, // Second dimension not observed
	})

	measurement := mat.NewDense(2, 1, []float64{2.0, 100.0}) // Second value should be ignored
	filter.Update(measurement, nil, H)

	state := filter.GetState()
	// First position should be updated towards 2.0
	if state.At(0, 0) < 1.5 {
		t.Errorf("Expected position x > 1.5, got %.2f", state.At(0, 0))
	}
	// Second position should stay close to 1.0 (not affected by the 100.0)
	testutil.AssertAlmostEqual(t, state.At(1, 0), 1.0, 0.1, "position y should not be updated")
}

func TestOptimizedKalmanFilter_PartialMeasurement(t *testing.T) {
	factory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)
	initialDetection := mat.NewDense(1, 2, []float64{1.0, 1.0})
	filter := factory.CreateFilter(initialDetection)

	// Create H matrix that only observes the first dimension
	H := mat.NewDense(2, 4, []float64{
		1, 0, 0, 0,
		0, 0, 0, 0, // Second dimension not observed
	})

	measurement := mat.NewDense(2, 1, []float64{2.0, 100.0}) // Second value should be ignored
	filter.Update(measurement, nil, H)

	state := filter.GetState()
	// First position should be updated towards 2.0
	if state.At(0, 0) < 1.5 {
		t.Errorf("Expected position x > 1.5, got %.2f", state.At(0, 0))
	}
	// Second position should stay close to 1.0 (not affected by the 100.0)
	testutil.AssertAlmostEqual(t, state.At(1, 0), 1.0, 0.1, "position y should not be updated")
}

// =============================================================================
// Multi-point Tests
// =============================================================================

func TestFilters_MultiPoint(t *testing.T) {
	// Test with 2 points, 2D each (e.g., bounding box corners)
	initialDetection := mat.NewDense(2, 2, []float64{
		0.0, 0.0,
		1.0, 1.0,
	})

	filterPyFactory := NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)
	optimizedFactory := NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)

	filterPy := filterPyFactory.CreateFilter(initialDetection)
	optimized := optimizedFactory.CreateFilter(initialDetection)

	if filterPy.GetDimZ() != 4 {
		t.Errorf("Expected dimZ=4 for 2 points, got %d", filterPy.GetDimZ())
	}

	// Update with new measurements
	measurement := mat.NewDense(4, 1, []float64{0.1, 0.1, 1.1, 1.1})
	filterPy.Update(measurement, nil, nil)
	optimized.Update(measurement, nil, nil)

	// Both should handle multi-point correctly
	statePy := filterPy.GetState()
	stateOpt := optimized.GetState()

	// Check that both filters updated properly
	testutil.AssertAlmostEqual(t, statePy.At(0, 0), 0.1, 0.1, "FilterPy point 1 x")
	testutil.AssertAlmostEqual(t, stateOpt.At(0, 0), 0.1, 0.1, "Optimized point 1 x")
}
