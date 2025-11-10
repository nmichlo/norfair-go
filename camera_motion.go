package norfairgo

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// CoordinateTransformation is an interface for transforming between relative and absolute coordinates.
// This is used for camera motion compensation in tracking.
//
// Detections' and tracked objects' coordinates can be interpreted in 2 references:
// - Relative: their position on the current frame, (0, 0) is top left
// - Absolute: their position in a fixed space, (0, 0) is the top left of the first frame
//
// Therefore, coordinate transformation in this context is an interface that can transform
// coordinates from one reference to another.
type CoordinateTransformation interface {
	// RelToAbs transforms points from relative (camera frame) to absolute (world frame) coordinates
	RelToAbs(points *mat.Dense) *mat.Dense

	// AbsToRel transforms points from absolute (world frame) to relative (camera frame) coordinates
	AbsToRel(points *mat.Dense) *mat.Dense
}

// TransformationGetter is an interface for finding CoordinateTransformation between 2 sets of points.
// It takes current and previous point correspondences and returns whether the reference frame
// should be updated and the computed transformation.
type TransformationGetter interface {
	// Call computes the transformation between current and previous points.
	// Returns: (shouldUpdateReference, transformation)
	Call(currPts, prevPts *mat.Dense) (bool, CoordinateTransformation)
}

// NilCoordinateTransformation is a no-op transformation that returns points unchanged.
// This is used when camera motion is not being tracked.
type NilCoordinateTransformation struct{}

// RelToAbs returns points unchanged
func (n *NilCoordinateTransformation) RelToAbs(points *mat.Dense) *mat.Dense {
	return points
}

// AbsToRel returns points unchanged
func (n *NilCoordinateTransformation) AbsToRel(points *mat.Dense) *mat.Dense {
	return points
}

//
// Translation Implementation
//

// TranslationTransformation represents a simple 2D translation (camera pan/tilt without rotation/zoom).
type TranslationTransformation struct {
	MovementVector []float64 // [dx, dy] translation vector
}

// NewTranslationTransformation creates a new translation transformation with the given movement vector.
func NewTranslationTransformation(movementVector []float64) (*TranslationTransformation, error) {
	if len(movementVector) != 2 {
		return nil, fmt.Errorf("movement vector must have exactly 2 elements, got %d", len(movementVector))
	}
	return &TranslationTransformation{
		MovementVector: movementVector,
	}, nil
}

// AbsToRel converts absolute coordinates to relative by adding the movement vector.
// In Python: points + movement_vector
func (t *TranslationTransformation) AbsToRel(points *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	if cols != 2 {
		// Return unchanged if not 2D points
		return points
	}

	result := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		result.Set(i, 0, points.At(i, 0)+t.MovementVector[0])
		result.Set(i, 1, points.At(i, 1)+t.MovementVector[1])
	}
	return result
}

// RelToAbs converts relative coordinates to absolute by subtracting the movement vector.
// In Python: points - movement_vector
func (t *TranslationTransformation) RelToAbs(points *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	if cols != 2 {
		// Return unchanged if not 2D points
		return points
	}

	result := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		result.Set(i, 0, points.At(i, 0)-t.MovementVector[0])
		result.Set(i, 1, points.At(i, 1)-t.MovementVector[1])
	}
	return result
}

// TranslationTransformationGetter calculates TranslationTransformation between points using optical flow mode.
//
// The camera movement is calculated as the mode of optical flow between the previous reference frame
// and the current frame. Comparing consecutive frames can make differences too small to correctly
// estimate the translation, so the reference frame is kept fixed as we progress through the video.
// Eventually, if the transformation can no longer match enough points, the reference frame is updated.
type TranslationTransformationGetter struct {
	// BinSize is the granularity for flow bucketing before calculating the mode.
	// Optical flow is rounded to the nearest bin_size before finding the mode.
	BinSize float64

	// ProportionPointsUsedThreshold is the minimum proportion of points that must be matched.
	// If the proportion falls below this threshold, the reference frame is updated.
	ProportionPointsUsedThreshold float64

	// data stores the accumulated transformation from the original reference frame.
	// nil on first call, then accumulates translations.
	data *[]float64
}

// NewTranslationTransformationGetter creates a new translation transformation getter.
func NewTranslationTransformationGetter(binSize, proportionPointsUsedThreshold float64) *TranslationTransformationGetter {
	return &TranslationTransformationGetter{
		BinSize:                       binSize,
		ProportionPointsUsedThreshold: proportionPointsUsedThreshold,
		data:                          nil,
	}
}

// Call computes the translation transformation between current and previous points.
// Returns (shouldUpdateReference, transformation).
//
// Algorithm:
// 1. Calculate optical flow: flow = currPts - prevPts
// 2. Bin the flow vectors (round to bin_size)
// 3. Find mode (most common flow vector)
// 4. Check if proportion of points using mode is above threshold
// 5. Accumulate with previous transformation if reference frame not updated
func (t *TranslationTransformationGetter) Call(currPts, prevPts *mat.Dense) (bool, CoordinateTransformation) {
	currRows, currCols := currPts.Dims()
	prevRows, prevCols := prevPts.Dims()

	if currRows != prevRows || currCols != prevCols {
		// Invalid input, return nil transformation
		return true, &TranslationTransformation{MovementVector: []float64{0, 0}}
	}

	if currCols != 2 {
		// Not 2D points, return nil transformation
		return true, &TranslationTransformation{MovementVector: []float64{0, 0}}
	}

	// Step 1: Calculate flow = currPts - prevPts
	flow := make([][]float64, currRows)
	for i := 0; i < currRows; i++ {
		flow[i] = []float64{
			currPts.At(i, 0) - prevPts.At(i, 0),
			currPts.At(i, 1) - prevPts.At(i, 1),
		}
	}

	// Step 2: Bin the flow vectors (round to nearest bin_size)
	binnedFlow := make([][]float64, currRows)
	for i := 0; i < currRows; i++ {
		binnedFlow[i] = []float64{
			math.Round(flow[i][0]/t.BinSize) * t.BinSize,
			math.Round(flow[i][1]/t.BinSize) * t.BinSize,
		}
	}

	// Step 3: Find mode (most common flow vector)
	// Group by binned flow and count occurrences
	flowCounts := make(map[string]int)
	flowVectors := make(map[string][]float64)

	for _, f := range binnedFlow {
		// Use string key for map (floats aren't directly comparable)
		key := fmt.Sprintf("%.10f,%.10f", f[0], f[1])
		flowCounts[key]++
		if _, exists := flowVectors[key]; !exists {
			flowVectors[key] = f
		}
	}

	// Find the flow with maximum count
	var maxKey string
	maxCount := 0
	for key, count := range flowCounts {
		if count > maxCount {
			maxCount = count
			maxKey = key
		}
	}

	flowMode := flowVectors[maxKey]

	// Step 4: Check proportion of points using the mode
	proportionPointsUsed := float64(maxCount) / float64(currRows)
	updatePrvs := proportionPointsUsed < t.ProportionPointsUsedThreshold

	// Step 5: Accumulate with previous transformation if available
	if t.data != nil {
		flowMode[0] += (*t.data)[0]
		flowMode[1] += (*t.data)[1]
	}

	// Update accumulated data if reference frame should be updated
	if updatePrvs {
		t.data = &flowMode
	}

	transformation, _ := NewTranslationTransformation(flowMode)
	return updatePrvs, transformation
}

//
// Homography Implementation
//

// HomographyTransformation represents a full perspective transformation using a 3x3 homography matrix.
// This handles camera rotation, scaling, skew, and perspective effects.
type HomographyTransformation struct {
	HomographyMatrix        *mat.Dense // 3x3 transformation matrix
	InverseHomographyMatrix *mat.Dense // Pre-computed inverse for efficiency
}

// NewHomographyTransformation creates a new homography transformation with the given matrix.
// The matrix must be 3x3. The inverse is pre-computed for efficiency.
func NewHomographyTransformation(homographyMatrix *mat.Dense) (*HomographyTransformation, error) {
	rows, cols := homographyMatrix.Dims()
	if rows != 3 || cols != 3 {
		return nil, fmt.Errorf("homography matrix must be 3x3, got %dx%d", rows, cols)
	}

	// Compute inverse matrix
	var inverse mat.Dense
	err := inverse.Inverse(homographyMatrix)
	if err != nil {
		return nil, fmt.Errorf("cannot invert homography matrix: %v", err)
	}

	return &HomographyTransformation{
		HomographyMatrix:        homographyMatrix,
		InverseHomographyMatrix: &inverse,
	}, nil
}

// AbsToRel converts absolute coordinates to relative using the homography matrix.
// In Python: points_with_ones @ self.homography_matrix.T, then perspective division
func (h *HomographyTransformation) AbsToRel(points *mat.Dense) *mat.Dense {
	return h.transformPoints(points, h.HomographyMatrix)
}

// RelToAbs converts relative coordinates to absolute using the inverse homography matrix.
// In Python: points_with_ones @ self.inverse_homography_matrix.T, then perspective division
func (h *HomographyTransformation) RelToAbs(points *mat.Dense) *mat.Dense {
	return h.transformPoints(points, h.InverseHomographyMatrix)
}

// transformPoints applies a homography transformation to 2D points.
// Algorithm:
// 1. Convert to homogeneous coordinates: [x, y] → [x, y, 1]
// 2. Matrix multiply: [x', y', w'] = [x, y, 1] @ H^T
// 3. Perspective division: x'' = x'/w', y'' = y'/w'
// 4. Handle division by zero: if w' == 0, set to 0.0000001
// 5. Return 2D points: [x'', y'']
func (h *HomographyTransformation) transformPoints(points *mat.Dense, transformMatrix *mat.Dense) *mat.Dense {
	rows, cols := points.Dims()
	if cols != 2 {
		// Return unchanged if not 2D points
		return points
	}

	// Step 1: Convert to homogeneous coordinates (add column of 1s)
	pointsHomogeneous := mat.NewDense(rows, 3, nil)
	for i := 0; i < rows; i++ {
		pointsHomogeneous.Set(i, 0, points.At(i, 0))
		pointsHomogeneous.Set(i, 1, points.At(i, 1))
		pointsHomogeneous.Set(i, 2, 1.0)
	}

	// Step 2: Matrix multiply: points_homogeneous @ transformMatrix.T
	// Get transpose of transform matrix
	transformMatrixT := transformMatrix.T()

	var pointsTransformed mat.Dense
	pointsTransformed.Mul(pointsHomogeneous, transformMatrixT)

	// Step 3: Get last column (w values) and handle division by zero
	lastColumn := make([]float64, rows)
	for i := 0; i < rows; i++ {
		w := pointsTransformed.At(i, 2)
		if w == 0 {
			w = 0.0000001
		}
		lastColumn[i] = w
	}

	// Step 4: Perspective division - divide all columns by w
	result := mat.NewDense(rows, 2, nil)
	for i := 0; i < rows; i++ {
		result.Set(i, 0, pointsTransformed.At(i, 0)/lastColumn[i])
		result.Set(i, 1, pointsTransformed.At(i, 1)/lastColumn[i])
	}

	return result
}

// HomographyTransformationGetter calculates HomographyTransformation between points using RANSAC.
//
// The camera movement is represented as a homography that matches the optical flow between
// the previous reference frame and the current. Comparing consecutive frames can make differences
// too small to correctly estimate the homography, so the reference frame is kept fixed as we
// progress through the video. Eventually, if the transformation can no longer match enough points,
// the reference frame is updated.
type HomographyTransformationGetter struct {
	// Method is the OpenCV method for finding homographies.
	// Valid options: gocv.HomographyMethodRANSAC (default), gocv.HomographyMethodLMEDS, gocv.HomographyMethodRHO
	Method gocv.HomographyMethod

	// RansacReprojThreshold is the maximum allowed reprojection error to treat a point pair as an inlier.
	RansacReprojThreshold float64

	// MaxIters is the maximum number of RANSAC iterations.
	MaxIters int

	// Confidence is the RANSAC confidence level (between 0 and 1).
	Confidence float64

	// ProportionPointsUsedThreshold is the minimum proportion of points that must be matched.
	// If the proportion falls below this threshold, the reference frame is updated.
	ProportionPointsUsedThreshold float64

	// data stores the accumulated homography from the original reference frame.
	// nil on first call, then accumulates homographies via matrix multiplication.
	data *mat.Dense
}

// NewHomographyTransformationGetter creates a new homography transformation getter with RANSAC.
func NewHomographyTransformationGetter(ransacReprojThreshold float64, maxIters int, confidence, proportionPointsUsedThreshold float64) *HomographyTransformationGetter {
	return &HomographyTransformationGetter{
		Method:                        gocv.HomographyMethodRANSAC,
		RansacReprojThreshold:        ransacReprojThreshold,
		MaxIters:                      maxIters,
		Confidence:                    confidence,
		ProportionPointsUsedThreshold: proportionPointsUsedThreshold,
		data:                          nil,
	}
}

// Call computes the homography transformation between current and previous points using RANSAC.
// Returns (shouldUpdateReference, transformation).
//
// Algorithm:
// 1. Validate minimum 4 points (homography requires ≥4 correspondences)
// 2. Call gocv.FindHomography() with RANSAC
// 3. Count inliers and check proportion
// 4. Accumulate homographies via matrix multiplication (NOT addition!)
// 5. Determine if reference frame should be updated
func (h *HomographyTransformationGetter) Call(currPts, prevPts *mat.Dense) (bool, CoordinateTransformation) {
	currRows, currCols := currPts.Dims()
	prevRows, prevCols := prevPts.Dims()

	// Validate minimum points and dimensions
	if currRows < 4 || prevRows < 4 || currCols != 2 || prevCols != 2 {
		log.Printf("Warning: Homography couldn't be computed due to insufficient points (need ≥4, got curr=%d, prev=%d)", currRows, prevRows)

		// Return previous transformation if available
		if h.data != nil {
			trans, _ := NewHomographyTransformation(h.data)
			return true, trans
		}
		return true, nil
	}

	// Convert gonum matrices to gocv Mat
	prevPtsGocv := matDenseToGocvMat(prevPts)
	currPtsGocv := matDenseToGocvMat(currPts)
	defer prevPtsGocv.Close()
	defer currPtsGocv.Close()

	// Call gocv.FindHomography with RANSAC
	mask := gocv.NewMat()
	defer mask.Close()

	homographyMat := gocv.FindHomography(
		prevPtsGocv,
		currPtsGocv,
		h.Method,
		h.RansacReprojThreshold,
		&mask,
		h.MaxIters,
		h.Confidence,
	)
	defer homographyMat.Close()

	// Check if homography computation failed
	if homographyMat.Empty() {
		log.Printf("Warning: FindHomography returned empty matrix")
		if h.data != nil {
			trans, _ := NewHomographyTransformation(h.data)
			return true, trans
		}
		return true, nil
	}

	// Convert gocv.Mat (3x3) to gonum *mat.Dense
	homographyMatrix := gocvMatToMatDense(homographyMat)

	// Count inliers from mask
	inlierCount := gocv.CountNonZero(mask)
	totalPoints := prevRows
	proportionPointsUsed := float64(inlierCount) / float64(totalPoints)

	// Determine if reference frame should be updated
	updatePrvs := proportionPointsUsed < h.ProportionPointsUsedThreshold

	// Accumulate homographies via matrix multiplication (NOT addition!)
	// Python: homography_matrix = homography_matrix @ self.data
	if h.data != nil {
		var accumulated mat.Dense
		accumulated.Mul(homographyMatrix, h.data)
		homographyMatrix = &accumulated
	}

	// Update accumulated data if reference frame should be updated
	if updatePrvs {
		h.data = homographyMatrix
	}

	// Create and return transformation
	transformation, err := NewHomographyTransformation(homographyMatrix)
	if err != nil {
		log.Printf("Warning: Failed to create HomographyTransformation: %v", err)
		if h.data != nil {
			trans, _ := NewHomographyTransformation(h.data)
			return true, trans
		}
		return true, nil
	}

	return updatePrvs, transformation
}

//
// gocv Conversion Helpers
//

// matDenseToGocvMat converts a gonum *mat.Dense (Nx2) to gocv.Mat for FindHomography.
// The input matrix should have shape (N, 2) where each row is an (x, y) point.
// Returns a CV_32FC2 Mat (2-channel float32).
func matDenseToGocvMat(m *mat.Dense) gocv.Mat {
	rows, _ := m.Dims()

	// Create a flat array of float32 data in interleaved format: [x1, y1, x2, y2, ...]
	data := make([]float32, rows*2)
	for i := 0; i < rows; i++ {
		data[i*2] = float32(m.At(i, 0))     // x
		data[i*2+1] = float32(m.At(i, 1))   // y
	}

	// Create CV_32FC2 Mat from bytes
	// CV_32FC2 means 2-channel float32, which is exactly what FindHomography expects
	result, err := gocv.NewMatFromBytes(rows, 1, gocv.MatTypeCV32FC2, toBytes(data))
	if err != nil {
		log.Printf("Error creating Mat from bytes: %v", err)
		return gocv.NewMat()
	}

	return result
}

// toBytes converts a slice of float32 to a slice of bytes
func toBytes(data []float32) []byte {
	bytes := make([]byte, len(data)*4) // 4 bytes per float32
	for i, v := range data {
		bits := math.Float32bits(v)
		bytes[i*4] = byte(bits)
		bytes[i*4+1] = byte(bits >> 8)
		bytes[i*4+2] = byte(bits >> 16)
		bytes[i*4+3] = byte(bits >> 24)
	}
	return bytes
}

// gocvMatToMatDense converts a gocv.Mat (3x3 homography matrix) to gonum *mat.Dense.
// The input should be a CV_64F or CV_32F matrix.
func gocvMatToMatDense(m gocv.Mat) *mat.Dense {
	rows := m.Rows()
	cols := m.Cols()

	data := make([]float64, rows*cols)

	// gocv stores matrices row-major
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// GetDoubleAt works for both CV_64F and CV_32F
			data[i*cols+j] = m.GetDoubleAt(i, j)
		}
	}

	return mat.NewDense(rows, cols, data)
}

//
// Motion Estimator
//

// MotionEstimator tracks camera motion across video frames using optical flow.
// It maintains a reference frame and tracks feature points between frames to compute
// coordinate transformations for camera motion compensation.
type MotionEstimator struct {
	// Corner detection parameters
	MaxPoints    int     // Maximum number of corner points to sample for optical flow
	MinDistance  int     // Minimum distance between sampled points
	BlockSize    int     // Size of averaging block for corner detection
	QualityLevel float64 // Minimal accepted quality for corner detection (0.0 to 1.0)

	// Transformation computation
	TransformationsGetter TransformationGetter // Strategy for computing coordinate transformations

	// Optional flow visualization
	DrawFlow  bool        // Enable visual debugging by drawing optical flow vectors
	FlowColor color.RGBA  // Color for flow visualization

	// Internal state
	grayPrvs                   gocv.Mat    // Reference frame (grayscale)
	grayNext                   gocv.Mat    // Current frame (grayscale)
	prevPts                    *mat.Dense  // Points from the previous reference frame
	prevMask                   gocv.Mat    // Mask from the previous reference frame
	transformationsGetterCopy  TransformationGetter // Deep copy for error recovery
}

// NewMotionEstimator creates a new MotionEstimator with the specified parameters.
// If transformationsGetter is nil, it defaults to HomographyTransformationGetter.
func NewMotionEstimator(
	maxPoints int,
	minDistance int,
	blockSize int,
	qualityLevel float64,
	transformationsGetter TransformationGetter,
	drawFlow bool,
	flowColor *color.RGBA,
) *MotionEstimator {
	// Default to HomographyTransformationGetter if nil
	if transformationsGetter == nil {
		transformationsGetter = NewHomographyTransformationGetter(3.0, 2000, 0.995, 0.9)
	}

	// Default flow color to blue if nil and drawFlow is true
	var flowCol color.RGBA
	if flowColor != nil {
		flowCol = *flowColor
	} else if drawFlow {
		flowCol = color.RGBA{R: 0, G: 0, B: 255, A: 0} // Blue
	}

	// TODO: Create deep copy of transformationsGetter for error recovery
	// For now, just use the same instance
	transformationsGetterCopy := transformationsGetter

	return &MotionEstimator{
		MaxPoints:                  maxPoints,
		MinDistance:                minDistance,
		BlockSize:                  blockSize,
		QualityLevel:               qualityLevel,
		TransformationsGetter:      transformationsGetter,
		DrawFlow:                   drawFlow,
		FlowColor:                  flowCol,
		grayPrvs:                   gocv.NewMat(),
		grayNext:                   gocv.NewMat(),
		prevPts:                    nil,
		prevMask:                   gocv.NewMat(),
		transformationsGetterCopy:  transformationsGetterCopy,
	}
}

// Close releases resources held by the MotionEstimator.
// Must be called when done using the MotionEstimator.
// Safe to call multiple times.
func (m *MotionEstimator) Close() {
	// Close grayPrvs if not already closed
	if m.grayPrvs.Ptr() != nil {
		if !m.grayPrvs.Empty() {
			m.grayPrvs.Close()
		}
		m.grayPrvs = gocv.NewMat()
	}

	// Close grayNext if not already closed
	if m.grayNext.Ptr() != nil {
		if !m.grayNext.Empty() {
			m.grayNext.Close()
		}
		m.grayNext = gocv.NewMat()
	}

	// Close prevMask if not already closed
	if m.prevMask.Ptr() != nil {
		if !m.prevMask.Empty() {
			m.prevMask.Close()
		}
		m.prevMask = gocv.NewMat()
	}
}

// getSparseFlow computes sparse optical flow between two frames.
// If prevPts is nil, it detects new corner points in grayPrvs.
// Returns matched point pairs (currPts, prevPts) as gonum matrices.
func (m *MotionEstimator) getSparseFlow(mask gocv.Mat) (*mat.Dense, *mat.Dense, error) {
	// Step 1: Detect corner points if we don't have previous points
	var prevPtsGocv gocv.Mat
	if m.prevPts == nil {
		// Use goodFeaturesToTrack to find corners
		corners := gocv.NewMat()
		defer corners.Close()

		gocv.GoodFeaturesToTrack(
			m.grayPrvs,
			&corners,
			m.MaxPoints,
			m.QualityLevel,
			float64(m.MinDistance),
		)

		// Apply mask if provided
		if !mask.Empty() {
			// TODO: Implement mask filtering for corners
			// For now, use all detected corners
		}

		if corners.Rows() == 0 {
			return nil, nil, fmt.Errorf("no corners detected")
		}

		prevPtsGocv = corners
	} else {
		// Convert previous points from gonum to gocv format
		prevPtsGocv = matDenseToGocvMat(m.prevPts)
		defer prevPtsGocv.Close()
	}

	// Step 2: Track points using optical flow
	currPtsGocv := gocv.NewMat()
	defer currPtsGocv.Close()

	status := gocv.NewMat()
	defer status.Close()

	errMat := gocv.NewMat()
	defer errMat.Close()

	// Calculate optical flow (Lucas-Kanade with pyramids)
	gocv.CalcOpticalFlowPyrLK(
		m.grayPrvs,
		m.grayNext,
		prevPtsGocv,
		currPtsGocv,
		&status,
		&errMat,
	)

	// Step 3: Filter to successfully tracked points (status == 1)
	var prevFiltered []float64
	var currFiltered []float64
	numPoints := 0

	for i := 0; i < status.Rows(); i++ {
		if status.GetUCharAt(i, 0) == 1 {
			// Get previous point
			prevVec := prevPtsGocv.GetVecfAt(i, 0)
			prevFiltered = append(prevFiltered, float64(prevVec[0]), float64(prevVec[1]))

			// Get current point
			currVec := currPtsGocv.GetVecfAt(i, 0)
			currFiltered = append(currFiltered, float64(currVec[0]), float64(currVec[1]))

			numPoints++
		}
	}

	if numPoints == 0 {
		return nil, nil, fmt.Errorf("no points successfully tracked")
	}

	// Convert to gonum matrices (N, 2)
	prevPtsMat := mat.NewDense(numPoints, 2, prevFiltered)
	currPtsMat := mat.NewDense(numPoints, 2, currFiltered)

	return currPtsMat, prevPtsMat, nil
}

// Update processes a new frame and computes the coordinate transformation for camera motion.
// Returns the transformation (or nil if it cannot be computed).
// The frame parameter is modified in-place if DrawFlow is enabled.
func (m *MotionEstimator) Update(frame gocv.Mat, mask gocv.Mat) CoordinateTransformation {
	// Step 1: Convert frame to grayscale
	gocv.CvtColor(frame, &m.grayNext, gocv.ColorBGRToGray)

	// Step 2: First frame initialization
	if m.grayPrvs.Empty() {
		m.grayNext.CopyTo(&m.grayPrvs)
		if !mask.Empty() {
			mask.CopyTo(&m.prevMask)
		}
		return nil // No transformation for first frame
	}

	// Step 3: Get sparse optical flow
	currPts, prevPts, err := m.getSparseFlow(mask)
	if err != nil {
		log.Printf("Warning: Optical flow calculation failed: %v", err)
		return nil
	}

	// Step 4: Optional flow visualization
	if m.DrawFlow && !frame.Empty() {
		m.drawOpticalFlow(frame, prevPts, currPts)
	}

	// Step 5: Compute transformation via TransformationsGetter
	var coordTransformations CoordinateTransformation
	updatePrvs := false

	// Try-catch around transformation calculation (error recovery)
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Warning: Transformation calculation failed: %v", r)
				// Restore from copy
				m.TransformationsGetter = m.transformationsGetterCopy
				coordTransformations = nil
			}
		}()

		updatePrvs, coordTransformations = m.TransformationsGetter.Call(currPts, prevPts)
	}()

	// Step 6: Handle reference frame update signal
	if updatePrvs {
		// Update reference frame
		m.grayNext.CopyTo(&m.grayPrvs)
		// Reset tracked points (will detect new corners on next frame)
		m.prevPts = nil
		// Update mask
		if !mask.Empty() {
			mask.CopyTo(&m.prevMask)
		} else {
			m.prevMask = gocv.NewMat()
		}
	} else {
		// Keep reference frame, update tracked points for next iteration
		m.prevPts = prevPts
	}

	// Step 7: Return transformation
	return coordTransformations
}

// drawOpticalFlow draws optical flow vectors on the frame for visualization.
// Modifies the frame in-place.
func (m *MotionEstimator) drawOpticalFlow(frame gocv.Mat, prevPts, currPts *mat.Dense) {
	numPoints, _ := prevPts.Dims()

	for i := 0; i < numPoints; i++ {
		// Get previous and current points
		prevX := int(prevPts.At(i, 0))
		prevY := int(prevPts.At(i, 1))
		currX := int(currPts.At(i, 0))
		currY := int(currPts.At(i, 1))

		// Draw line from previous to current position
		gocv.Line(
			&frame,
			image.Pt(prevX, prevY),
			image.Pt(currX, currY),
			m.FlowColor,
			2, // thickness
		)

		// Draw circle at current position
		gocv.Circle(
			&frame,
			image.Pt(currX, currY),
			3, // radius
			m.FlowColor,
			-1, // filled
		)
	}
}
