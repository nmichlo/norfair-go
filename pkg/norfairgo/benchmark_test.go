package norfairgo

import (
	"testing"

	"gonum.org/v1/gonum/mat"
)

// Benchmark helpers
func createTestDetections(n int) []*Detection {
	detections := make([]*Detection, n)
	for i := 0; i < n; i++ {
		x := float64(i * 100)
		y := float64(i * 50)
		points := mat.NewDense(2, 2, []float64{
			x, y,
			x + 50, y + 50,
		})
		detections[i], _ = NewDetection(points, nil)
	}
	return detections
}

// ============================================================================
// Tracker Benchmarks
// ============================================================================

func BenchmarkTrackerUpdate_10Objects(b *testing.B) {
	tracker, _ := NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName("euclidean"),
		DistanceThreshold:   50.0,
		HitCounterMax:       30,
		InitializationDelay: 3,
		FilterFactory:       NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0),
	})
	detections := createTestDetections(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(detections, 1, nil)
	}
}

func BenchmarkTrackerUpdate_50Objects(b *testing.B) {
	tracker, _ := NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName("euclidean"),
		DistanceThreshold:   50.0,
		HitCounterMax:       30,
		InitializationDelay: 3,
		FilterFactory:       NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0),
	})
	detections := createTestDetections(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(detections, 1, nil)
	}
}

func BenchmarkTrackerUpdate_100Objects(b *testing.B) {
	tracker, _ := NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName("euclidean"),
		DistanceThreshold:   50.0,
		HitCounterMax:       30,
		InitializationDelay: 3,
		FilterFactory:       NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0),
	})
	detections := createTestDetections(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(detections, 1, nil)
	}
}

// ============================================================================
// Tracker with Different Filters
// ============================================================================

func BenchmarkTrackerUpdate_100Objects_FilterPyKalman(b *testing.B) {
	tracker, _ := NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName("euclidean"),
		DistanceThreshold:   50.0,
		HitCounterMax:       30,
		InitializationDelay: 3,
		FilterFactory:       NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0),
	})
	detections := createTestDetections(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(detections, 1, nil)
	}
}

func BenchmarkTrackerUpdate_100Objects_NoFilter(b *testing.B) {
	tracker, _ := NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName("euclidean"),
		DistanceThreshold:   50.0,
		HitCounterMax:       30,
		InitializationDelay: 3,
		FilterFactory:       NewNoFilterFactory(),
	})
	detections := createTestDetections(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(detections, 1, nil)
	}
}

// ============================================================================
// Tracker with Different Distance Functions
// ============================================================================

func BenchmarkTrackerUpdate_100Objects_IoU(b *testing.B) {
	tracker, _ := NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName("iou"),
		DistanceThreshold:   0.5,
		HitCounterMax:       30,
		InitializationDelay: 3,
		FilterFactory:       NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0),
	})
	detections := createTestDetections(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Update(detections, 1, nil)
	}
}
