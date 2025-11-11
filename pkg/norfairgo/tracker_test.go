package norfairgo

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Basic Tracker Tests
// =============================================================================

// Python equivalent: norfair/tracker.py::Tracker.__init__()
//
//	from norfair import Tracker
//
//	tracker = Tracker(
//	    distance_function="euclidean",
//	    distance_threshold=100.0,
//	    hit_counter_max=15,
//	    initialization_delay=None,  # defaults to hit_counter_max/2
//	    pointwise_hit_counter_max=4,
//	    detection_threshold=0.0,
//	    filter_factory=None,  # defaults to OptimizedKalmanFilterFactory
//	    past_detections_length=4,
//	)
//
// Validation: tools/validate_tracker/main.py tests full tracker behavioral equivalence
func TestTracker_NewTracker(t *testing.T) {
	// Test basic tracker creation
	tracker, err := NewTracker(&TrackerConfig{
		DistanceFunction:       DistanceByName("euclidean"),
		DistanceThreshold:      100.0,
		HitCounterMax:          15,
		InitializationDelay:    -1, // use default: 15/2 = 7
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		// FilterFactory: nil (use default)
		PastDetectionsLength: 4,
		// ReidDistanceFunction: nil (disabled)
		// ReidDistanceThreshold: 0.0
		// ReidHitCounterMax: nil (disabled)
	})

	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	if tracker == nil {
		t.Fatal("Tracker is nil")
	}

	// Verify configuration
	if tracker.Config.DistanceThreshold != 100.0 {
		t.Errorf("Expected distance threshold 100.0, got %f", tracker.Config.DistanceThreshold)
	}

	if tracker.Config.HitCounterMax != 15 {
		t.Errorf("Expected hit counter max 15, got %d", tracker.Config.HitCounterMax)
	}

	if tracker.Config.InitializationDelay != 7 {
		t.Errorf("Expected initialization delay 7 (15/2), got %d", tracker.Config.InitializationDelay)
	}

	// Verify initial state
	if len(tracker.TrackedObjects) != 0 {
		t.Errorf("Expected 0 tracked objects initially, got %d", len(tracker.TrackedObjects))
	}

	if tracker.CurrentObjectCount() != 0 {
		t.Errorf("Expected current object count 0, got %d", tracker.CurrentObjectCount())
	}

	if tracker.TotalObjectCount() != 0 {
		t.Errorf("Expected total object count 0, got %d", tracker.TotalObjectCount())
	}
}

func TestTracker_InvalidInitializationDelay(t *testing.T) {
	// Test that negative initialization_delay is rejected (note: -1 is sentinel for "use default")
	_, err := NewTracker(&TrackerConfig{
		DistanceFunction:       DistanceByName("euclidean"),
		DistanceThreshold:      100.0,
		HitCounterMax:          15,
		InitializationDelay:    -2, // invalid negative value (not sentinel -1)
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})

	if err == nil {
		t.Error("Expected error for negative initialization_delay, got nil")
	}

	// Test that initialization_delay >= hit_counter_max is rejected
	_, err = NewTracker(&TrackerConfig{
		DistanceFunction:       DistanceByName("euclidean"),
		DistanceThreshold:      100.0,
		HitCounterMax:          15,
		InitializationDelay:    15, // equal to hit_counter_max (invalid)
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})

	if err == nil {
		t.Error("Expected error for initialization_delay >= hit_counter_max, got nil")
	}
}

// Python equivalent: norfair/tracker.py::Tracker.update()
//
//	from norfair import Detection, Tracker
//	import numpy as np
//
//	tracker = Tracker(
//	    distance_function="euclidean",
//	    distance_threshold=100.0,
//	    hit_counter_max=5,
//	    initialization_delay=2,
//	)
//	detection = Detection(points=np.array([[10.0, 20.0]]))
//	active_objects = tracker.update(detections=[detection])
//	# Object is initializing, not yet returned in active_objects
//	# After initialization_delay hits, object becomes permanent
//
// Validation: tools/validate_tracker/main.py::run_scenario_1_simple_static()
func TestTracker_SimpleUpdate(t *testing.T) {
	// Create tracker
	tracker, err := NewTracker(&TrackerConfig{
		DistanceFunction:       DistanceByName("euclidean"),
		DistanceThreshold:      100.0,
		HitCounterMax:          5,
		InitializationDelay:    -1, // use default: 5/2 = 2
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})

	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Create a detection
	points := mat.NewDense(1, 2, []float64{10.0, 20.0})
	detection, err := NewDetection(points, nil)
	if err != nil {
		t.Fatalf("Failed to create detection: %v", err)
	}

	// Update with detection
	activeObjects := tracker.Update([]*Detection{detection}, 1, nil)

	// Should have 0 active objects (still initializing)
	if len(activeObjects) != 0 {
		t.Errorf("Expected 0 active objects (initializing), got %d", len(activeObjects))
	}

	// Should have 1 tracked object total
	if len(tracker.TrackedObjects) != 1 {
		t.Errorf("Expected 1 tracked object, got %d", len(tracker.TrackedObjects))
	}

	// Total count should be 0 (object hasn't gotten permanent ID yet)
	if tracker.TotalObjectCount() != 0 {
		t.Errorf("Expected total count 0 (still initializing), got %d", tracker.TotalObjectCount())
	}

	// Object should be initializing
	if !tracker.TrackedObjects[0].IsInitializing {
		t.Error("Expected object to be initializing")
	}

	// Object should have initializing ID but not permanent ID
	if tracker.TrackedObjects[0].InitializingID == nil {
		t.Error("Expected initializing ID to be set")
	}
	if tracker.TrackedObjects[0].ID != nil {
		t.Error("Expected permanent ID to be nil (still initializing)")
	}
}

// Python equivalent: norfair/tracker.py::Tracker.update() with empty detections
//
//	from norfair import Tracker
//
//	tracker = Tracker(distance_function="euclidean", distance_threshold=100.0)
//	# Update with no detections - objects decay and eventually die
//	active_objects = tracker.update(detections=[])
//	# Tracked objects' hit_counter decrements, removed when reaches 0
//
// Validation: tools/validate_tracker/main.py::run_scenario_4_object_death()
func TestTracker_UpdateEmptyDetections(t *testing.T) {
	// Create tracker
	tracker, err := NewTracker(&TrackerConfig{
		DistanceFunction:       DistanceByName("euclidean"),
		DistanceThreshold:      100.0,
		HitCounterMax:          5,
		InitializationDelay:    -1, // use default
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})

	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Update with no detections
	activeObjects := tracker.Update([]*Detection{}, 1, nil)

	if len(activeObjects) != 0 {
		t.Errorf("Expected 0 active objects, got %d", len(activeObjects))
	}

	// Update with nil detections
	activeObjects = tracker.Update(nil, 1, nil)

	if len(activeObjects) != 0 {
		t.Errorf("Expected 0 active objects, got %d", len(activeObjects))
	}
}

// Python equivalent: norfair/tracker.py::Detection
//
//	from norfair import Detection
//	import numpy as np
//
//	# Simple detection with points only
//	detection = Detection(points=np.array([[x, y]]))
//
//	# Detection with points and scores
//	detection = Detection(
//	    points=np.array([[x1, y1], [x2, y2]]),
//	    scores=np.array([0.9, 0.8])
//	)
func TestDetection_Creation(t *testing.T) {
	// Test valid 2D points
	points2D := mat.NewDense(3, 2, []float64{
		1.0, 2.0,
		3.0, 4.0,
		5.0, 6.0,
	})

	det2D, err := NewDetection(points2D, nil)
	if err != nil {
		t.Fatalf("Failed to create 2D detection: %v", err)
	}

	if det2D == nil {
		t.Fatal("Detection is nil")
	}

	// Verify points are copied
	rows, cols := det2D.Points.Dims()
	if rows != 3 || cols != 2 {
		t.Errorf("Expected points shape (3,2), got (%d,%d)", rows, cols)
	}

	// Test valid 3D points
	points3D := mat.NewDense(2, 3, []float64{
		1.0, 2.0, 3.0,
		4.0, 5.0, 6.0,
	})

	det3D, err := NewDetection(points3D, nil)
	if err != nil {
		t.Fatalf("Failed to create 3D detection: %v", err)
	}

	rows, cols = det3D.Points.Dims()
	if rows != 2 || cols != 3 {
		t.Errorf("Expected points shape (2,3), got (%d,%d)", rows, cols)
	}
}

// Python equivalent: norfair/tracker.py::TrackedObject (internal class)
//
//	from norfair.tracker import TrackedObject
//
//	# TrackedObject is created internally by Tracker
//	# It maintains:
//	# - id: permanent ID (once initialized)
//	# - global_id: unique across all trackers
//	# - initializing_id: temporary ID during initialization
//	# - is_initializing: whether object is still initializing
//	# - hit_counter: remaining frames before death
//	# - age: frames since first detection
//	# - estimate: current state estimate from Kalman filter
//	# - last_detection: most recent matched detection
//	# - past_detections: history of detections
func TestTrackedObject_Creation(t *testing.T) {
	// Create detection
	points := mat.NewDense(2, 2, []float64{
		10.0, 20.0,
		30.0, 40.0,
	})
	detection, err := NewDetection(points, nil)
	if err != nil {
		t.Fatalf("Failed to create detection: %v", err)
	}

	// Create factory
	factory := NewTrackedObjectFactory()

	// Create config
	filterFactory := NewOptimizedKalmanFilterFactory(1.0, 1.0, 10.0, 0.0, 1.0)
	config := &TrackerConfig{
		HitCounterMax:          15,
		InitializationDelay:    7,
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		FilterFactory:          filterFactory,
		PastDetectionsLength:   4,
		ReidHitCounterMax:      nil,
	}

	// Create tracked object
	obj, err := NewTrackedObject(
		factory,
		detection,
		config,
		1,   // period
		nil, // coord_transformations
	)

	if err != nil {
		t.Fatalf("Failed to create tracked object: %v", err)
	}

	// Verify initialization
	if obj.NumPoints != 2 {
		t.Errorf("Expected 2 points, got %d", obj.NumPoints)
	}

	if obj.DimPoints != 2 {
		t.Errorf("Expected 2D points, got %d", obj.DimPoints)
	}

	if obj.HitCounter != 1 {
		t.Errorf("Expected hit counter 1, got %d", obj.HitCounter)
	}

	if !obj.IsInitializing {
		t.Error("Expected object to be initializing")
	}

	if obj.InitializingID == nil {
		t.Error("Expected initializing ID to be set")
	}

	if obj.ID != nil {
		t.Error("Expected permanent ID to be nil (still initializing)")
	}
}

//
// Mock CoordinateTransformation for testing camera motion
//

type mockCoordinateTransformation struct {
	relativePoints *mat.Dense
	absolutePoints *mat.Dense
}

func newMockCoordinateTransformation(relativePoints, absolutePoints *mat.Dense) *mockCoordinateTransformation {
	return &mockCoordinateTransformation{
		relativePoints: relativePoints,
		absolutePoints: absolutePoints,
	}
}

func (m *mockCoordinateTransformation) RelToAbs(points *mat.Dense) *mat.Dense {
	// Verify input matches expected relative points
	if !matApproxEqual(points, m.relativePoints, 0.001) {
		panic("RelToAbs: input points don't match expected relative points")
	}
	return m.absolutePoints
}

func (m *mockCoordinateTransformation) AbsToRel(points *mat.Dense) *mat.Dense {
	// Debug: log what we're receiving
	// Verify input matches expected absolute points
	if !matApproxEqual(points, m.absolutePoints, 0.001) {
		// panic with details
		panic(fmt.Sprintf("AbsToRel: input points don't match expected absolute points\nExpected: %v\nGot: %v",
			mat.Formatted(m.absolutePoints), mat.Formatted(points)))
	}
	return m.relativePoints
}

// Python equivalent: norfair/tracker.py::Tracker.update() with camera_motion
//
//	from norfair import Tracker
//	from norfair.camera_motion import MotionEstimator
//
//	tracker = Tracker(distance_function="euclidean", distance_threshold=100.0)
//	motion_estimator = MotionEstimator()
//
//	# Update tracker with camera motion compensation
//	coord_transform = motion_estimator.update(frame, tracked_objects)
//	active_objects = tracker.update(
//	    detections=detections,
//	    coord_transformations=coord_transform
//	)
//	# Tracked object positions are compensated for camera movement
func TestTracker_CameraMotion(t *testing.T) {
	// Test camera motion with both 1D [1,1] and 2D [[1,1]] point formats
	testCases := []struct {
		name           string
		absolutePoints []float64
		absoluteShape  [2]int // rows, cols
	}{
		{
			name:           "1D points",
			absolutePoints: []float64{1, 1},
			absoluteShape:  [2]int{1, 2}, // 1 row, 2 cols
		},
		{
			name:           "2D points",
			absolutePoints: []float64{1, 1},
			absoluteShape:  [2]int{1, 2}, // 1 row, 2 cols (same as 1D after ValidatePoints)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create tracker with euclidean distance, threshold=1, initialization_delay=0
			tracker, err := NewTracker(&TrackerConfig{
				DistanceFunction:       DistanceByName("euclidean"),
				DistanceThreshold:      1.0,
				HitCounterMax:          1,
				InitializationDelay:    0, // no initialization delay
				PointwiseHitCounterMax: 4,
				DetectionThreshold:     0.0,
				PastDetectionsLength:   4,
			})
			if err != nil {
				t.Fatalf("Failed to create tracker: %v", err)
			}

			// Setup: absolute_points = [1, 1], relative_points = [2, 2] (camera moved by +1, +1)
			absolutePts := mat.NewDense(tc.absoluteShape[0], tc.absoluteShape[1], tc.absolutePoints)
			relativePts := mat.NewDense(tc.absoluteShape[0], tc.absoluteShape[1], []float64{2, 2})

			// Create mock coordinate transformation
			coordTransform := newMockCoordinateTransformation(relativePts, absolutePts)

			// Create detection with relative points
			detection, err := NewDetection(relativePts, nil)
			if err != nil {
				t.Fatalf("Failed to create detection: %v", err)
			}

			// Update tracker with coordinate transformation (period=1)
			trackedObjects := tracker.Update([]*Detection{detection}, 1, coordTransform)

			// Assert that the detection was correctly updated
			// detection.AbsolutePoints should be transformed to absolute coordinates
			if !matApproxEqual(detection.AbsolutePoints, absolutePts, 0.001) {
				t.Errorf("Detection.AbsolutePoints mismatch:\nExpected:\n%v\nGot:\n%v",
					mat.Formatted(absolutePts),
					mat.Formatted(detection.AbsolutePoints))
			}

			// detection.Points should remain in relative coordinates (unchanged)
			if !matApproxEqual(detection.Points, relativePts, 0.001) {
				t.Errorf("Detection.Points should remain unchanged:\nExpected:\n%v\nGot:\n%v",
					mat.Formatted(relativePts),
					mat.Formatted(detection.Points))
			}

			// Check the tracked object
			if len(trackedObjects) != 1 {
				t.Fatalf("Expected 1 tracked object, got %d", len(trackedObjects))
			}

			obj := trackedObjects[0]

			// GetEstimate(false) should return relative coordinates
			estimateRel, err := obj.GetEstimate(false)
			if err != nil {
				t.Fatalf("Failed to get relative estimate: %v", err)
			}
			if !matApproxEqual(estimateRel, relativePts, 0.1) {
				t.Errorf("GetEstimate(false) should return relative coordinates:\nExpected:\n%v\nGot:\n%v",
					mat.Formatted(relativePts),
					mat.Formatted(estimateRel))
			}

			// GetEstimate(true) should return absolute coordinates
			estimateAbs, err := obj.GetEstimate(true)
			if err != nil {
				t.Fatalf("Failed to get absolute estimate: %v", err)
			}
			if !matApproxEqual(estimateAbs, absolutePts, 0.1) {
				t.Errorf("GetEstimate(true) should return absolute coordinates:\nExpected:\n%v\nGot:\n%v",
					mat.Formatted(absolutePts),
					mat.Formatted(estimateAbs))
			}

			// obj.Estimate (default) should be relative coordinates
			if !matApproxEqual(obj.Estimate, relativePts, 0.1) {
				t.Errorf("obj.Estimate (default) should be relative coordinates:\nExpected:\n%v\nGot:\n%v",
					mat.Formatted(relativePts),
					mat.Formatted(obj.Estimate))
			}
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
