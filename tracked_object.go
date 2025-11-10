package norfairgo

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

// TrackedObject represents an object being tracked across frames.
type TrackedObject struct {
	// Configuration (shared reference to tracker config, immutable after creation)
	config *TrackerConfig

	// Factory reference
	objFactory *TrackedObjectFactory

	// Dimensions
	DimPoints int // Dimensionality per point (2 or 3)
	NumPoints int // Number of points

	// Runtime state (not in config)
	InitialPeriod int // Initial period value (stored for Merge())

	// State counters
	HitCounter      int   // Current hit counter (object-level)
	ReidHitCounter  *int  // Current ReID counter (nil until object dies)
	PointHitCounter []int // Per-point hit counters
	Age             int   // Age in frames
	IsInitializing  bool  // Whether still in initialization phase

	// IDs
	InitializingID *int // Temporary ID during initialization
	ID             *int // Permanent instance ID (nil until initialized)
	GlobalID       *int // Permanent global ID (nil until initialized)

	// Detection tracking
	LastDetection             *Detection   // Last matched detection
	LastDistance              *float64     // Distance from last match
	CurrentMinDistance        *float64     // Current minimum distance (debug)
	DetectedAtLeastOncePoints []bool       // Which points have been detected at least once
	PastDetections            []*Detection // Past detections stored

	// Filter
	Filter   Filter     // Kalman filter for state estimation
	DimZ     int        // Measurement dimension (dimPoints * numPoints)
	Estimate *mat.Dense // Cached position estimate (updated after filter operations)

	// Label and coordinate transform
	Label    *string                     // Class label
	AbsToRel func(*mat.Dense) *mat.Dense // Absolute to relative coordinate transform
}

// NewTrackedObject creates a new tracked object from an initial detection.
//
// Parameters:
//   - objFactory: Factory for ID generation
//   - initialDetection: First detection to initialize object
//   - config: Tracker configuration (shared reference, immutable)
//   - period: Current frame period (varies per update)
//   - coordTransformations: Coordinate transformation for camera motion (optional)
func NewTrackedObject(
	objFactory *TrackedObjectFactory,
	initialDetection *Detection,
	config *TrackerConfig,
	period int,
	coordTransformations CoordinateTransformation,
) (*TrackedObject, error) {
	// Validate initial detection
	if initialDetection == nil {
		return nil, fmt.Errorf("initial_detection must be a Detection instance")
	}

	// Get dimensions from detection
	rows, cols := initialDetection.AbsolutePoints.Dims()
	numPoints := rows
	dimPoints := cols

	if dimPoints != 2 && dimPoints != 3 {
		return nil, fmt.Errorf("detection points must have 2 or 3 dimensions, got %d", dimPoints)
	}

	dimZ := numPoints * dimPoints

	// Create tracked object
	to := &TrackedObject{
		config:             config, // Store shared config reference
		objFactory:         objFactory,
		DimPoints:          dimPoints,
		NumPoints:          numPoints,
		InitialPeriod:      period,
		HitCounter:         period, // Starts at period!
		ReidHitCounter:     nil,    // Not set until object dies
		Age:                0,
		LastDetection:      initialDetection,
		LastDistance:       nil,
		CurrentMinDistance: nil,
		DimZ:               dimZ,
		Label:              initialDetection.Label,
	}

	// Set initialization state
	to.IsInitializing = to.HitCounter <= to.config.InitializationDelay

	// Assign IDs
	initID := objFactory.GetInitializingID()
	to.InitializingID = &initID
	to.ID = nil
	to.GlobalID = nil
	if !to.IsInitializing {
		to.acquireIDs()
	}

	// Initialize point tracking
	to.DetectedAtLeastOncePoints = make([]bool, numPoints)
	to.PointHitCounter = make([]int, numPoints)

	if initialDetection.Scores == nil {
		// No scores - all points detected
		for i := 0; i < numPoints; i++ {
			to.DetectedAtLeastOncePoints[i] = true
			to.PointHitCounter[i] = 1
		}
	} else {
		// Use scores to determine detected points
		for i := 0; i < numPoints; i++ {
			if initialDetection.Scores[i] > to.config.DetectionThreshold {
				to.DetectedAtLeastOncePoints[i] = true
				to.PointHitCounter[i] = 1
			} else {
				to.DetectedAtLeastOncePoints[i] = false
				to.PointHitCounter[i] = 0
			}
		}
	}

	// Initialize past detections
	initialDetection.Age = to.Age
	if to.config.PastDetectionsLength > 0 {
		to.PastDetections = []*Detection{initialDetection}
	} else {
		to.PastDetections = []*Detection{}
	}

	// Create filter
	to.Filter = to.config.FilterFactory.CreateFilter(initialDetection.AbsolutePoints)

	// Set coordinate transformation BEFORE updating estimate
	// (estimate needs AbsToRel to convert from absolute to relative coords)
	if coordTransformations != nil {
		to.AbsToRel = coordTransformations.AbsToRel
	} else {
		to.AbsToRel = nil
	}

	// Initialize estimate from filter state
	to.updateEstimate()

	return to, nil
}

// TrackerStep is called once per frame for all tracked objects.
// It decrements counters, increments age, and calls filter prediction.
func (to *TrackedObject) TrackerStep() {
	// ReID counter management
	if to.ReidHitCounter == nil {
		// If object just died, initialize ReID counter
		if to.HitCounter <= 0 && to.config.ReidHitCounterMax != nil {
			reidCounter := *to.config.ReidHitCounterMax
			to.ReidHitCounter = &reidCounter
		}
	} else {
		// Decrement ReID counter
		*to.ReidHitCounter -= 1
	}

	// Decrement counters
	to.HitCounter -= 1
	for i := range to.PointHitCounter {
		to.PointHitCounter[i] -= 1
	}

	// Increment age
	to.Age += 1

	// Predict next state
	to.Filter.Predict()

	// Update cached estimate
	to.updateEstimate()
}

// Hit is called when the object is matched with a detection.
// It updates the Kalman filter and manages hit counters.
func (to *TrackedObject) Hit(detection *Detection, period int) error {
	to.conditionallyAddToPastDetections(detection)
	to.updateHitCounters(period)

	pointsOverThresholdMask, hPos := to.buildMeasurementMask(detection, period)
	H := to.buildFullHMatrix(hPos)
	detectionFlatten := to.flattenDetectionPoints(detection)

	to.Filter.Update(detectionFlatten, nil, H)
	to.handleFirstDetections(pointsOverThresholdMask, detectionFlatten)
	to.updateDetectedMask(pointsOverThresholdMask)
	to.updateEstimate()

	return nil
}

func (to *TrackedObject) updateHitCounters(period int) {
	to.HitCounter = min(to.HitCounter+2*period, to.config.HitCounterMax)

	if to.IsInitializing && to.HitCounter > to.config.InitializationDelay {
		to.IsInitializing = false
		to.acquireIDs()
	}
}

func (to *TrackedObject) buildMeasurementMask(detection *Detection, period int) ([]bool, *mat.Dense) {
	if detection.Scores != nil {
		return to.buildPartialMask(detection, period)
	}
	return to.buildFullMask(period)
}

func (to *TrackedObject) buildPartialMask(detection *Detection, period int) ([]bool, *mat.Dense) {
	pointsMask := make([]bool, to.NumPoints)
	sensorsMask := make([]float64, to.DimZ)

	for i := 0; i < to.NumPoints; i++ {
		pointsMask[i] = detection.Scores[i] > to.config.DetectionThreshold
		if pointsMask[i] {
			to.PointHitCounter[i] += 2 * period
			for d := 0; d < to.DimPoints; d++ {
				sensorsMask[i*to.DimPoints+d] = 1.0
			}
		}
	}

	to.clampPointHitCounters()

	hPos := mat.NewDense(to.DimZ, to.DimZ, nil)
	for i := 0; i < to.DimZ; i++ {
		hPos.Set(i, i, sensorsMask[i])
	}

	return pointsMask, hPos
}

func (to *TrackedObject) buildFullMask(period int) ([]bool, *mat.Dense) {
	pointsMask := make([]bool, to.NumPoints)
	for i := 0; i < to.NumPoints; i++ {
		pointsMask[i] = true
		to.PointHitCounter[i] += 2 * period
	}

	to.clampPointHitCounters()

	hPos := mat.NewDense(to.DimZ, to.DimZ, nil)
	for i := 0; i < to.DimZ; i++ {
		hPos.Set(i, i, 1.0)
	}

	return pointsMask, hPos
}

func (to *TrackedObject) clampPointHitCounters() {
	for i := 0; i < to.NumPoints; i++ {
		if to.PointHitCounter[i] >= to.config.PointwiseHitCounterMax {
			to.PointHitCounter[i] = to.config.PointwiseHitCounterMax
		}
		if to.PointHitCounter[i] < 0 {
			to.PointHitCounter[i] = 0
		}
	}
}

func (to *TrackedObject) buildFullHMatrix(hPos *mat.Dense) *mat.Dense {
	H := mat.NewDense(to.DimZ, 2*to.DimZ, nil)
	for i := 0; i < to.DimZ; i++ {
		for j := 0; j < to.DimZ; j++ {
			H.Set(i, j, hPos.At(i, j))
		}
	}
	return H
}

func (to *TrackedObject) flattenDetectionPoints(detection *Detection) *mat.Dense {
	flattened := mat.NewDense(to.DimZ, 1, nil)
	flatIdx := 0
	for i := 0; i < to.NumPoints; i++ {
		for d := 0; d < to.DimPoints; d++ {
			flattened.Set(flatIdx, 0, detection.AbsolutePoints.At(i, d))
			flatIdx++
		}
	}
	return flattened
}

func (to *TrackedObject) handleFirstDetections(pointsMask []bool, detectionFlatten *mat.Dense) {
	firstDetectionMask := make([]bool, to.DimZ)
	for i := 0; i < to.NumPoints; i++ {
		if pointsMask[i] && !to.DetectedAtLeastOncePoints[i] {
			for d := 0; d < to.DimPoints; d++ {
				firstDetectionMask[i*to.DimPoints+d] = true
			}
		}
	}

	stateVector := to.Filter.GetStateVector()

	for i := 0; i < to.DimZ; i++ {
		if firstDetectionMask[i] {
			stateVector.Set(i, 0, detectionFlatten.At(i, 0))
		}
	}

	for i := 0; i < to.NumPoints; i++ {
		if !to.DetectedAtLeastOncePoints[i] {
			for d := 0; d < to.DimPoints; d++ {
				stateVector.Set(to.DimZ+i*to.DimPoints+d, 0, 0.0)
			}
		}
	}

	to.Filter.SetStateVector(stateVector)
}

func (to *TrackedObject) updateDetectedMask(pointsMask []bool) {
	for i := 0; i < to.NumPoints; i++ {
		if pointsMask[i] {
			to.DetectedAtLeastOncePoints[i] = true
		}
	}
}

// Merge merges another tracked object into this one (ReID matching).
// This keeps the old object's ID but takes the new object's state.
func (to *TrackedObject) Merge(trackedObject *TrackedObject) {
	// Reset ReID counter (back to life!)
	to.ReidHitCounter = nil

	// Restore hit counter
	to.HitCounter = to.InitialPeriod * 2

	// Take new object's state
	to.PointHitCounter = make([]int, len(trackedObject.PointHitCounter))
	copy(to.PointHitCounter, trackedObject.PointHitCounter)

	to.LastDistance = trackedObject.LastDistance
	to.CurrentMinDistance = trackedObject.CurrentMinDistance
	to.LastDetection = trackedObject.LastDetection

	to.DetectedAtLeastOncePoints = make([]bool, len(trackedObject.DetectedAtLeastOncePoints))
	copy(to.DetectedAtLeastOncePoints, trackedObject.DetectedAtLeastOncePoints)

	// Take new filter state
	to.Filter = trackedObject.Filter

	// Merge past detections
	for _, pastDetection := range trackedObject.PastDetections {
		to.conditionallyAddToPastDetections(pastDetection)
	}

	// Update cached estimate
	to.updateEstimate()
}

// GetEstimate returns the position estimate from the Kalman filter.
//
// Parameters:
//   - absolute: If true, returns absolute coordinates; if false, returns relative
func (to *TrackedObject) GetEstimate(absolute bool) (*mat.Dense, error) {
	// Extract position from filter state (first dimZ elements)
	stateVector := to.Filter.GetStateVector()
	positions := mat.NewDense(to.NumPoints, to.DimPoints, nil)

	for i := 0; i < to.NumPoints; i++ {
		for d := 0; d < to.DimPoints; d++ {
			idx := i*to.DimPoints + d
			positions.Set(i, d, stateVector.At(idx, 0))
		}
	}

	if to.AbsToRel == nil {
		// No coordinate transformation set
		if !absolute {
			// Positions are in relative coordinates
			return positions, nil
		} else {
			return nil, fmt.Errorf("you must provide 'coord_transformations' to get absolute coordinates")
		}
	} else {
		// Coordinate transformation is set (positions are in absolute coordinates)
		if absolute {
			return positions, nil
		} else {
			// Convert to relative
			return to.AbsToRel(positions), nil
		}
	}
}

// EstimateVelocity returns the velocity estimate from the Kalman filter.
func (to *TrackedObject) EstimateVelocity() *mat.Dense {
	stateVector := to.Filter.GetStateVector()
	velocities := mat.NewDense(to.NumPoints, to.DimPoints, nil)

	for i := 0; i < to.NumPoints; i++ {
		for d := 0; d < to.DimPoints; d++ {
			idx := i*to.DimPoints + d
			velocities.Set(i, d, stateVector.At(to.DimZ+idx, 0))
		}
	}

	return velocities
}

// LivePoints returns a boolean mask of which points are currently live.
func (to *TrackedObject) LivePoints() []bool {
	livePoints := make([]bool, to.NumPoints)
	for i := 0; i < to.NumPoints; i++ {
		livePoints[i] = to.PointHitCounter[i] > 0
	}
	return livePoints
}

// GetLivePoints returns a boolean mask of which points are currently live.
// Alias for LivePoints() required by drawing.TrackedObjectLike interface.
func (to *TrackedObject) GetLivePoints() []bool {
	return to.LivePoints()
}

// GetID returns the object's permanent ID.
// Required by drawing.TrackedObjectLike interface.
func (to *TrackedObject) GetID() *int {
	return to.ID
}

// GetLabel returns the object's label.
// Required by drawing.TrackedObjectLike interface.
func (to *TrackedObject) GetLabel() *string {
	return to.Label
}

// HitCounterIsPositive returns whether the object is alive.
func (to *TrackedObject) HitCounterIsPositive() bool {
	return to.HitCounter >= 0
}

// ReidHitCounterIsPositive returns whether the object can still be ReID'd.
func (to *TrackedObject) ReidHitCounterIsPositive() bool {
	return to.ReidHitCounter == nil || *to.ReidHitCounter >= 0
}

// UpdateCoordinateTransformation updates the coordinate transformation function.
func (to *TrackedObject) UpdateCoordinateTransformation(coordTransform CoordinateTransformation) {
	if coordTransform != nil {
		to.AbsToRel = coordTransform.AbsToRel

		// Transform last detection if it exists
		if to.LastDetection != nil {
			to.LastDetection.UpdateCoordinateTransformation(coordTransform)
		}

		// Transform past detections
		for _, pastDet := range to.PastDetections {
			pastDet.UpdateCoordinateTransformation(coordTransform)
		}
	}
}

// acquireIDs gets permanent IDs from the factory.
// Called when object transitions from initializing to initialized.
func (to *TrackedObject) acquireIDs() {
	id, globalID := to.objFactory.GetIDs()
	to.ID = &id
	to.GlobalID = &globalID
}

// conditionallyAddToPastDetections manages the past detections storage.
// Maintains a uniform distribution of past detections across the object's lifetime.
func (to *TrackedObject) conditionallyAddToPastDetections(detection *Detection) {
	if to.config.PastDetectionsLength == 0 {
		return
	}

	if len(to.PastDetections) < to.config.PastDetectionsLength {
		// Still have room
		detection.Age = to.Age
		to.PastDetections = append(to.PastDetections, detection)
	} else if len(to.PastDetections) > 0 {
		// Check if we should replace the oldest detection
		// This maintains uniform distribution: ages 0, N, 2N, 3N, ...
		if to.Age >= to.PastDetections[0].Age*to.config.PastDetectionsLength {
			// Remove oldest
			to.PastDetections = to.PastDetections[1:]
			// Add new
			detection.Age = to.Age
			to.PastDetections = append(to.PastDetections, detection)
		}
	}
}

// updateEstimate updates the cached Estimate field from the current filter state.
// This is called after filter operations (predict, update) to keep Estimate in sync.
// The cached estimate is always in relative (camera frame) coordinates.
func (to *TrackedObject) updateEstimate() {
	// Extract position from filter state (first dimZ elements)
	// The filter state is in absolute coordinates if coordinate transformations are active
	stateVector := to.Filter.GetStateVector()
	estimate := mat.NewDense(to.NumPoints, to.DimPoints, nil)

	for i := 0; i < to.NumPoints; i++ {
		for d := 0; d < to.DimPoints; d++ {
			idx := i*to.DimPoints + d
			estimate.Set(i, d, stateVector.At(idx, 0))
		}
	}

	// If coordinate transformations are active, convert from absolute to relative
	if to.AbsToRel != nil {
		estimate = to.AbsToRel(estimate)
	}

	to.Estimate = estimate
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
