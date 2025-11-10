package norfairgo

import (
	"github.com/nmichlo/norfair-go/internal/filterpy"
	"gonum.org/v1/gonum/mat"
)

// FilterFactory is an interface for creating filters
type FilterFactory interface {
	CreateFilter(initialDetection *mat.Dense) Filter
}

// Filter is an interface that all filters must implement
type Filter interface {
	Predict()
	Update(detectionPointsFlatten *mat.Dense, R, H *mat.Dense)
	GetState() *mat.Dense
	GetDimZ() int
	// GetStateVector returns direct access to the state vector (for TrackedObject)
	GetStateVector() *mat.Dense
	// SetStateVector sets the state vector directly (for first-time point handling)
	SetStateVector(x *mat.Dense)
}

// =============================================================================
// FilterPyKalmanFilter - Full Kalman Filter Implementation
// =============================================================================

// FilterPyKalmanFilter wraps the internal filterpy.KalmanFilter to satisfy the Filter interface
type FilterPyKalmanFilter struct {
	*filterpy.KalmanFilter
}

// FilterPyKalmanFilterFactory creates FilterPyKalmanFilter instances
type FilterPyKalmanFilterFactory struct {
	RMult float64 // Multiplier for sensor measurement noise matrix
	QMult float64 // Multiplier for process uncertainty
	PMult float64 // Multiplier for initial covariance matrix (position entries)
}

// NewFilterPyKalmanFilterFactory creates a factory with default parameters
func NewFilterPyKalmanFilterFactory(rMult, qMult, pMult float64) *FilterPyKalmanFilterFactory {
	return &FilterPyKalmanFilterFactory{
		RMult: rMult,
		QMult: qMult,
		PMult: pMult,
	}
}

// CreateFilter creates a new FilterPyKalmanFilter instance
func (f *FilterPyKalmanFilterFactory) CreateFilter(initialDetection *mat.Dense) Filter {
	numPoints, dimPoints := initialDetection.Dims()
	dimZ := numPoints * dimPoints
	dimX := 2 * dimZ

	// Create internal Kalman filter
	kf := filterpy.NewKalmanFilter(dimX, dimZ)

	// Initialize F (state transition matrix)
	// F = [[I, dt*I], [0, I]] where dt = 1
	F := kf.GetF()
	for i := 0; i < dimX; i++ {
		F.Set(i, i, 1.0)
	}
	dt := 1.0
	for i := 0; i < dimZ; i++ {
		F.Set(i, dimZ+i, dt)
	}

	// Initialize H (measurement matrix) = [I, 0]
	H := kf.GetH()
	for i := 0; i < dimZ; i++ {
		H.Set(i, i, 1.0)
	}

	// Initialize R (measurement noise) - identity * RMult
	R := kf.GetR()
	for i := 0; i < dimZ; i++ {
		R.Set(i, i, f.RMult)
	}

	// Initialize Q (process noise)
	// Start with identity, then multiply velocity part
	Q := kf.GetQ()
	for i := 0; i < dimX; i++ {
		Q.Set(i, i, 1.0)
	}
	// Multiply only velocity part by QMult
	for i := dimZ; i < dimX; i++ {
		Q.Set(i, i, Q.At(i, i)*f.QMult)
	}

	// Initialize state x
	// Position part = flattened initial detection
	// Velocity part = 0
	x := kf.GetX()
	flatDetection := flattenDetection(initialDetection)
	for i := 0; i < dimZ; i++ {
		x.Set(i, 0, flatDetection[i])
	}

	// Initialize P (covariance)
	// Start with identity
	P := kf.GetP()
	for i := 0; i < dimX; i++ {
		P.Set(i, i, 1.0)
	}
	// Multiply position entries by PMult (velocity stays at 1.0)
	// This matches filterpy: P = [[p*I_pos, 0], [0, I_vel]]
	for i := 0; i < dimZ; i++ {
		P.Set(i, i, f.PMult)
	}

	return &FilterPyKalmanFilter{KalmanFilter: kf}
}

// GetStateVector returns the state vector (wrapper for GetState to satisfy Filter interface)
func (kf *FilterPyKalmanFilter) GetStateVector() *mat.Dense {
	return kf.GetState()
}

// SetStateVector sets the state vector (wrapper for SetState to satisfy Filter interface)
func (kf *FilterPyKalmanFilter) SetStateVector(x *mat.Dense) {
	kf.SetState(x)
}

// =============================================================================
// NoFilter - Simple No-Op Filter
// =============================================================================

// NoFilter implements a filter that does no prediction
type NoFilter struct {
	dimX int
	dimZ int
	x    *mat.Dense
}

// NoFilterFactory creates NoFilter instances
type NoFilterFactory struct{}

// NewNoFilterFactory creates a new NoFilterFactory
func NewNoFilterFactory() *NoFilterFactory {
	return &NoFilterFactory{}
}

func (f *NoFilterFactory) CreateFilter(initialDetection *mat.Dense) Filter {
	numPoints, dimPoints := initialDetection.Dims()
	dimZ := numPoints * dimPoints
	dimX := 2 * dimZ

	filter := &NoFilter{
		dimX: dimX,
		dimZ: dimZ,
		x:    mat.NewDense(dimX, 1, nil),
	}

	// Initialize state x
	flatDetection := flattenDetection(initialDetection)
	for i := 0; i < dimZ; i++ {
		filter.x.Set(i, 0, flatDetection[i])
	}

	return filter
}

func (nf *NoFilter) Predict() {
	// No-op
}

func (nf *NoFilter) Update(detectionPointsFlatten *mat.Dense, R, H *mat.Dense) {
	// Extract diagonal from H if provided
	var diagonal []float64
	var oneMinusDiagonal []float64

	if H != nil {
		rows, _ := H.Dims()
		diagonal = make([]float64, rows)
		oneMinusDiagonal = make([]float64, rows)
		for i := 0; i < rows; i++ {
			diagonal[i] = H.At(i, i)
			oneMinusDiagonal[i] = 1.0 - diagonal[i]
		}
	} else {
		// Default: all ones (all measurements valid)
		diagonal = make([]float64, nf.dimZ)
		oneMinusDiagonal = make([]float64, nf.dimZ)
		for i := 0; i < nf.dimZ; i++ {
			diagonal[i] = 1.0
			oneMinusDiagonal[i] = 0.0
		}
	}

	// Update: detection_points_flatten = diagonal * detection + (1-diagonal) * x[:dimZ]
	for i := 0; i < nf.dimZ; i++ {
		newVal := diagonal[i]*detectionPointsFlatten.At(i, 0) + oneMinusDiagonal[i]*nf.x.At(i, 0)
		nf.x.Set(i, 0, newVal)
	}
}

func (nf *NoFilter) GetState() *mat.Dense {
	return nf.x
}

func (nf *NoFilter) GetDimZ() int {
	return nf.dimZ
}

func (nf *NoFilter) GetStateVector() *mat.Dense {
	return nf.x
}

func (nf *NoFilter) SetStateVector(x *mat.Dense) {
	nf.x.Copy(x)
}

// =============================================================================
// OptimizedKalmanFilter - Fast Simplified Implementation
// =============================================================================

// OptimizedKalmanFilter implements an optimized Kalman filter
type OptimizedKalmanFilter struct {
	dimX int
	dimZ int
	x    *mat.Dense

	// Simplified covariance representation (vectors instead of matrices)
	PosVariance      []float64
	PosVelCovariance []float64
	VelVariance      []float64
	qQ               float64
	defaultR         []float64
}

// OptimizedKalmanFilterFactory creates OptimizedKalmanFilter instances
type OptimizedKalmanFilterFactory struct {
	RMult             float64
	QMult             float64
	PosVariance       float64
	PosVelCovariance  float64
	VelVariance       float64
}

// NewOptimizedKalmanFilterFactory creates a factory with default parameters
func NewOptimizedKalmanFilterFactory(rMult, qMult, posVar, posVelCov, velVar float64) *OptimizedKalmanFilterFactory {
	return &OptimizedKalmanFilterFactory{
		RMult:            rMult,
		QMult:            qMult,
		PosVariance:      posVar,
		PosVelCovariance: posVelCov,
		VelVariance:      velVar,
	}
}

func (f *OptimizedKalmanFilterFactory) CreateFilter(initialDetection *mat.Dense) Filter {
	numPoints, dimPoints := initialDetection.Dims()
	dimZ := numPoints * dimPoints
	dimX := 2 * dimZ

	filter := &OptimizedKalmanFilter{
		dimX:             dimX,
		dimZ:             dimZ,
		x:                mat.NewDense(dimX, 1, nil),
		PosVariance:      make([]float64, dimZ),
		PosVelCovariance: make([]float64, dimZ),
		VelVariance:      make([]float64, dimZ),
		qQ:               f.QMult,
		defaultR:         make([]float64, dimZ),
	}

	// Initialize covariance vectors
	for i := 0; i < dimZ; i++ {
		filter.PosVariance[i] = f.PosVariance
		filter.PosVelCovariance[i] = f.PosVelCovariance
		filter.VelVariance[i] = f.VelVariance
		filter.defaultR[i] = f.RMult
	}

	// Initialize state x
	flatDetection := flattenDetection(initialDetection)
	for i := 0; i < dimZ; i++ {
		filter.x.Set(i, 0, flatDetection[i])
	}

	return filter
}

func (okf *OptimizedKalmanFilter) Predict() {
	// x[:dimZ] += x[dimZ:]
	for i := 0; i < okf.dimZ; i++ {
		newPos := okf.x.At(i, 0) + okf.x.At(okf.dimZ+i, 0)
		okf.x.Set(i, 0, newPos)
	}
}

func (okf *OptimizedKalmanFilter) Update(detectionPointsFlatten *mat.Dense, R, H *mat.Dense) {
	// Extract diagonal from H and R
	var diagonal []float64
	var oneMinusDiagonal []float64
	var kalmanR []float64

	if H != nil {
		rows, _ := H.Dims()
		diagonal = make([]float64, rows)
		oneMinusDiagonal = make([]float64, rows)
		for i := 0; i < rows; i++ {
			diagonal[i] = H.At(i, i)
			oneMinusDiagonal[i] = 1.0 - diagonal[i]
		}
	} else {
		diagonal = make([]float64, okf.dimZ)
		oneMinusDiagonal = make([]float64, okf.dimZ)
		for i := 0; i < okf.dimZ; i++ {
			diagonal[i] = 1.0
			oneMinusDiagonal[i] = 0.0
		}
	}

	if R != nil {
		rows, _ := R.Dims()
		kalmanR = make([]float64, rows)
		for i := 0; i < rows; i++ {
			kalmanR[i] = R.At(i, i)
		}
	} else {
		kalmanR = make([]float64, okf.dimZ)
		copy(kalmanR, okf.defaultR)
	}

	// Compute error (innovation)
	error := make([]float64, okf.dimZ)
	for i := 0; i < okf.dimZ; i++ {
		error[i] = (detectionPointsFlatten.At(i, 0) - okf.x.At(i, 0)) * diagonal[i]
	}

	// Compute Kalman gains
	velVarPlusPosVelCov := make([]float64, okf.dimZ)
	addedVariances := make([]float64, okf.dimZ)
	for i := 0; i < okf.dimZ; i++ {
		velVarPlusPosVelCov[i] = okf.PosVelCovariance[i] + okf.VelVariance[i]
		addedVariances[i] = okf.PosVariance[i] + okf.PosVelCovariance[i] +
			velVarPlusPosVelCov[i] + okf.qQ + kalmanR[i]
	}

	kalmanROverAddedVariances := make([]float64, okf.dimZ)
	velVarPlusPosVelCovOverAddedVariances := make([]float64, okf.dimZ)
	for i := 0; i < okf.dimZ; i++ {
		kalmanROverAddedVariances[i] = kalmanR[i] / addedVariances[i]
		velVarPlusPosVelCovOverAddedVariances[i] = velVarPlusPosVelCov[i] / addedVariances[i]
	}

	addedVariancesOrKalmanR := make([]float64, okf.dimZ)
	for i := 0; i < okf.dimZ; i++ {
		addedVariancesOrKalmanR[i] = addedVariances[i]*oneMinusDiagonal[i] + kalmanR[i]*diagonal[i]
	}

	// Update state
	for i := 0; i < okf.dimZ; i++ {
		okf.x.Set(i, 0, okf.x.At(i, 0)+diagonal[i]*(1.0-kalmanROverAddedVariances[i])*error[i])
	}
	for i := 0; i < okf.dimZ; i++ {
		okf.x.Set(okf.dimZ+i, 0, okf.x.At(okf.dimZ+i, 0)+
			diagonal[i]*velVarPlusPosVelCovOverAddedVariances[i]*error[i])
	}

	// Update covariance vectors
	for i := 0; i < okf.dimZ; i++ {
		okf.PosVariance[i] = (1.0 - kalmanROverAddedVariances[i]) * addedVariancesOrKalmanR[i]
		okf.PosVelCovariance[i] = velVarPlusPosVelCovOverAddedVariances[i] * addedVariancesOrKalmanR[i]
		okf.VelVariance[i] += okf.qQ - diagonal[i]*
			velVarPlusPosVelCovOverAddedVariances[i]*velVarPlusPosVelCovOverAddedVariances[i]*
			addedVariances[i]
	}
}

func (okf *OptimizedKalmanFilter) GetState() *mat.Dense {
	return okf.x
}

func (okf *OptimizedKalmanFilter) GetDimZ() int {
	return okf.dimZ
}

func (okf *OptimizedKalmanFilter) GetStateVector() *mat.Dense {
	return okf.x
}

func (okf *OptimizedKalmanFilter) SetStateVector(x *mat.Dense) {
	okf.x.Copy(x)
}

// =============================================================================
// Helper Functions
// =============================================================================

// flattenDetection flattens a detection matrix into a 1D slice
func flattenDetection(detection *mat.Dense) []float64 {
	rows, cols := detection.Dims()
	flat := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			flat[i*cols+j] = detection.At(i, j)
		}
	}
	return flat
}
