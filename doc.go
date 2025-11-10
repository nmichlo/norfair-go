/*
Package norfairgo provides real-time multi-object tracking.

- norfairgo is a golang port of python's norfair https://github.com/tryolabs/norfair
- This project is in **no** way associated with the original

Track objects across video frames using distance metrics and Kalman filtering.

# Basic Usage

	tracker := norfairgo.NewTracker(
		norfairgo.WithDistanceFunction(
			norfairgo.GetDistanceFunctionByName("iou", nil),
		),
		norfairgo.WithDistanceThreshold(0.7),
	)

	for frame := range videoFrames {
		detections := getDetections(frame)
		trackedObjects := tracker.Update(detections)

		for _, obj := range trackedObjects {
			fmt.Printf("ID: %d, Position: %v\n", obj.ID, obj.LastDetection.Points)
		}
	}

# Core Types

Detection represents detected objects with keypoints or bounding boxes.

TrackedObject maintains state across frames:
  - ID: Unique identifier
  - LastDetection: Most recent matching detection
  - Age: Frames since first detection
  - HitCounter: Consecutive frames with detections

Tracker manages object lifecycle, association, and filtering.

# Distance Functions

Built-in: iou, euclidean, frobenius, manhattan
Custom: Implement Distance interface

# Filtering

  - OptimizedKalmanFilter: Fast, simplified covariance (default)
  - FilterPyKalmanFilter: Full Kalman, filterpy-compatible
  - NoFilter: No prediction
*/
package norfairgo
