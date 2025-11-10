package filterpy

import (
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go/internal/testutil"
)

// TestNewKalmanFilter verifies proper initialization
func TestNewKalmanFilter(t *testing.T) {
	dimX, dimZ := 4, 2
	kf := NewKalmanFilter(dimX, dimZ)

	// Verify dimensions
	if kf.GetDimX() != dimX {
		t.Errorf("Expected dimX=%d, got %d", dimX, kf.GetDimX())
	}
	if kf.GetDimZ() != dimZ {
		t.Errorf("Expected dimZ=%d, got %d", dimZ, kf.GetDimZ())
	}

	// Verify identity matrices initialization
	F := kf.GetF()
	for i := 0; i < dimX; i++ {
		for j := 0; j < dimX; j++ {
			expected := 0.0
			if i == j {
				expected = 1.0
			}
			if F.At(i, j) != expected {
				t.Errorf("F[%d,%d]: expected %f, got %f", i, j, expected, F.At(i, j))
			}
		}
	}

	// Verify H matrix is identity for measurement dimensions
	H := kf.GetH()
	for i := 0; i < dimZ; i++ {
		for j := 0; j < dimX; j++ {
			expected := 0.0
			if i == j {
				expected = 1.0
			}
			if H.At(i, j) != expected {
				t.Errorf("H[%d,%d]: expected %f, got %f", i, j, expected, H.At(i, j))
			}
		}
	}

	// Verify initial state is zero
	x := kf.GetX()
	for i := 0; i < dimX; i++ {
		if x.At(i, 0) != 0.0 {
			t.Errorf("Initial state x[%d]: expected 0, got %f", i, x.At(i, 0))
		}
	}
}

// TestKalmanFilter_Predict verifies prediction step
func TestKalmanFilter_Predict(t *testing.T) {
	kf := NewKalmanFilter(2, 1)

	// Set initial state [position=1, velocity=2]
	x := mat.NewDense(2, 1, []float64{1.0, 2.0})
	kf.SetState(x)

	// Set F matrix for constant velocity model: [1 dt; 0 1]
	dt := 1.0
	F := mat.NewDense(2, 2, []float64{
		1, dt,
		0, 1,
	})
	kf.F.Copy(F)

	// Set Q (process noise)
	Q := mat.NewDense(2, 2, []float64{
		0.1, 0,
		0, 0.1,
	})
	kf.Q.Copy(Q)

	// Initial covariance
	P := mat.NewDense(2, 2, []float64{
		1.0, 0,
		0, 1.0,
	})
	kf.SetCovariance(P)

	// Predict
	kf.Predict()

	// After prediction: x = F @ x = [1+2*1, 2] = [3, 2]
	state := kf.GetState()
	testutil.AssertAlmostEqual(t, state.At(0, 0), 3.0, 1e-10, "Predicted position")
	testutil.AssertAlmostEqual(t, state.At(1, 0), 2.0, 1e-10, "Predicted velocity")

	// Covariance should increase: P = F @ P @ F' + Q
	// Expected: [2.1, 1; 1, 1.1]
	covar := kf.GetCovariance()
	testutil.AssertAlmostEqual(t, covar.At(0, 0), 2.1, 1e-10, "P[0,0]")
	testutil.AssertAlmostEqual(t, covar.At(0, 1), 1.0, 1e-10, "P[0,1]")
	testutil.AssertAlmostEqual(t, covar.At(1, 0), 1.0, 1e-10, "P[1,0]")
	testutil.AssertAlmostEqual(t, covar.At(1, 1), 1.1, 1e-10, "P[1,1]")
}

// TestKalmanFilter_Update verifies update step
func TestKalmanFilter_Update(t *testing.T) {
	kf := NewKalmanFilter(2, 1)

	// Set initial state
	x := mat.NewDense(2, 1, []float64{0.0, 0.0})
	kf.SetState(x)

	// Set measurement matrix H = [1, 0] (measure position only)
	H := mat.NewDense(1, 2, []float64{1, 0})
	kf.H.Copy(H)

	// Set R (measurement noise)
	R := mat.NewDense(1, 1, []float64{1.0})
	kf.R.Copy(R)

	// Set P (large initial uncertainty)
	P := mat.NewDense(2, 2, []float64{
		10.0, 0,
		0, 10.0,
	})
	kf.SetCovariance(P)

	// Measurement: position = 5.0
	z := mat.NewDense(1, 1, []float64{5.0})

	// Update
	kf.Update(z, nil, nil)

	// With large P and small R, estimate should move significantly toward measurement
	state := kf.GetState()
	// Kalman gain K ≈ P @ H' @ (H @ P @ H' + R)^-1 ≈ [10/(10+1), 0]' ≈ [0.909, 0]'
	// x = x + K @ (z - H @ x) = [0, 0] + [0.909, 0]' @ 5.0 ≈ [4.545, 0]
	testutil.AssertAlmostEqual(t, state.At(0, 0), 4.545454545, 1e-8, "Updated position")
	testutil.AssertAlmostEqual(t, state.At(1, 0), 0.0, 1e-10, "Updated velocity (unmeasured)")
}

// TestKalmanFilter_PredictUpdateCycle verifies full cycle
func TestKalmanFilter_PredictUpdateCycle(t *testing.T) {
	kf := NewKalmanFilter(2, 1)

	// Simple 1D position+velocity tracking
	// State: [position, velocity]
	x := mat.NewDense(2, 1, []float64{0.0, 1.0}) // start at 0, moving at 1 unit/step
	kf.SetState(x)

	// F matrix for dt=1: x_new = x + v, v_new = v
	F := mat.NewDense(2, 2, []float64{
		1, 1,
		0, 1,
	})
	kf.F.Copy(F)

	// H matrix: measure position only
	H := mat.NewDense(1, 2, []float64{1, 0})
	kf.H.Copy(H)

	// Low noise
	Q := mat.NewDense(2, 2, []float64{
		0.01, 0,
		0, 0.01,
	})
	kf.Q.Copy(Q)

	R := mat.NewDense(1, 1, []float64{0.1})
	kf.R.Copy(R)

	P := mat.NewDense(2, 2, []float64{
		1.0, 0,
		0, 1.0,
	})
	kf.SetCovariance(P)

	// Simulate movement: object moves from 0 to 5 in 5 steps
	measurements := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	for i, z_val := range measurements {
		// Predict where object will be
		kf.Predict()

		// Measure actual position
		z := mat.NewDense(1, 1, []float64{z_val})
		kf.Update(z, nil, nil)

		// After a few iterations, state should track the linear motion
		state := kf.GetState()
		t.Logf("Step %d: position=%.3f, velocity=%.3f", i+1, state.At(0, 0), state.At(1, 0))

		// Position should be close to measurement
		if i >= 2 { // After a few iterations
			diff := state.At(0, 0) - z_val
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.5 {
				t.Errorf("Step %d: position %.3f too far from measurement %.3f", i+1, state.At(0, 0), z_val)
			}

			// Velocity should be close to 1.0
			velDiff := state.At(1, 0) - 1.0
			if velDiff < 0 {
				velDiff = -velDiff
			}
			if velDiff > 0.5 {
				t.Errorf("Step %d: velocity %.3f should be close to 1.0", i+1, state.At(1, 0))
			}
		}
	}
}

// TestKalmanFilter_PartialMeasurement tests H matrix with partial observation
func TestKalmanFilter_PartialMeasurement(t *testing.T) {
	// 2D position tracking: [x, y, vx, vy]
	kf := NewKalmanFilter(4, 2)

	// Initial state: at (1,2) with zero velocity
	x := mat.NewDense(4, 1, []float64{1.0, 2.0, 0.0, 0.0})
	kf.SetState(x)

	// H matrix that only observes x position, not y
	H := mat.NewDense(2, 4, []float64{
		1, 0, 0, 0, // Measure x position
		0, 0, 0, 0, // Don't measure y position
	})

	R := mat.NewDense(2, 2, []float64{
		1.0, 0,
		0, 1.0,
	})

	P := mat.NewDense(4, 4, nil)
	for i := 0; i < 4; i++ {
		P.Set(i, i, 10.0)
	}
	kf.SetCovariance(P)

	// Measurement: x=3, y=999 (y should be ignored)
	z := mat.NewDense(2, 1, []float64{3.0, 999.0})

	kf.Update(z, R, H)

	state := kf.GetState()

	// X position should update toward 3.0
	if state.At(0, 0) < 1.5 || state.At(0, 0) > 2.9 {
		t.Errorf("X position should update toward measurement: got %.3f", state.At(0, 0))
	}

	// Y position should remain close to 2.0 (not affected by measurement)
	testutil.AssertAlmostEqual(t, state.At(1, 0), 2.0, 0.1, "Y position (unobserved)")
}

// TestKalmanFilter_SingularInnovationCovariance tests robustness
func TestKalmanFilter_SingularInnovationCovariance(t *testing.T) {
	kf := NewKalmanFilter(2, 1)

	// Create a scenario where S becomes singular
	x := mat.NewDense(2, 1, []float64{1.0, 0.0})
	kf.SetState(x)

	// Set P and R to zero (will make S singular)
	P := mat.NewDense(2, 2, nil) // Zero covariance
	kf.SetCovariance(P)

	R := mat.NewDense(1, 1, nil) // Zero measurement noise
	kf.R.Copy(R)

	z := mat.NewDense(1, 1, []float64{5.0})

	// Should not crash - update should be skipped when S is singular
	kf.Update(z, nil, nil)

	// State should remain unchanged
	state := kf.GetState()
	testutil.AssertAlmostEqual(t, state.At(0, 0), 1.0, 1e-10, "State unchanged on singular S")
}

// TestKalmanFilter_GettersSetters verifies all accessor methods
func TestKalmanFilter_GettersSetters(t *testing.T) {
	kf := NewKalmanFilter(4, 2)

	// Test state getter/setter
	newX := mat.NewDense(4, 1, []float64{1, 2, 3, 4})
	kf.SetState(newX)
	x := kf.GetState()
	testutil.AssertMatrixAlmostEqual(t, x, newX, 1e-10, "State getter/setter")

	// Test covariance getter/setter
	newP := mat.NewDense(4, 4, nil)
	for i := 0; i < 4; i++ {
		newP.Set(i, i, float64(i+1))
	}
	kf.SetCovariance(newP)
	P := kf.GetCovariance()
	testutil.AssertMatrixAlmostEqual(t, P, newP, 1e-10, "Covariance getter/setter")

	// Test GetP alias
	P2 := kf.GetP()
	testutil.AssertMatrixAlmostEqual(t, P2, newP, 1e-10, "GetP alias")

	// Test GetX alias
	x2 := kf.GetX()
	testutil.AssertMatrixAlmostEqual(t, x2, newX, 1e-10, "GetX alias")

	// Test matrix getters
	F := kf.GetF()
	if r, c := F.Dims(); r != 4 || c != 4 {
		t.Errorf("F dimensions: expected (4,4), got (%d,%d)", r, c)
	}

	H := kf.GetH()
	if r, c := H.Dims(); r != 2 || c != 4 {
		t.Errorf("H dimensions: expected (2,4), got (%d,%d)", r, c)
	}

	R := kf.GetR()
	if r, c := R.Dims(); r != 2 || c != 2 {
		t.Errorf("R dimensions: expected (2,2), got (%d,%d)", r, c)
	}

	Q := kf.GetQ()
	if r, c := Q.Dims(); r != 4 || c != 4 {
		t.Errorf("Q dimensions: expected (4,4), got (%d,%d)", r, c)
	}
}

// TestKalmanFilter_MultiDimensional tests higher dimensional tracking
func TestKalmanFilter_MultiDimensional(t *testing.T) {
	// 3D position tracking: [x, y, z, vx, vy, vz]
	dimX := 6
	dimZ := 3
	kf := NewKalmanFilter(dimX, dimZ)

	// Initial state
	x := mat.NewDense(dimX, 1, []float64{
		1.0, 2.0, 3.0, // positions
		0.5, 0.5, 0.5, // velocities
	})
	kf.SetState(x)

	// Constant velocity model
	dt := 1.0
	F := mat.NewDense(dimX, dimX, []float64{
		1, 0, 0, dt, 0, 0,
		0, 1, 0, 0, dt, 0,
		0, 0, 1, 0, 0, dt,
		0, 0, 0, 1, 0, 0,
		0, 0, 0, 0, 1, 0,
		0, 0, 0, 0, 0, 1,
	})
	kf.F.Copy(F)

	// Predict
	kf.Predict()

	// After prediction with dt=1:
	// x_new = x + vx = 1.0 + 0.5 = 1.5
	// y_new = y + vy = 2.0 + 0.5 = 2.5
	// z_new = z + vz = 3.0 + 0.5 = 3.5
	// velocities unchanged
	state := kf.GetState()
	testutil.AssertAlmostEqual(t, state.At(0, 0), 1.5, 1e-10, "Predicted x")
	testutil.AssertAlmostEqual(t, state.At(1, 0), 2.5, 1e-10, "Predicted y")
	testutil.AssertAlmostEqual(t, state.At(2, 0), 3.5, 1e-10, "Predicted z")
	testutil.AssertAlmostEqual(t, state.At(3, 0), 0.5, 1e-10, "Predicted vx")
	testutil.AssertAlmostEqual(t, state.At(4, 0), 0.5, 1e-10, "Predicted vy")
	testutil.AssertAlmostEqual(t, state.At(5, 0), 0.5, 1e-10, "Predicted vz")
}
