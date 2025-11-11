package main

import (
	"fmt"
	"time"

	"github.com/nmichlo/norfair-go/pkg/norfairgo"
	"gonum.org/v1/gonum/mat"
)

// ========================================================================	//
// RNG
// ========================================================================	//

// SimpleRNG implements a simple Linear Congruential Generator (LCG)
// that produces identical sequences in Go and Python
type SimpleRNG struct {
	state uint64
}

// NewSimpleRNG creates a new RNG with the given seed
func NewSimpleRNG(seed int64) *SimpleRNG {
	return &SimpleRNG{state: uint64(seed)}
}

// Next returns the next random uint64
func (r *SimpleRNG) Next() uint64 {
	// LCG parameters from Numerical Recipes
	const (
		a = 1664525
		c = 1013904223
		m = 1 << 32
	)
	r.state = (a*r.state + c) % m
	return r.state
}

// Float64 returns a random float64 in [0.0, 1.0)
func (r *SimpleRNG) Float64() float64 {
	return float64(r.Next()) / float64(1<<32)
}

// ========================================================================	//
// SIM
// ========================================================================	//

// Rectangle represents a moving bounding box in the simulation
type Rectangle struct {
	ID     int     // Ground truth ID
	X      float64 // Center X position
	Y      float64 // Center Y position
	Width  float64 // Box width
	Height float64 // Box height
	VX     float64 // Velocity X (pixels per frame)
	VY     float64 // Velocity Y (pixels per frame)
}

// Simulation manages the rectangle physics environment
type Simulation struct {
	Width      int         // Grid width
	Height     int         // Grid height
	Rectangles []Rectangle // All rectangles
	Rng        *SimpleRNG  // Deterministic RNG
}

// NewSimulation creates a new simulation with N rectangles
func NewSimulation(width, height, numRectangles int, seed int64) *Simulation {
	rng := NewSimpleRNG(seed)

	sim := &Simulation{
		Width:      width,
		Height:     height,
		Rectangles: make([]Rectangle, numRectangles),
		Rng:        rng,
	}

	// Initialize rectangles with random positions and velocities
	for i := 0; i < numRectangles; i++ {
		sim.Rectangles[i] = Rectangle{
			ID:     i + 1, // IDs start at 1
			X:      rng.Float64() * float64(width),
			Y:      rng.Float64() * float64(height),
			Width:  20.0 + rng.Float64()*60.0, // 20-80 pixels
			Height: 20.0 + rng.Float64()*60.0, // 20-80 pixels
			VX:     -5.0 + rng.Float64()*10.0, // -5 to +5 pixels/frame
			VY:     -5.0 + rng.Float64()*10.0, // -5 to +5 pixels/frame
		}
	}

	return sim
}

// Update advances the simulation by one frame
func (s *Simulation) Update() {
	for i := range s.Rectangles {
		rect := &s.Rectangles[i]

		// Update position
		rect.X += rect.VX
		rect.Y += rect.VY

		// Bounce off walls (left/right)
		halfW := rect.Width / 2.0
		if rect.X-halfW < 0 {
			rect.X = halfW
			rect.VX = -rect.VX
		} else if rect.X+halfW > float64(s.Width) {
			rect.X = float64(s.Width) - halfW
			rect.VX = -rect.VX
		}

		// Bounce off walls (top/bottom)
		halfH := rect.Height / 2.0
		if rect.Y-halfH < 0 {
			rect.Y = halfH
			rect.VY = -rect.VY
		} else if rect.Y+halfH > float64(s.Height) {
			rect.Y = float64(s.Height) - halfH
			rect.VY = -rect.VY
		}
	}
}

// GetBoundingBoxes returns bounding boxes in [[x_min, y_min], [x_max, y_max]] format
func (s *Simulation) GetBoundingBoxes() [][][]float64 {
	boxes := make([][][]float64, len(s.Rectangles))

	for i, rect := range s.Rectangles {
		halfW := rect.Width / 2.0
		halfH := rect.Height / 2.0

		xMin := rect.X - halfW
		yMin := rect.Y - halfH
		xMax := rect.X + halfW
		yMax := rect.Y + halfH

		boxes[i] = [][]float64{
			{xMin, yMin},
			{xMax, yMax},
		}
	}

	return boxes
}

// GetGroundTruthIDs returns the ground truth IDs for validation
func (s *Simulation) GetGroundTruthIDs() []int {
	ids := make([]int, len(s.Rectangles))
	for i, rect := range s.Rectangles {
		ids[i] = rect.ID
	}
	return ids
}

// ========================================================================	//
// MAIN
// ========================================================================	//

func runBenchmark(filterName string, filterFactory norfairgo.FilterFactory, numObjects, numFrames int) {
	// Create tracker
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:       norfairgo.DistanceByName("iou"),
		DistanceThreshold:      0.5,
		HitCounterMax:          30,
		InitializationDelay:    3,
		PointwiseHitCounterMax: 4,
		DetectionThreshold:     0.0,
		PastDetectionsLength:   4,
		FilterFactory:          filterFactory,
		ReidHitCounterMax:      nil, // Disable ReID for pure tracking performance
	})
	if err != nil {
		panic(err)
	}

	// Create simulation
	sim := NewSimulation(1920, 1080, numObjects, 42)

	// Warmup (10 frames)
	for i := 0; i < 10; i++ {
		sim.Update()
		boxes := sim.GetBoundingBoxes()
		detections := make([]*norfairgo.Detection, len(boxes))
		for j, box := range boxes {
			points := mat.NewDense(2, 2, []float64{
				box[0][0], box[0][1],
				box[1][0], box[1][1],
			})
			detections[j], _ = norfairgo.NewDetection(points, nil)
		}
		tracker.Update(detections, 1, nil)
	}

	// Benchmark
	totalTimeSim := 0
	totalTimeBench := 0
	for frame := 0; frame < numFrames; frame++ {
		t0 := time.Now()
		sim.Update()
		boxes := sim.GetBoundingBoxes()

		// Create detections from bounding boxes
		t1 := time.Now()
		detections := make([]*norfairgo.Detection, len(boxes))
		for i, box := range boxes {
			points := mat.NewDense(2, 2, []float64{
				box[0][0], box[0][1],
				box[1][0], box[1][1],
			})
			detections[i], _ = norfairgo.NewDetection(points, nil)
		}

		// Track objects
		_ = tracker.Update(detections, 1, nil)

		// update
		t2 := time.Now()
		totalTimeSim += int(t1.Sub(t0).Nanoseconds())
		totalTimeBench += int(t2.Sub(t1).Nanoseconds())
	}

	// Calculate metrics
	fps := float64(numFrames) / (float64(totalTimeBench) / 1000_000_000.0)
	avgTimeMs := (float64(totalTimeBench) / float64(numFrames)) / 1_000_000.0

	fmt.Printf("Go     - Filter: %-15s Objects: %3d  |  FPS: %7.1f  |  Avg: %6.3fms\n",
		filterName, numObjects, fps, avgTimeMs)
}

func main() {
	numFrames := 1000

	// Test configurations
	filters := []struct {
		name    string
		factory norfairgo.FilterFactory
	}{
		{"OptimizedKalman", norfairgo.NewOptimizedKalmanFilterFactory(4.0, 0.1, 10.0, 0.0, 1.0)},
		{"FilterPyKalman", norfairgo.NewFilterPyKalmanFilterFactory(4.0, 0.1, 10.0)},
		{"NoFilter", norfairgo.NewNoFilterFactory()},
	}

	objectCounts := []int{10, 50, 100}

	for _, count := range objectCounts {
		fmt.Printf("\n--- %d Objects, %d frames ---\n", count, numFrames)
		for _, filter := range filters {
			runBenchmark(filter.name, filter.factory, count, numFrames)
		}
	}
}
