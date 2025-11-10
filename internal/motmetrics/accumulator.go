// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: MIT
//
// This file contains a Go port of py-motmetrics MOTAccumulator
// Original source: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
//
// Original Copyright (c) 2017-2019 Christoph Heindl, Jack Valmadre
// Original License: MIT
//
// See LICENSE file in this directory and THIRD_PARTY_LICENSES.md in repository root.

package motmetrics

// TrackLifecycle tracks the lifecycle of a single ground truth object.
//
// This is a Go port of py-motmetrics track lifecycle tracking used to compute
// extended metrics like MT (Mostly Tracked), ML (Mostly Lost), PT (Partially
// Tracked), and Frag (Fragmentations).
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
type TrackLifecycle struct {
	GTID int // Ground truth object ID

	// Lifetime span
	FirstFrame int // First frame where this GT object appeared
	LastFrame  int // Last frame where this GT object appeared

	// Tracking quality
	TrackedFrames  int  // Number of frames where object was matched
	DetectedFrames int  // Number of frames where GT object existed
	Fragmentations int  // Number of track breaks (miss → match transitions)
	WasMatched     bool // Was object matched in previous frame? (for frag detection)
}

// NewTrackLifecycle creates a new lifecycle tracker for a GT object.
func NewTrackLifecycle(gtID int, firstFrame int) *TrackLifecycle {
	return &TrackLifecycle{
		GTID:           gtID,
		FirstFrame:     firstFrame,
		LastFrame:      firstFrame,
		TrackedFrames:  0,
		DetectedFrames: 0,
		Fragmentations: 0,
		WasMatched:     false,
	}
}

// UpdateMatched updates lifecycle when object is matched (tracked).
func (tl *TrackLifecycle) UpdateMatched(frameID int) {
	tl.LastFrame = frameID
	tl.DetectedFrames++
	tl.TrackedFrames++

	// Detect fragmentation: was missed, now matched again
	if !tl.WasMatched && tl.DetectedFrames > 1 {
		tl.Fragmentations++
	}

	tl.WasMatched = true
}

// UpdateMissed updates lifecycle when object is missed (not tracked).
func (tl *TrackLifecycle) UpdateMissed(frameID int) {
	tl.LastFrame = frameID
	tl.DetectedFrames++
	tl.WasMatched = false
}

// Coverage returns the proportion of frames where object was tracked.
//
// Returns: tracked_frames / detected_frames
func (tl *TrackLifecycle) Coverage() float64 {
	if tl.DetectedFrames == 0 {
		return 0.0
	}
	return float64(tl.TrackedFrames) / float64(tl.DetectedFrames)
}

// =============================================================================
// MOTAccumulator - Per-Video Tracking Event Accumulation
// =============================================================================

// MOTAccumulator accumulates tracking events for a single video sequence.
//
// This is a Go port of py-motmetrics.MOTAccumulator which implements the core
// accumulation logic for MOTChallenge evaluation. Events (MATCH, MISS, FP,
// SWITCH) are accumulated frame-by-frame and used to compute final metrics
// (MOTA, MOTP, etc.).
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
type MOTAccumulator struct {
	// VideoName identifies this sequence
	VideoName string

	// Event counters (accumulated across all frames)
	NumMatches        int     // True positives (correct matches)
	NumFalsePositives int     // Tracker detections with no GT match
	NumMisses         int     // Ground truth objects with no tracker match
	NumSwitches       int     // ID switches (same GT, different tracker ID)
	TotalDistance     float64 // Sum of IoU distances for MOTP
	NumObjects        int     // Total ground truth objects across all frames

	// ID switch detection (tracks GT→Tracker mapping across frames)
	PreviousMapping map[int]int // map[gtID]trackerID from previous frame
	FrameID         int         // Current frame number (1-indexed)

	// Track lifecycle tracking (for MT/ML/PT/Frag metrics)
	TrackLifecycles map[int]*TrackLifecycle // map[gtID]*lifecycle
}

// NewMOTAccumulator creates a new accumulator for a single video sequence.
//
// Parameters:
//   - videoName: Name of the video sequence being evaluated
//
// Returns: Initialized MOTAccumulator
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
func NewMOTAccumulator(videoName string) *MOTAccumulator {
	return &MOTAccumulator{
		VideoName:       videoName,
		PreviousMapping: make(map[int]int),
		TrackLifecycles: make(map[int]*TrackLifecycle),
		FrameID:         0, // Will increment to 1 on first update
	}
}

// Update processes a single frame, updating event counters.
//
// This is a Go port of py-motmetrics.MOTAccumulator.update() which processes
// a frame by computing IoU distance matrix, performing Hungarian matching,
// and accumulating events.
//
// Parameters:
//   - gtBBoxes: Ground truth bounding boxes [x_min, y_min, x_max, y_max]
//   - gtIDs: Ground truth object IDs
//   - predBBoxes: Predicted bounding boxes
//   - predIDs: Tracker object IDs
//   - threshold: IoU distance threshold for valid match (default 0.5)
//   - hungarianFn: Hungarian matching function (accepts distance matrix and threshold)
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
func (acc *MOTAccumulator) Update(
	gtBBoxes [][]float64,
	gtIDs []int,
	predBBoxes [][]float64,
	predIDs []int,
	threshold float64,
	hungarianFn func([][]float64, float64) ([][2]int, []int, []int),
) {
	acc.FrameID++ // 1-indexed frames (MOTChallenge standard)

	// Edge case: no GT, no predictions
	if len(gtBBoxes) == 0 && len(predBBoxes) == 0 {
		return
	}

	// Edge case: no GT, only predictions → all false positives
	if len(gtBBoxes) == 0 {
		acc.NumFalsePositives += len(predBBoxes)
		return
	}

	// Edge case: no predictions, only GT → all misses
	if len(predBBoxes) == 0 {
		acc.NumMisses += len(gtBBoxes)
		acc.NumObjects += len(gtBBoxes)

		// Update lifecycles: all GT objects are missed
		for _, gtID := range gtIDs {
			lifecycle, exists := acc.TrackLifecycles[gtID]
			if !exists {
				lifecycle = NewTrackLifecycle(gtID, acc.FrameID)
				acc.TrackLifecycles[gtID] = lifecycle
			}
			lifecycle.UpdateMissed(acc.FrameID)
		}
		return
	}

	// Compute IoU distance matrix
	distanceMatrix := ComputeIoUMatrix(gtBBoxes, predBBoxes)

	// Hungarian matching with threshold
	matches, unmatchedGT, unmatchedPred := hungarianFn(distanceMatrix, threshold)

	// Accumulate events
	acc.NumMatches += len(matches)
	acc.NumMisses += len(unmatchedGT)
	acc.NumFalsePositives += len(unmatchedPred)
	acc.NumObjects += len(gtBBoxes)

	// Accumulate distances for MOTP
	for _, match := range matches {
		gtIdx, predIdx := match[0], match[1]
		acc.TotalDistance += distanceMatrix[gtIdx][predIdx]
	}

	// Update lifecycles for matched GT objects
	for _, match := range matches {
		gtIdx := match[0]
		gtID := gtIDs[gtIdx]

		lifecycle, exists := acc.TrackLifecycles[gtID]
		if !exists {
			lifecycle = NewTrackLifecycle(gtID, acc.FrameID)
			acc.TrackLifecycles[gtID] = lifecycle
		}
		lifecycle.UpdateMatched(acc.FrameID)
	}

	// Update lifecycles for missed GT objects
	for _, gtIdx := range unmatchedGT {
		gtID := gtIDs[gtIdx]

		lifecycle, exists := acc.TrackLifecycles[gtID]
		if !exists {
			lifecycle = NewTrackLifecycle(gtID, acc.FrameID)
			acc.TrackLifecycles[gtID] = lifecycle
		}
		lifecycle.UpdateMissed(acc.FrameID)
	}

	// Detect ID switches
	switches := acc.detectSwitches(matches, gtIDs, predIDs)
	acc.NumSwitches += switches
}

// detectSwitches counts ID switches by comparing current to previous frame mappings.
//
// An ID switch occurs when the same GT object is matched to a different tracker ID
// compared to the previous frame.
//
// Parameters:
//   - matches: Current frame matches [gtIdx, predIdx]
//   - gtIDs: Ground truth IDs
//   - predIDs: Tracker IDs
//
// Returns: Number of ID switches detected in this frame
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
func (acc *MOTAccumulator) detectSwitches(matches [][2]int, gtIDs, predIDs []int) int {
	switches := 0
	currentMapping := make(map[int]int)

	for _, match := range matches {
		gtID := gtIDs[match[0]]
		predID := predIDs[match[1]]

		// Check if this GT was tracked in previous frame
		if prevPredID, exists := acc.PreviousMapping[gtID]; exists {
			if prevPredID != predID {
				switches++ // Same GT, different tracker ID = switch
			}
		}
		// Note: First appearance of GT is NOT a switch

		currentMapping[gtID] = predID
	}

	// Update mapping for next frame
	acc.PreviousMapping = currentMapping
	return switches
}

// ComputeExtendedMetrics computes MT/ML/PT/Frag from track lifecycles.
//
// This is a Go port of py-motmetrics extended metrics computation.
//
// Returns:
//   - MT count (Mostly Tracked: coverage >= 80%)
//   - ML count (Mostly Lost: coverage <= 20%)
//   - PT count (Partially Tracked: 20% < coverage < 80%)
//   - Total fragmentations across all tracks
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/mot.py
func (acc *MOTAccumulator) ComputeExtendedMetrics() (int, int, int, int) {
	mt := 0 // Mostly tracked
	ml := 0 // Mostly lost
	pt := 0 // Partially tracked
	totalFragmentations := 0

	for _, lifecycle := range acc.TrackLifecycles {
		coverage := lifecycle.Coverage()

		// MOTChallenge standard:
		// MT: >= 80% (includes boundary), ML: < 20% (excludes boundary), PT: between
		if coverage >= 0.8 {
			mt++
		} else if coverage < 0.2 {
			ml++
		} else {
			pt++
		}

		totalFragmentations += lifecycle.Fragmentations
	}

	return mt, ml, pt, totalFragmentations
}
