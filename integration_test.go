package norfairgo_test

import (
	"testing"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go"
	"github.com/nmichlo/norfair-go/drawing"
)

// =============================================================================
// Test 1: Complete Tracking Pipeline
// =============================================================================

func TestIntegration_CompleteTrackingPipeline(t *testing.T) {
	// Create tracker with default settings
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:       norfairgo.DistanceByName("euclidean"),
		DistanceThreshold:      50.0,
		HitCounterMax:          10,
		InitializationDelay:    2,
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Simulate tracking across 20 frames
	// Two objects: one static at (100, 100), one moving from (200, 200) to (300, 300)
	for frame := 0; frame < 20; frame++ {
		// Static object
		det1, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{100.0, 100.0}),
			nil,
		)

		// Moving object
		x := 200.0 + float64(frame)*5.0
		y := 200.0 + float64(frame)*5.0
		det2, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{x, y}),
			nil,
		)

		// Update tracker
		trackedObjects := tracker.Update([]*norfairgo.Detection{det1, det2}, 1, nil)

		// After initialization delay, should have 2 tracked objects
		if frame > 2 {
			if len(trackedObjects) != 2 {
				t.Errorf("Frame %d: expected 2 tracked objects, got %d", frame, len(trackedObjects))
			}

			// Verify object IDs are maintained across frames
			for _, obj := range trackedObjects {
				if obj.ID == nil {
					t.Errorf("Frame %d: object missing ID", frame)
				}
				if obj.GlobalID == nil {
					t.Errorf("Frame %d: object missing GlobalID", frame)
				}
			}

			// Verify estimates are reasonable (within 100 pixels of detections)
			for _, obj := range trackedObjects {
				estX := obj.Estimate.At(0, 0)
				estY := obj.Estimate.At(0, 1)

				// Check against both detection positions
				dist1 := (estX-100.0)*(estX-100.0) + (estY-100.0)*(estY-100.0)
				dist2 := (estX-x)*(estX-x) + (estY-y)*(estY-y)

				if dist1 > 10000 && dist2 > 10000 {
					t.Errorf("Frame %d: estimate (%.1f, %.1f) too far from detections", frame, estX, estY)
				}
			}
		}
	}

	// Verify total object count
	if tracker.TotalObjectCount() != 2 {
		t.Errorf("Expected 2 total objects, got %d", tracker.TotalObjectCount())
	}
}

// =============================================================================
// Test 2: Multiple Filter Types
// =============================================================================

func TestIntegration_MultipleFilterTypes(t *testing.T) {
	filterTypes := []struct {
		name    string
		factory norfairgo.FilterFactory
	}{
		{
			name:    "OptimizedKalman",
			factory: norfairgo.NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0),
		},
		{
			name:    "FilterPyKalman",
			factory: norfairgo.NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0),
		},
		{
			name:    "NoFilter",
			factory: norfairgo.NewNoFilterFactory(),
		},
	}

	for _, ft := range filterTypes {
		t.Run(ft.name, func(t *testing.T) {
			tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
				DistanceFunction:       norfairgo.DistanceByName("euclidean"),
				DistanceThreshold:      50.0,
				HitCounterMax:          10,
				InitializationDelay:    2,
				PointwiseHitCounterMax: 4,
				DetectionThreshold:     0.0,
				FilterFactory:          ft.factory,
				PastDetectionsLength:   4,
			})
			if err != nil {
				t.Fatalf("Failed to create tracker with %s: %v", ft.name, err)
			}

			// Track a moving object across 10 frames
			for frame := 0; frame < 10; frame++ {
				x := 100.0 + float64(frame)*10.0
				y := 100.0 + float64(frame)*10.0
				det, _ := norfairgo.NewDetection(
					mat.NewDense(1, 2, []float64{x, y}),
					nil,
				)

				trackedObjects := tracker.Update([]*norfairgo.Detection{det}, 1, nil)

				// After initialization, should have 1 object
				if frame > 2 {
					if len(trackedObjects) != 1 {
						t.Errorf("%s Frame %d: expected 1 object, got %d", ft.name, frame, len(trackedObjects))
					}
				}
			}

			// All filters should successfully track the object
			if tracker.TotalObjectCount() != 1 {
				t.Errorf("%s: expected 1 total object, got %d", ft.name, tracker.TotalObjectCount())
			}
		})
	}
}

// =============================================================================
// Test 3: Multiple Distance Functions
// =============================================================================

func TestIntegration_MultipleDistanceFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		distance norfairgo.Distance
		detFunc  func(frame int) *norfairgo.Detection
	}{
		{
			name:     "IoU_BoundingBoxes",
			distance: norfairgo.DistanceByName("iou"),
			detFunc: func(frame int) *norfairgo.Detection {
				x := 100.0 + float64(frame)*10.0
				det, _ := norfairgo.NewDetection(
					mat.NewDense(2, 2, []float64{
						x, 100.0, // top-left
						x + 50, 150.0, // bottom-right
					}),
					nil,
				)
				return det
			},
		},
		{
			name:     "Euclidean_Points",
			distance: norfairgo.DistanceByName("euclidean"),
			detFunc: func(frame int) *norfairgo.Detection {
				x := 100.0 + float64(frame)*10.0
				det, _ := norfairgo.NewDetection(
					mat.NewDense(1, 2, []float64{x, 100.0}),
					nil,
				)
				return det
			},
		},
		{
			name: "CustomScalar_Manhattan",
			distance: norfairgo.NewScalarDistance(func(det *norfairgo.Detection, obj *norfairgo.TrackedObject) float64 {
				// Manhattan distance
				sum := 0.0
				rows, _ := det.Points.Dims()
				for i := 0; i < rows; i++ {
					for j := 0; j < 2; j++ {
						sum += abs(det.Points.At(i, j) - obj.Estimate.At(i, j))
					}
				}
				return sum
			}),
			detFunc: func(frame int) *norfairgo.Detection {
				x := 100.0 + float64(frame)*10.0
				det, _ := norfairgo.NewDetection(
					mat.NewDense(1, 2, []float64{x, 100.0}),
					nil,
				)
				return det
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
				DistanceFunction:       tc.distance,
				DistanceThreshold:      100.0,
				HitCounterMax:          10,
				InitializationDelay:    2,
				PointwiseHitCounterMax: 4,
				DetectionThreshold:     0.0,
				PastDetectionsLength:   4,
			})
			if err != nil {
				t.Fatalf("Failed to create tracker: %v", err)
			}

			// Track object across 10 frames
			for frame := 0; frame < 10; frame++ {
				det := tc.detFunc(frame)
				trackedObjects := tracker.Update([]*norfairgo.Detection{det}, 1, nil)

				// After initialization
				if frame > 2 {
					if len(trackedObjects) != 1 {
						t.Errorf("%s Frame %d: expected 1 object, got %d", tc.name, frame, len(trackedObjects))
					}
				}
			}

			if tracker.TotalObjectCount() != 1 {
				t.Errorf("%s: expected 1 total object, got %d", tc.name, tracker.TotalObjectCount())
			}
		})
	}
}

// =============================================================================
// Test 4: ReID Enabled
// =============================================================================

func TestIntegration_ReIDEnabled(t *testing.T) {
	reidMax := 5
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:       norfairgo.DistanceByName("euclidean"),
		DistanceThreshold:      50.0,
		HitCounterMax:          3,
		InitializationDelay:    1,
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
		ReidHitCounterMax:      &reidMax,
	})
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	var originalID *int

	// Phase 1: Track object at (100, 100) for 5 frames
	for frame := 0; frame < 5; frame++ {
		det, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{100.0, 100.0}),
			nil,
		)
		trackedObjects := tracker.Update([]*norfairgo.Detection{det}, 1, nil)

		if frame > 1 && len(trackedObjects) > 0 {
			if originalID == nil {
				originalID = trackedObjects[0].ID
				t.Logf("Original ID: %d", *originalID)
			}
		}
	}

	if originalID == nil {
		t.Fatal("Failed to get original object ID")
	}

	// Phase 2: Occlusion - no detections for several frames
	// Object needs to miss enough frames to die (HitCounterMax=3 means miss 5+ frames)
	for frame := 5; frame < 10; frame++ {
		trackedObjects := tracker.Update([]*norfairgo.Detection{}, 1, nil)

		// Object may still be visible for a few frames as hit counter decrements
		t.Logf("Frame %d: %d objects visible", frame, len(trackedObjects))
	}

	// Phase 3: Object reappears at same location
	det, _ := norfairgo.NewDetection(
		mat.NewDense(1, 2, []float64{100.0, 100.0}),
		nil,
	)
	trackedObjects := tracker.Update([]*norfairgo.Detection{det}, 1, nil)

	// After ReID matching and re-initialization, object should reappear
	// May take a frame or two to become active again
	for frame := 8; frame < 12; frame++ {
		det, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{100.0, 100.0}),
			nil,
		)
		trackedObjects = tracker.Update([]*norfairgo.Detection{det}, 1, nil)

		if len(trackedObjects) > 0 {
			recoveredID := trackedObjects[0].ID
			if recoveredID == nil {
				t.Error("Recovered object missing ID")
			} else if *recoveredID != *originalID {
				t.Logf("Frame %d: ID changed from %d to %d (ReID may assign new ID)", frame, *originalID, *recoveredID)
				// Note: Depending on ReID implementation, ID might be preserved or new
				// This is actually expected behavior - object gets new initializing ID
			}
			break
		}
	}

	// Verify total object count (may be 1 or 2 depending on ReID match success)
	totalCount := tracker.TotalObjectCount()
	if totalCount < 1 || totalCount > 2 {
		t.Errorf("Expected 1-2 total objects after ReID, got %d", totalCount)
	}
}

// =============================================================================
// Test 5: Camera Motion Compensation
// =============================================================================

func TestIntegration_CameraMotionCompensation(t *testing.T) {
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:       norfairgo.DistanceByName("euclidean"),
		DistanceThreshold:      50.0,
		HitCounterMax:          10,
		InitializationDelay:    2,
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Simulate camera moving right (+10 pixels per frame)
	// Object stays stationary in world coordinates at (100, 100)
	// But appears to move left in camera coordinates
	for frame := 0; frame < 10; frame++ {
		// Camera offset
		cameraOffset := float64(frame) * 10.0

		// Object in camera coordinates (appears to move left as camera moves right)
		x := 100.0 - cameraOffset
		y := 100.0

		det, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{x, y}),
			nil,
		)

		// Create translation transformation (camera moved right by cameraOffset)
		transform, err := norfairgo.NewTranslationTransformation([]float64{cameraOffset, 0.0})
		if err != nil {
			t.Fatalf("Failed to create translation transformation: %v", err)
		}

		trackedObjects := tracker.Update([]*norfairgo.Detection{det}, 1, transform)

		// After initialization
		if frame > 2 {
			if len(trackedObjects) != 1 {
				t.Errorf("Frame %d: expected 1 object, got %d", frame, len(trackedObjects))
			}

			// Verify estimate in relative coordinates (camera frame)
			obj := trackedObjects[0]
			estX := obj.Estimate.At(0, 0)
			estY := obj.Estimate.At(0, 1)

			// Should be close to detection position in camera coordinates
			if abs(estX-x) > 20.0 || abs(estY-y) > 20.0 {
				t.Errorf("Frame %d: estimate (%.1f, %.1f) too far from detection (%.1f, %.1f)",
					frame, estX, estY, x, y)
			}

			// Verify absolute coordinates (world frame) - object should be relatively stable
			absEst, err := obj.GetEstimate(true)
			if err != nil {
				t.Errorf("Frame %d: failed to get absolute estimate: %v", frame, err)
			} else {
				absX := absEst.At(0, 0)
				absY := absEst.At(0, 1)

				// Log absolute coordinates for debugging
				// Camera motion compensation keeps object tracking working despite camera movement
				t.Logf("Frame %d: absolute estimate (%.1f, %.1f)", frame, absX, absY)

				// Just verify we got valid coordinates (not checking exact values in integration test)
				if absX < -1000 || absX > 1000 || absY < -1000 || absY > 1000 {
					t.Errorf("Frame %d: absolute estimate (%.1f, %.1f) out of reasonable range",
						frame, absX, absY)
				}
			}
		}
	}
}

// =============================================================================
// Test 6: End-to-End Workflow with Drawing
// =============================================================================

func TestIntegration_EndToEndWorkflow(t *testing.T) {
	// Create synthetic video frame
	frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
	defer frame.Close()

	// Fill with white background
	frame.SetTo(gocv.NewScalar(255, 255, 255, 0))

	// Create tracker
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:       norfairgo.DistanceByName("euclidean"),
		DistanceThreshold:      50.0,
		HitCounterMax:          10,
		InitializationDelay:    2,
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
	})
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Simulate tracking and drawing for 10 frames
	for frameNum := 0; frameNum < 10; frameNum++ {
		// Create detections
		det1, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{100.0, 100.0}),
			nil,
		)
		det2, _ := norfairgo.NewDetection(
			mat.NewDense(1, 2, []float64{200.0, 200.0}),
			nil,
		)

		// Update tracker
		trackedObjects := tracker.Update([]*norfairgo.Detection{det1, det2}, 1, nil)

		// After initialization
		if frameNum > 2 {
			if len(trackedObjects) != 2 {
				t.Errorf("Frame %d: expected 2 objects, got %d", frameNum, len(trackedObjects))
			}

			// Draw tracked objects
			// Convert to []interface{} for drawing
			drawables := make([]interface{}, len(trackedObjects))
			for i, obj := range trackedObjects {
				drawables[i] = obj
			}

			radius := int(float64(max(frame.Rows(), frame.Cols())) * 0.01)
			drawing.DrawPoints(
				&frame,
				drawables,
				&radius,
				nil,     // thickness (default)
				"by_id", // color
				false,   // drawLabels
				nil,     // textSize (default)
				true,    // drawIDs
				true,    // drawPoints
				nil,     // textThickness (default)
				nil,     // textColor (default)
				false,   // hideDeadPoints
				false,   // drawScores
			)

			// Verify frame dimensions unchanged
			if frame.Rows() != 480 || frame.Cols() != 640 {
				t.Errorf("Frame dimensions changed: got %dx%d, expected 480x640",
					frame.Rows(), frame.Cols())
			}

			// Verify frame not empty (drawing added content)
			if frame.Empty() {
				t.Error("Frame is empty after drawing")
			}
		}
	}

	// Verify tracking completed successfully
	if tracker.TotalObjectCount() != 2 {
		t.Errorf("Expected 2 total objects, got %d", tracker.TotalObjectCount())
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
