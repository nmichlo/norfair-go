package main

import (
	"fmt"
	"log"

	norfairgo "github.com/nmichlo/norfair-go"
	"gonum.org/v1/gonum/mat"
)

func main() {
	fmt.Println("Simple Tracking Example")
	fmt.Println("========================")

	// Create tracker with IoU distance and default parameters
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:       norfairgo.DistanceByName("iou"),
		DistanceThreshold:      0.5,
		HitCounterMax:          30,
		InitializationDelay:    3, // Objects need 3 detections to be initialized
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		// FilterFactory defaults to OptimizedKalmanFilter
		PastDetectionsLength: 4,
		// ReID disabled by default
	})
	if err != nil {
		log.Fatalf("Failed to create tracker: %v", err)
	}

	fmt.Println("\nSimulating 20 frames of tracking...")
	fmt.Println("Two objects moving across the frame\n")

	n := 20
	// Simulate detections from an object detector over multiple frames
	// Format: bounding boxes as [[x_min, y_min], [x_max, y_max]]
	for frame := 0; frame < n; frame++ {
		// Object 1: Moving right and down
		x1 := float64(100 + frame*10)
		y1 := float64(100 + frame*5)
		det1, err := norfairgo.NewDetection(
			mat.NewDense(2, 2, []float64{
				x1, y1, // top-left corner
				x1 + 100, y1 + 100, // bottom-right corner
			}),
			nil, // optional config (scores, data, label, embedding)
		)
		if err != nil {
			log.Fatalf("Failed to create detection 1: %v", err)
		}

		// Object 2: Moving left and down
		x2 := float64(500 - frame*8)
		y2 := float64(150 + frame*6)
		det2, err := norfairgo.NewDetection(
			mat.NewDense(2, 2, []float64{
				x2, y2,
				x2 + 100, y2 + 100,
			}),
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to create detection 2: %v", err)
		}

		// only add frames based on certain conditions to simulate loss
		var detections []*norfairgo.Detection
		if frame < n-12 {
			detections = append(detections, det1)
		}
		if frame >= 5 {
			detections = append(detections, det2)
		}

		// Update tracker with current frame detections
		trackedObjects := tracker.Update(detections, 1, nil)

		// Display tracking results
		fmt.Printf("Frame %2d: ", frame)
		if len(trackedObjects) > 0 {
			fmt.Printf("%d tracked objects", len(trackedObjects))
			for _, obj := range trackedObjects {
				// Get the center point of the bounding box
				centerX := obj.Estimate.At(0, 0)
				centerY := obj.Estimate.At(1, 0)
				fmt.Printf(" | ID %d: (%.1f, %.1f) age=%d hits=%d",
					obj.ID,
					centerX,
					centerY,
					obj.Age,
					obj.HitCounter,
				)
			}
			fmt.Println()
		} else {
			fmt.Println("(initializing...)")
		}
	}

	fmt.Println("\n✓ Tracking completed successfully!")
	fmt.Println("\nNote: Objects are initialized after 3 consecutive detections.")
	fmt.Println("      This prevents tracking of spurious detections.")
}
