package norfairgo

import (
	"fmt"
	"math"
)

// TrackerConfig contains all configuration parameters for a Tracker.
// This separates configuration (immutable after creation) from state (mutable during tracking).
type TrackerConfig struct {
	// Distance function for matching detections to tracked objects.
	// Use DistanceByName("euclidean") or pass a custom Distance implementation.
	DistanceFunction Distance

	// Maximum distance for a valid detection-object match.
	// Pairs with distance > threshold will not be matched.
	DistanceThreshold float64

	// Maximum "hits" an object can accumulate before being clamped.
	// Objects lose 1 hit per frame without detection, gain 2*period hits per match.
	// Default: 15
	HitCounterMax int

	// Number of hits required before an object becomes "active".
	// Objects in initialization phase are not returned to users.
	// Use -1 for default (hitCounterMax / 2).
	// Default: hitCounterMax / 2
	InitializationDelay int

	// Maximum hits for individual points (for pose estimation).
	// Each point has its own hit counter for tracking reliability.
	// Default: 4
	PointwiseHitCounterMax int

	// Minimum confidence score for a point to be considered detected.
	// Points with scores below this threshold are excluded from Kalman updates.
	// Default: 0.0
	DetectionThreshold float64

	// Factory for creating Kalman filters for tracked objects.
	// Default: OptimizedKalmanFilterFactory with default parameters
	FilterFactory FilterFactory

	// Number of past detections to store per object.
	// Used for metric learning and appearance-based distance functions.
	// Default: 4
	PastDetectionsLength int

	// Re-identification (ReID) distance function for recovering lost identities.
	// Set to nil to disable ReID.
	// Default: nil (disabled)
	ReidDistanceFunction Distance

	// Maximum distance for ReID matching.
	// Only applies when ReID is enabled.
	// Default: 0.0
	ReidDistanceThreshold float64

	// How long a "dead" object survives for ReID matching.
	// Objects with hit_counter <= 0 enter ReID mode with this counter value.
	// Set to nil or 0 to disable ReID.
	// Default: nil (disabled)
	ReidHitCounterMax *int
}

// Tracker is the main object tracking class that manages the lifecycle of tracked objects.
type Tracker struct {
	// Configuration (immutable after creation)
	Config *TrackerConfig

	// State (mutable during tracking)
	TrackedObjects []*TrackedObject
	objFactory     *TrackedObjectFactory
}

// NewTracker creates a new Tracker from a configuration.
//
// Example:
//
//	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
//	    DistanceFunction:  norfairgo.DistanceByName("iou"),
//	    DistanceThreshold: 0.5,
//	    HitCounterMax:     30,
//	    // Omitted fields will use sensible defaults
//	})
//
// Zero/nil values in config will be replaced with defaults:
//   - DistanceFunction: euclidean (if nil)
//   - DistanceThreshold: 1.0 (if 0)
//   - HitCounterMax: 15 (if 0)
//   - InitializationDelay: hitCounterMax/2 (if -1)
//   - PointwiseHitCounterMax: 4 (if 0)
//   - DetectionThreshold: 0.0
//   - FilterFactory: OptimizedKalmanFilterFactory (if nil)
//   - PastDetectionsLength: 4 (if 0)
//   - ReidDistanceFunction: nil (disabled)
//   - ReidDistanceThreshold: 0.0
//   - ReidHitCounterMax: nil (disabled)
func NewTracker(config *TrackerConfig) (*Tracker, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Apply defaults for zero/nil values
	if config.DistanceFunction == nil {
		euclidean := GetDistanceByName("euclidean")
		if euclidean == nil {
			return nil, fmt.Errorf("failed to get default euclidean distance")
		}
		config.DistanceFunction = euclidean
	}

	if config.DistanceThreshold == 0 {
		config.DistanceThreshold = 1.0
	}

	if config.HitCounterMax == 0 {
		config.HitCounterMax = 15
	}

	if config.InitializationDelay == -1 {
		config.InitializationDelay = config.HitCounterMax / 2
	}

	if config.PointwiseHitCounterMax == 0 {
		config.PointwiseHitCounterMax = 4
	}

	if config.FilterFactory == nil {
		config.FilterFactory = NewOptimizedKalmanFilterFactory(
			1.0,  // R_mult
			1.0,  // Q_mult
			10.0, // pos_var
			0.0,  // pos_vel_cov
			1.0,  // vel_var
		)
	}

	if config.PastDetectionsLength == 0 {
		config.PastDetectionsLength = 4
	}

	// Validate configuration
	if config.PastDetectionsLength < 0 {
		return nil, fmt.Errorf("past_detections_length must be >= 0, got %d", config.PastDetectionsLength)
	}

	if config.InitializationDelay < 0 || config.InitializationDelay >= config.HitCounterMax {
		return nil, fmt.Errorf(
			"initialization_delay must be >= 0 and < hit_counter_max (%d), got %d",
			config.HitCounterMax,
			config.InitializationDelay,
		)
	}

	// Create tracker with config and initial state
	return &Tracker{
		Config:         config,
		TrackedObjects: []*TrackedObject{},
		objFactory:     NewTrackedObjectFactory(),
	}, nil
}

// Update processes detections for the current frame and returns active tracked objects.
//
// This implements the 8-stage tracking pipeline:
// 1. Coordinate transformation
// 2. Object cleanup
// 3. State prediction
// 4. Match initialized objects
// 5. Match initializing objects
// 6. ReID matching
// 7. Create new objects
// 8. Return active objects
//
// Parameters:
//   - detections: List of detections for this frame (nil = no detections)
//   - period: Time period since last update (default: 1)
//   - coordTransformations: Coordinate transformation for camera motion (nil = no transformation)
func (t *Tracker) Update(
	detections []*Detection,
	period int,
	coordTransformations CoordinateTransformation,
) []*TrackedObject {
	// Handle nil detections
	if detections == nil {
		detections = []*Detection{}
	}

	// =========================================================================
	// STAGE 1: Coordinate Transformation
	// =========================================================================
	if coordTransformations != nil {
		for _, det := range detections {
			det.UpdateCoordinateTransformation(coordTransformations)
		}
	}

	// =========================================================================
	// STAGE 2: Object Cleanup
	// =========================================================================
	var aliveObjects []*TrackedObject
	var deadObjects []*TrackedObject

	if t.Config.ReidHitCounterMax == nil {
		// No ReID: Remove objects with hit_counter < 0
		newTrackedObjects := []*TrackedObject{}
		for _, obj := range t.TrackedObjects {
			if obj.HitCounterIsPositive() {
				newTrackedObjects = append(newTrackedObjects, obj)
			}
		}
		t.TrackedObjects = newTrackedObjects
		aliveObjects = t.TrackedObjects
	} else {
		// With ReID: Keep objects with reid_hit_counter >= 0
		newTrackedObjects := []*TrackedObject{}
		for _, obj := range t.TrackedObjects {
			if obj.ReidHitCounterIsPositive() {
				newTrackedObjects = append(newTrackedObjects, obj)
				if obj.HitCounterIsPositive() {
					aliveObjects = append(aliveObjects, obj)
				} else {
					deadObjects = append(deadObjects, obj)
				}
			}
		}
		t.TrackedObjects = newTrackedObjects
	}

	// =========================================================================
	// STAGE 3: State Prediction
	// =========================================================================
	for _, obj := range t.TrackedObjects {
		obj.TrackerStep() // Decrements counters, increments age, calls filter.predict()
		obj.UpdateCoordinateTransformation(coordTransformations)
	}

	// =========================================================================
	// STAGE 4: Match Initialized Objects
	// =========================================================================
	// Filter to non-initializing objects
	initializedObjects := []*TrackedObject{}
	for _, obj := range aliveObjects {
		if !obj.IsInitializing {
			initializedObjects = append(initializedObjects, obj)
		}
	}

	unmatchedDetections, _, unmatchedInitTrackers := t.updateObjectsInPlace(
		t.Config.DistanceFunction,
		t.Config.DistanceThreshold,
		initializedObjects,
		detections,
		period,
	)

	// =========================================================================
	// STAGE 5: Match Initializing Objects
	// =========================================================================
	// Filter to initializing objects
	initializingObjects := []*TrackedObject{}
	for _, obj := range aliveObjects {
		if obj.IsInitializing {
			initializingObjects = append(initializingObjects, obj)
		}
	}

	unmatchedDetections, matchedNotInitTrackers, _ := t.updateObjectsInPlace(
		t.Config.DistanceFunction,
		t.Config.DistanceThreshold,
		initializingObjects,
		unmatchedDetections,
		period,
	)

	// =========================================================================
	// STAGE 6: ReID Matching
	// =========================================================================
	if t.Config.ReidDistanceFunction != nil {
		// Combine unmatched initialized objects with dead objects
		reidCandidates := append(unmatchedInitTrackers, deadObjects...)

		t.updateObjectsInPlace(
			t.Config.ReidDistanceFunction,
			t.Config.ReidDistanceThreshold,
			reidCandidates,
			matchedNotInitTrackers,
			period,
		)
	}

	// =========================================================================
	// STAGE 7: Create New Objects
	// =========================================================================
	// Type-assert unmatchedDetections to []*Detection
	unmatchedDets, ok := unmatchedDetections.([]*Detection)
	if !ok {
		// If not detections, it might be tracked objects (shouldn't happen in stage 7)
		unmatchedDets = []*Detection{}
	}

	for _, detection := range unmatchedDets {
		newObj, err := NewTrackedObject(
			t.objFactory,
			detection,
			t.Config,
			period,
			coordTransformations,
		)
		if err != nil {
			// Skip invalid detections
			fmt.Printf("Warning: failed to create tracked object: %v\n", err)
			continue
		}
		t.TrackedObjects = append(t.TrackedObjects, newObj)
	}

	// =========================================================================
	// STAGE 8: Return Active Objects
	// =========================================================================
	return t.GetActiveObjects()
}

// updateObjectsInPlace matches candidates to objects and updates them in place.
//
// Parameters:
//   - distanceFunction: Distance metric to use
//   - distanceThreshold: Maximum distance for valid match
//   - objects: Objects to match against
//   - candidates: Candidates to match (Detections or TrackedObjects)
//   - period: Time period for hit counter updates
//
// Returns:
//   - unmatchedCandidates: Candidates that were not matched
//   - matchedObjects: Objects that were matched
//   - unmatchedObjects: Objects that were not matched
func (t *Tracker) updateObjectsInPlace(
	distanceFunction Distance,
	distanceThreshold float64,
	objects []*TrackedObject,
	candidates interface{},
	period int,
) (unmatchedCandidates interface{}, matchedObjects []*TrackedObject, unmatchedObjects []*TrackedObject) {
	// Handle empty candidates
	if candidates == nil {
		return []interface{}{}, []*TrackedObject{}, objects
	}

	// Convert candidates to slice
	var candList interface{}
	var numCandidates int

	switch c := candidates.(type) {
	case []*Detection:
		candList = c
		numCandidates = len(c)
	case []*TrackedObject:
		candList = c
		numCandidates = len(c)
	default:
		panic(fmt.Sprintf("unsupported candidates type: %T", candidates))
	}

	if numCandidates == 0 {
		return candidates, []*TrackedObject{}, objects
	}

	// Handle empty objects
	if len(objects) == 0 {
		return candidates, []*TrackedObject{}, objects
	}

	// Compute distance matrix
	distanceMatrix := distanceFunction.GetDistances(objects, candList)

	// Validate for NaN
	err := ValidateDistanceMatrix(distanceMatrix)
	if err != nil {
		panic(fmt.Sprintf("distance function error: %v", err))
	}

	// Store minimum distances for debugging
	rows, cols := distanceMatrix.Dims()
	for i := 0; i < cols; i++ {
		if i >= len(objects) {
			break
		}
		// Find minimum in column i
		minVal := math.Inf(1)
		for j := 0; j < rows; j++ {
			val := distanceMatrix.At(j, i)
			if val < minVal {
				minVal = val
			}
		}
		if minVal < distanceThreshold {
			objects[i].CurrentMinDistance = &minVal
		} else {
			objects[i].CurrentMinDistance = nil
		}
	}

	// Greedy matching
	matchedCandIndices, matchedObjIndices := MatchDetectionsAndObjects(distanceMatrix, distanceThreshold)

	// Process matches
	if len(matchedCandIndices) > 0 {
		// Build sets of matched indices
		matchedCandSet := make(map[int]bool)
		matchedObjSet := make(map[int]bool)
		for i := range matchedCandIndices {
			matchedCandSet[matchedCandIndices[i]] = true
			matchedObjSet[matchedObjIndices[i]] = true
		}

		// Separate unmatched candidates and objects
		var unmatchedCandList []interface{}
		var unmatchedObjList []*TrackedObject

		switch cands := candList.(type) {
		case []*Detection:
			for i, cand := range cands {
				if !matchedCandSet[i] {
					unmatchedCandList = append(unmatchedCandList, cand)
				}
			}
			unmatchedCandidates = convertToDetectionSlice(unmatchedCandList)
		case []*TrackedObject:
			for i, cand := range cands {
				if !matchedCandSet[i] {
					unmatchedCandList = append(unmatchedCandList, cand)
				}
			}
			unmatchedCandidates = convertToTrackedObjectSlice(unmatchedCandList)
		}

		for i, obj := range objects {
			if !matchedObjSet[i] {
				unmatchedObjList = append(unmatchedObjList, obj)
			}
		}

		// Process each match
		matchedObjList := []*TrackedObject{}
		for i := range matchedCandIndices {
			candIdx := matchedCandIndices[i]
			objIdx := matchedObjIndices[i]
			distance := distanceMatrix.At(candIdx, objIdx)

			if distance < distanceThreshold {
				matchedObject := objects[objIdx]

				// Check candidate type
				switch cands := candList.(type) {
				case []*Detection:
					// Candidate is Detection - update object
					matchedCandidate := cands[candIdx]
					matchedObject.Hit(matchedCandidate, period)
					matchedObject.LastDistance = &distance
					matchedObjList = append(matchedObjList, matchedObject)

				case []*TrackedObject:
					// Candidate is TrackedObject - merge (ReID case)
					matchedCandidate := cands[candIdx]
					matchedObject.Merge(matchedCandidate)

					// Remove matched candidate from tracker's object list
					t.removeTrackedObject(matchedCandidate)
				}
			} else {
				// Distance >= threshold - add to unmatched
				switch cands := candList.(type) {
				case []*Detection:
					unmatchedCandidates = append(unmatchedCandidates.([]*Detection), cands[candIdx])
				case []*TrackedObject:
					unmatchedCandidates = append(unmatchedCandidates.([]*TrackedObject), cands[candIdx])
				}
				unmatchedObjList = append(unmatchedObjList, objects[objIdx])
			}
		}

		return unmatchedCandidates, matchedObjList, unmatchedObjList
	}

	// No matches
	return candidates, []*TrackedObject{}, objects
}

// CurrentObjectCount returns the number of currently active objects.
func (t *Tracker) CurrentObjectCount() int {
	return len(t.GetActiveObjects())
}

// TotalObjectCount returns the total number of objects ever created.
func (t *Tracker) TotalObjectCount() int {
	return t.objFactory.Count()
}

// GetActiveObjects returns objects that are not initializing and have positive hit counter.
func (t *Tracker) GetActiveObjects() []*TrackedObject {
	activeObjects := []*TrackedObject{}
	for _, obj := range t.TrackedObjects {
		if !obj.IsInitializing && obj.HitCounterIsPositive() {
			activeObjects = append(activeObjects, obj)
		}
	}
	return activeObjects
}

// removeTrackedObject removes a tracked object from the tracker's list.
// This is used during ReID merging.
func (t *Tracker) removeTrackedObject(objToRemove *TrackedObject) {
	newList := []*TrackedObject{}
	for _, obj := range t.TrackedObjects {
		if obj != objToRemove {
			newList = append(newList, obj)
		}
	}
	t.TrackedObjects = newList
}

// Helper functions for type conversion
func convertToDetectionSlice(list []interface{}) []*Detection {
	result := make([]*Detection, len(list))
	for i, item := range list {
		result[i] = item.(*Detection)
	}
	return result
}

func convertToTrackedObjectSlice(list []interface{}) []*TrackedObject {
	result := make([]*TrackedObject, len(list))
	for i, item := range list {
		result[i] = item.(*TrackedObject)
	}
	return result
}
