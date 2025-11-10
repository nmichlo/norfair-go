package motmetrics

import (
	"testing"

	"github.com/nmichlo/norfair-go/internal/testutil"
)

// ==============================================================================
// TrackLifecycle Tests
// ==============================================================================

// TestNewTrackLifecycle verifies lifecycle initialization
func TestNewTrackLifecycle(t *testing.T) {
	lifecycle := NewTrackLifecycle(42, 10)

	if lifecycle.GTID != 42 {
		t.Errorf("Expected GTID=42, got %d", lifecycle.GTID)
	}
	if lifecycle.FirstFrame != 10 {
		t.Errorf("Expected FirstFrame=10, got %d", lifecycle.FirstFrame)
	}
	if lifecycle.LastFrame != 10 {
		t.Errorf("Expected LastFrame=10, got %d", lifecycle.LastFrame)
	}
	if lifecycle.TrackedFrames != 0 {
		t.Errorf("Expected TrackedFrames=0, got %d", lifecycle.TrackedFrames)
	}
	if lifecycle.DetectedFrames != 0 {
		t.Errorf("Expected DetectedFrames=0, got %d", lifecycle.DetectedFrames)
	}
	if lifecycle.Fragmentations != 0 {
		t.Errorf("Expected Fragmentations=0, got %d", lifecycle.Fragmentations)
	}
	if lifecycle.WasMatched != false {
		t.Errorf("Expected WasMatched=false, got %v", lifecycle.WasMatched)
	}
}

// TestTrackLifecycle_UpdateMatched verifies matched tracking
func TestTrackLifecycle_UpdateMatched(t *testing.T) {
	lifecycle := NewTrackLifecycle(1, 1)

	// Frame 1: First match
	lifecycle.UpdateMatched(1)
	if lifecycle.TrackedFrames != 1 {
		t.Errorf("Frame 1: Expected TrackedFrames=1, got %d", lifecycle.TrackedFrames)
	}
	if lifecycle.DetectedFrames != 1 {
		t.Errorf("Frame 1: Expected DetectedFrames=1, got %d", lifecycle.DetectedFrames)
	}
	if lifecycle.Fragmentations != 0 {
		t.Errorf("Frame 1: Expected no fragmentation on first match, got %d", lifecycle.Fragmentations)
	}
	if lifecycle.WasMatched != true {
		t.Errorf("Frame 1: Expected WasMatched=true, got %v", lifecycle.WasMatched)
	}

	// Frame 2: Consecutive match
	lifecycle.UpdateMatched(2)
	if lifecycle.TrackedFrames != 2 {
		t.Errorf("Frame 2: Expected TrackedFrames=2, got %d", lifecycle.TrackedFrames)
	}
	if lifecycle.Fragmentations != 0 {
		t.Errorf("Frame 2: Expected no fragmentation on consecutive match, got %d", lifecycle.Fragmentations)
	}
}

// TestTrackLifecycle_UpdateMissed verifies missed tracking
func TestTrackLifecycle_UpdateMissed(t *testing.T) {
	lifecycle := NewTrackLifecycle(1, 1)

	lifecycle.UpdateMissed(1)
	if lifecycle.TrackedFrames != 0 {
		t.Errorf("Expected TrackedFrames=0 after miss, got %d", lifecycle.TrackedFrames)
	}
	if lifecycle.DetectedFrames != 1 {
		t.Errorf("Expected DetectedFrames=1, got %d", lifecycle.DetectedFrames)
	}
	if lifecycle.WasMatched != false {
		t.Errorf("Expected WasMatched=false after miss, got %v", lifecycle.WasMatched)
	}
}

// TestTrackLifecycle_Fragmentation verifies fragmentation detection
func TestTrackLifecycle_Fragmentation(t *testing.T) {
	lifecycle := NewTrackLifecycle(1, 1)

	// Frame 1: Match
	lifecycle.UpdateMatched(1)
	if lifecycle.Fragmentations != 0 {
		t.Errorf("Frame 1: Expected no fragmentation, got %d", lifecycle.Fragmentations)
	}

	// Frame 2: Miss (track break)
	lifecycle.UpdateMissed(2)
	if lifecycle.Fragmentations != 0 {
		t.Errorf("Frame 2: Expected no fragmentation on miss, got %d", lifecycle.Fragmentations)
	}

	// Frame 3: Match (fragmentation: miss → match)
	lifecycle.UpdateMatched(3)
	if lifecycle.Fragmentations != 1 {
		t.Errorf("Frame 3: Expected 1 fragmentation (miss → match), got %d", lifecycle.Fragmentations)
	}

	// Frame 4: Miss
	lifecycle.UpdateMissed(4)

	// Frame 5: Match (second fragmentation)
	lifecycle.UpdateMatched(5)
	if lifecycle.Fragmentations != 2 {
		t.Errorf("Frame 5: Expected 2 fragmentations, got %d", lifecycle.Fragmentations)
	}
}

// TestTrackLifecycle_Coverage verifies coverage calculation
func TestTrackLifecycle_Coverage(t *testing.T) {
	lifecycle := NewTrackLifecycle(1, 1)

	// No detections: coverage = 0
	coverage := lifecycle.Coverage()
	testutil.AssertAlmostEqual(t, coverage, 0.0, 1e-10, "Empty lifecycle coverage")

	// 3 tracked out of 5 detected: coverage = 0.6
	lifecycle.UpdateMatched(1) // Tracked
	lifecycle.UpdateMatched(2) // Tracked
	lifecycle.UpdateMissed(3)  // Missed
	lifecycle.UpdateMatched(4) // Tracked
	lifecycle.UpdateMissed(5)  // Missed

	coverage = lifecycle.Coverage()
	testutil.AssertAlmostEqual(t, coverage, 0.6, 1e-10, "3/5 coverage")
}

// ==============================================================================
// MOTAccumulator Tests
// ==============================================================================

// TestNewMOTAccumulator verifies accumulator initialization
func TestNewMOTAccumulator(t *testing.T) {
	acc := NewMOTAccumulator("video1")

	if acc.VideoName != "video1" {
		t.Errorf("Expected VideoName='video1', got '%s'", acc.VideoName)
	}
	if acc.FrameID != 0 {
		t.Errorf("Expected FrameID=0, got %d", acc.FrameID)
	}
	if acc.NumMatches != 0 {
		t.Errorf("Expected NumMatches=0, got %d", acc.NumMatches)
	}
	if acc.PreviousMapping == nil {
		t.Error("Expected PreviousMapping to be initialized")
	}
	if acc.TrackLifecycles == nil {
		t.Error("Expected TrackLifecycles to be initialized")
	}
}

// TestMOTAccumulator_Update_EmptyFrame verifies empty frame handling
func TestMOTAccumulator_Update_EmptyFrame(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Both empty: should increment frame but no events
	acc.Update([][]float64{}, []int{}, [][]float64{}, []int{}, 0.5, mockHungarian)

	if acc.FrameID != 1 {
		t.Errorf("Expected FrameID=1, got %d", acc.FrameID)
	}
	if acc.NumMatches != 0 || acc.NumMisses != 0 || acc.NumFalsePositives != 0 {
		t.Errorf("Expected no events for empty frame, got matches=%d, misses=%d, fp=%d",
			acc.NumMatches, acc.NumMisses, acc.NumFalsePositives)
	}
}

// TestMOTAccumulator_Update_OnlyPredictions verifies false positives
func TestMOTAccumulator_Update_OnlyPredictions(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// No GT, 3 predictions → 3 false positives
	predBBoxes := [][]float64{
		{0, 0, 10, 10},
		{20, 20, 30, 30},
		{40, 40, 50, 50},
	}
	predIDs := []int{1, 2, 3}

	acc.Update([][]float64{}, []int{}, predBBoxes, predIDs, 0.5, mockHungarian)

	if acc.NumFalsePositives != 3 {
		t.Errorf("Expected 3 false positives, got %d", acc.NumFalsePositives)
	}
	if acc.NumMatches != 0 || acc.NumMisses != 0 {
		t.Errorf("Expected no matches/misses, got matches=%d, misses=%d",
			acc.NumMatches, acc.NumMisses)
	}
}

// TestMOTAccumulator_Update_OnlyGT verifies misses
func TestMOTAccumulator_Update_OnlyGT(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// 2 GT, no predictions → 2 misses
	gtBBoxes := [][]float64{
		{0, 0, 10, 10},
		{20, 20, 30, 30},
	}
	gtIDs := []int{1, 2}

	acc.Update(gtBBoxes, gtIDs, [][]float64{}, []int{}, 0.5, mockHungarian)

	if acc.NumMisses != 2 {
		t.Errorf("Expected 2 misses, got %d", acc.NumMisses)
	}
	if acc.NumObjects != 2 {
		t.Errorf("Expected 2 objects, got %d", acc.NumObjects)
	}

	// Verify lifecycles were created and updated
	if len(acc.TrackLifecycles) != 2 {
		t.Errorf("Expected 2 lifecycles, got %d", len(acc.TrackLifecycles))
	}
	for _, gtID := range gtIDs {
		lifecycle := acc.TrackLifecycles[gtID]
		if lifecycle.DetectedFrames != 1 {
			t.Errorf("GT %d: Expected 1 detected frame, got %d", gtID, lifecycle.DetectedFrames)
		}
		if lifecycle.TrackedFrames != 0 {
			t.Errorf("GT %d: Expected 0 tracked frames, got %d", gtID, lifecycle.TrackedFrames)
		}
	}
}

// TestMOTAccumulator_Update_PerfectMatch verifies perfect matches
func TestMOTAccumulator_Update_PerfectMatch(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Perfect match: same GT and predictions
	boxes := [][]float64{
		{0, 0, 10, 10},
		{20, 20, 30, 30},
	}
	ids := []int{1, 2}

	// Mock Hungarian returns all matches
	hungarianFn := func(distances [][]float64, threshold float64) ([][2]int, []int, []int) {
		return [][2]int{{0, 0}, {1, 1}}, []int{}, []int{}
	}

	acc.Update(boxes, ids, boxes, ids, 0.5, hungarianFn)

	if acc.NumMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", acc.NumMatches)
	}
	if acc.NumMisses != 0 || acc.NumFalsePositives != 0 {
		t.Errorf("Expected no misses/fps, got misses=%d, fp=%d",
			acc.NumMisses, acc.NumFalsePositives)
	}

	// TotalDistance should be 0 (perfect overlap)
	testutil.AssertAlmostEqual(t, acc.TotalDistance, 0.0, 1e-10, "Perfect match distance")
}

// TestMOTAccumulator_Update_PartialMatch verifies mixed events
func TestMOTAccumulator_Update_PartialMatch(t *testing.T) {
	acc := NewMOTAccumulator("test")

	gtBBoxes := [][]float64{
		{0, 0, 10, 10},   // Will match
		{20, 20, 30, 30}, // Will miss
	}
	gtIDs := []int{1, 2}

	predBBoxes := [][]float64{
		{0, 0, 10, 10},   // Matches GT 0
		{50, 50, 60, 60}, // False positive
	}
	predIDs := []int{1, 2}

	// Mock Hungarian: GT0↔Pred0 match, GT1 unmatched, Pred1 unmatched
	hungarianFn := func(distances [][]float64, threshold float64) ([][2]int, []int, []int) {
		return [][2]int{{0, 0}}, []int{1}, []int{1}
	}

	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianFn)

	if acc.NumMatches != 1 {
		t.Errorf("Expected 1 match, got %d", acc.NumMatches)
	}
	if acc.NumMisses != 1 {
		t.Errorf("Expected 1 miss, got %d", acc.NumMisses)
	}
	if acc.NumFalsePositives != 1 {
		t.Errorf("Expected 1 false positive, got %d", acc.NumFalsePositives)
	}
	if acc.NumObjects != 2 {
		t.Errorf("Expected 2 objects, got %d", acc.NumObjects)
	}
}

// TestMOTAccumulator_DetectSwitches verifies ID switch detection
func TestMOTAccumulator_DetectSwitches(t *testing.T) {
	acc := NewMOTAccumulator("test")

	boxes := [][]float64{{0, 0, 10, 10}}

	// Frame 1: GT1 → Tracker1
	hungarianFn := func([][]float64, float64) ([][2]int, []int, []int) {
		return [][2]int{{0, 0}}, []int{}, []int{}
	}
	acc.Update(boxes, []int{1}, boxes, []int{1}, 0.5, hungarianFn)

	if acc.NumSwitches != 0 {
		t.Errorf("Frame 1: Expected 0 switches, got %d", acc.NumSwitches)
	}

	// Frame 2: GT1 → Tracker2 (switch!)
	acc.Update(boxes, []int{1}, boxes, []int{2}, 0.5, hungarianFn)

	if acc.NumSwitches != 1 {
		t.Errorf("Frame 2: Expected 1 switch, got %d", acc.NumSwitches)
	}

	// Frame 3: GT1 → Tracker2 (no switch, same as previous)
	acc.Update(boxes, []int{1}, boxes, []int{2}, 0.5, hungarianFn)

	if acc.NumSwitches != 1 {
		t.Errorf("Frame 3: Expected 1 switch total, got %d", acc.NumSwitches)
	}
}

// TestMOTAccumulator_MultiFrame verifies multi-frame accumulation
func TestMOTAccumulator_MultiFrame(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Perfect Hungarian matcher
	hungarianFn := func(distances [][]float64, threshold float64) ([][2]int, []int, []int) {
		numGT := len(distances)
		numPred := 0
		if numGT > 0 {
			numPred = len(distances[0])
		}
		numMatches := numGT
		if numPred < numMatches {
			numMatches = numPred
		}

		matches := make([][2]int, numMatches)
		for i := 0; i < numMatches; i++ {
			matches[i] = [2]int{i, i}
		}

		var unmatchedGT, unmatchedPred []int
		for i := numMatches; i < numGT; i++ {
			unmatchedGT = append(unmatchedGT, i)
		}
		for i := numMatches; i < numPred; i++ {
			unmatchedPred = append(unmatchedPred, i)
		}

		return matches, unmatchedGT, unmatchedPred
	}

	// Frame 1: 2 GT, 2 predictions (2 matches)
	acc.Update(
		[][]float64{{0, 0, 10, 10}, {20, 20, 30, 30}},
		[]int{1, 2},
		[][]float64{{0, 0, 10, 10}, {20, 20, 30, 30}},
		[]int{1, 2},
		0.5,
		hungarianFn,
	)

	// Frame 2: 2 GT, 1 prediction (1 match, 1 miss)
	acc.Update(
		[][]float64{{0, 0, 10, 10}, {20, 20, 30, 30}},
		[]int{1, 2},
		[][]float64{{0, 0, 10, 10}},
		[]int{1},
		0.5,
		hungarianFn,
	)

	// Frame 3: 1 GT, 2 predictions (1 match, 1 FP)
	acc.Update(
		[][]float64{{0, 0, 10, 10}},
		[]int{1},
		[][]float64{{0, 0, 10, 10}, {50, 50, 60, 60}},
		[]int{1, 3},
		0.5,
		hungarianFn,
	)

	// Verify totals: 2+1+1=4 matches, 0+1+0=1 miss, 0+0+1=1 FP
	if acc.NumMatches != 4 {
		t.Errorf("Expected 4 total matches, got %d", acc.NumMatches)
	}
	if acc.NumMisses != 1 {
		t.Errorf("Expected 1 total miss, got %d", acc.NumMisses)
	}
	if acc.NumFalsePositives != 1 {
		t.Errorf("Expected 1 total FP, got %d", acc.NumFalsePositives)
	}
	if acc.NumObjects != 5 { // Frame1: 2, Frame2: 2, Frame3: 1
		t.Errorf("Expected 5 total objects, got %d", acc.NumObjects)
	}
}

// ==============================================================================
// Extended Metrics Tests
// ==============================================================================

// TestComputeExtendedMetrics_MostlyTracked verifies MT classification
func TestComputeExtendedMetrics_MostlyTracked(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Create lifecycle with 80% coverage (MT threshold)
	lifecycle1 := NewTrackLifecycle(1, 1)
	for i := 0; i < 8; i++ {
		lifecycle1.UpdateMatched(i + 1)
	}
	for i := 0; i < 2; i++ {
		lifecycle1.UpdateMissed(i + 9)
	}

	// Create lifecycle with 100% coverage
	lifecycle2 := NewTrackLifecycle(2, 1)
	for i := 0; i < 10; i++ {
		lifecycle2.UpdateMatched(i + 1)
	}

	acc.TrackLifecycles[1] = lifecycle1
	acc.TrackLifecycles[2] = lifecycle2

	mt, ml, pt, _ := acc.ComputeExtendedMetrics()

	if mt != 2 {
		t.Errorf("Expected 2 MT tracks, got %d", mt)
	}
	if ml != 0 {
		t.Errorf("Expected 0 ML tracks, got %d", ml)
	}
	if pt != 0 {
		t.Errorf("Expected 0 PT tracks, got %d", pt)
	}
}

// TestComputeExtendedMetrics_MostlyLost verifies ML classification
func TestComputeExtendedMetrics_MostlyLost(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Create lifecycle with 0% coverage
	lifecycle1 := NewTrackLifecycle(1, 1)
	for i := 0; i < 10; i++ {
		lifecycle1.UpdateMissed(i + 1)
	}

	// Create lifecycle with 16.67% coverage (1/6, just below 20%)
	lifecycle2 := NewTrackLifecycle(2, 1)
	lifecycle2.UpdateMatched(1)
	for i := 0; i < 5; i++ {
		lifecycle2.UpdateMissed(i + 2)
	}

	acc.TrackLifecycles[1] = lifecycle1
	acc.TrackLifecycles[2] = lifecycle2

	mt, ml, pt, _ := acc.ComputeExtendedMetrics()

	if mt != 0 {
		t.Errorf("Expected 0 MT tracks, got %d", mt)
	}
	if ml != 2 {
		t.Errorf("Expected 2 ML tracks, got %d", ml)
	}
	if pt != 0 {
		t.Errorf("Expected 0 PT tracks, got %d", pt)
	}
}

// TestComputeExtendedMetrics_PartiallyTracked verifies PT classification
func TestComputeExtendedMetrics_PartiallyTracked(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Create lifecycle with 50% coverage
	lifecycle := NewTrackLifecycle(1, 1)
	for i := 0; i < 5; i++ {
		lifecycle.UpdateMatched(i*2 + 1)
		lifecycle.UpdateMissed(i*2 + 2)
	}

	acc.TrackLifecycles[1] = lifecycle

	mt, ml, pt, _ := acc.ComputeExtendedMetrics()

	if mt != 0 {
		t.Errorf("Expected 0 MT tracks, got %d", mt)
	}
	if ml != 0 {
		t.Errorf("Expected 0 ML tracks, got %d", ml)
	}
	if pt != 1 {
		t.Errorf("Expected 1 PT track, got %d", pt)
	}
}

// TestComputeExtendedMetrics_Fragmentations verifies fragmentation counting
func TestComputeExtendedMetrics_Fragmentations(t *testing.T) {
	acc := NewMOTAccumulator("test")

	// Track 1: 2 fragmentations
	lifecycle1 := NewTrackLifecycle(1, 1)
	lifecycle1.UpdateMatched(1)
	lifecycle1.UpdateMissed(2)
	lifecycle1.UpdateMatched(3) // Frag 1
	lifecycle1.UpdateMissed(4)
	lifecycle1.UpdateMatched(5) // Frag 2

	// Track 2: 0 fragmentations
	lifecycle2 := NewTrackLifecycle(2, 1)
	lifecycle2.UpdateMatched(1)
	lifecycle2.UpdateMatched(2)

	acc.TrackLifecycles[1] = lifecycle1
	acc.TrackLifecycles[2] = lifecycle2

	_, _, _, totalFrag := acc.ComputeExtendedMetrics()

	if totalFrag != 2 {
		t.Errorf("Expected 2 total fragmentations, got %d", totalFrag)
	}
}

// ==============================================================================
// Helper Functions
// ==============================================================================

// mockHungarian is a simple mock that returns no matches
func mockHungarian(distances [][]float64, threshold float64) ([][2]int, []int, []int) {
	numGT := len(distances)
	numPred := 0
	if numGT > 0 {
		numPred = len(distances[0])
	}

	unmatchedGT := make([]int, numGT)
	for i := range unmatchedGT {
		unmatchedGT[i] = i
	}

	unmatchedPred := make([]int, numPred)
	for i := range unmatchedPred {
		unmatchedPred[i] = i
	}

	return [][2]int{}, unmatchedGT, unmatchedPred
}
