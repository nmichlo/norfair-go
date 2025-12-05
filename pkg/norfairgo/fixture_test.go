package norfairgo

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// ============================================================================
// Fixture JSON Schema
// ============================================================================

type Fixture struct {
	TrackerConfig TrackerConfigJSON `json:"tracker_config"`
	Steps         []Step            `json:"steps"`
}

type TrackerConfigJSON struct {
	DistanceFunction    string  `json:"distance_function"`
	DistanceThreshold   float64 `json:"distance_threshold"`
	HitCounterMax       int     `json:"hit_counter_max"`
	InitializationDelay int     `json:"initialization_delay"`
}

type Step struct {
	FrameID int     `json:"frame_id"`
	Inputs  Inputs  `json:"inputs"`
	Outputs Outputs `json:"outputs"`
}

type Inputs struct {
	Detections []DetectionJSON `json:"detections"`
}

type DetectionJSON struct {
	Bbox          []float64 `json:"bbox"`
	GroundTruthID int       `json:"ground_truth_id"`
}

type Outputs struct {
	TrackedObjects []TrackedObjectJSON `json:"tracked_objects"`
	AllObjects     []TrackedObjectJSON `json:"all_objects"`
}

type TrackedObjectJSON struct {
	ID             *int        `json:"id"`
	InitializingID int         `json:"initializing_id"`
	Estimate       [][]float64 `json:"estimate"`
	Age            int         `json:"age"`
	HitCounter     int         `json:"hit_counter"`
	IsInitializing bool        `json:"is_initializing"`
}

// ============================================================================
// Test Helpers
// ============================================================================

func findTestdataDir() (string, error) {
	candidates := []string{
		"testdata/fixtures",
		"../../testdata/fixtures",
		"../../../testdata/fixtures",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find testdata/fixtures directory")
}

func loadFixture(scenario string) (*Fixture, error) {
	testdataDir, err := findTestdataDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(testdataDir, fmt.Sprintf("fixture_%s.json", scenario))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file %s: %v", path, err)
	}

	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("failed to parse fixture file %s: %v", path, err)
	}

	return &fixture, nil
}

func createTracker(config TrackerConfigJSON) (*Tracker, error) {
	return NewTracker(&TrackerConfig{
		DistanceFunction:    DistanceByName(config.DistanceFunction),
		DistanceThreshold:   config.DistanceThreshold,
		HitCounterMax:       config.HitCounterMax,
		InitializationDelay: config.InitializationDelay,
	})
}

func compareTrackedObjects(
	stepIdx int,
	frameID int,
	expected []TrackedObjectJSON,
	actual []*TrackedObject,
	label string,
	tolerance float64,
) error {
	// First compare counts
	if len(expected) != len(actual) {
		msg := fmt.Sprintf("FIRST DIVERGENCE at step %d (frame_id=%d):\n", stepIdx, frameID)
		msg += fmt.Sprintf("  Expected %s: %d\n", label, len(expected))
		msg += fmt.Sprintf("  Actual %s: %d\n\n", label, len(actual))

		msg += "Expected objects:\n"
		for _, obj := range expected {
			msg += fmt.Sprintf("  ID=%v, initializing_id=%d, estimate=%v, age=%d, hit_counter=%d, is_initializing=%v\n",
				obj.ID, obj.InitializingID, obj.Estimate, obj.Age, obj.HitCounter, obj.IsInitializing)
		}
		msg += "\nActual objects:\n"
		for _, obj := range actual {
			estimate := matrixToSlice(obj.Estimate)
			msg += fmt.Sprintf("  ID=%v, initializing_id=%v, estimate=%v, age=%d, hit_counter=%d, is_initializing=%v\n",
				obj.ID, obj.InitializingID, estimate, obj.Age, obj.HitCounter, obj.IsInitializing)
		}
		return fmt.Errorf("%s", msg)
	}

	// Compare each object
	for i := range expected {
		exp := expected[i]
		act := actual[i]

		// Compare ID (Python and Go both start at 1)
		if !intPtrEqual(exp.ID, act.ID) {
			return fmt.Errorf("Step %d frame %d: Object %d ID mismatch: expected %v, got %v",
				stepIdx, frameID, i, ptrToStr(exp.ID), ptrToStr(act.ID))
		}

		// Compare initializing_id (Python and Go both start at 1)
		var actInitID int
		if act.InitializingID != nil {
			actInitID = *act.InitializingID
		} else {
			actInitID = -1
		}
		if exp.InitializingID != actInitID {
			return fmt.Errorf("Step %d frame %d: Object %d initializing_id mismatch: expected %d, got %d",
				stepIdx, frameID, i, exp.InitializingID, actInitID)
		}

		// Compare age
		if exp.Age != act.Age {
			return fmt.Errorf("Step %d frame %d: Object %d age mismatch: expected %d, got %d",
				stepIdx, frameID, i, exp.Age, act.Age)
		}

		// Compare hit_counter
		if exp.HitCounter != act.HitCounter {
			return fmt.Errorf("Step %d frame %d: Object %d hit_counter mismatch: expected %d, got %d",
				stepIdx, frameID, i, exp.HitCounter, act.HitCounter)
		}

		// Compare is_initializing
		if exp.IsInitializing != act.IsInitializing {
			return fmt.Errorf("Step %d frame %d: Object %d is_initializing mismatch: expected %v, got %v",
				stepIdx, frameID, i, exp.IsInitializing, act.IsInitializing)
		}

		// Compare estimates (with tolerance)
		actEstimate := matrixToSlice(act.Estimate)
		for rowIdx, expRow := range exp.Estimate {
			if rowIdx >= len(actEstimate) {
				return fmt.Errorf("Step %d frame %d: Object %d estimate row %d missing",
					stepIdx, frameID, i, rowIdx)
			}
			for colIdx, expVal := range expRow {
				if colIdx >= len(actEstimate[rowIdx]) {
					return fmt.Errorf("Step %d frame %d: Object %d estimate[%d][%d] missing",
						stepIdx, frameID, i, rowIdx, colIdx)
				}
				actVal := actEstimate[rowIdx][colIdx]
				diff := math.Abs(expVal - actVal)
				if diff > tolerance {
					return fmt.Errorf("Step %d frame %d: Object %d estimate[%d][%d] mismatch: expected %f, got %f (diff=%f)",
						stepIdx, frameID, i, rowIdx, colIdx, expVal, actVal, diff)
				}
			}
		}
	}

	return nil
}

func matrixToSlice(m *mat.Dense) [][]float64 {
	if m == nil {
		return nil
	}
	rows, cols := m.Dims()
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = m.At(i, j)
		}
	}
	return result
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrToStr(p *int) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *p)
}

// ============================================================================
// Fixture Test Runner
// ============================================================================

func runFixtureTest(t *testing.T, scenario string) {
	fixture, err := loadFixture(scenario)
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	tracker, err := createTracker(fixture.TrackerConfig)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	// Reset the global counter for deterministic testing
	ResetGlobalCount()

	// Tolerance for numerical comparisons
	// Start strict per CLAUDE.md guidance - loosen only if there are real precision differences
	tolerance := 1e-6

	for stepIdx, step := range fixture.Steps {
		// Convert inputs to detections
		detections := make([]*Detection, 0, len(step.Inputs.Detections))
		for _, det := range step.Inputs.Detections {
			// Create 2x2 matrix for bounding box (top-left, bottom-right)
			points := mat.NewDense(2, 2, []float64{
				det.Bbox[0], det.Bbox[1], // top-left
				det.Bbox[2], det.Bbox[3], // bottom-right
			})
			detection, err := NewDetection(points, nil)
			if err != nil {
				t.Fatalf("Failed to create detection: %v", err)
			}
			detections = append(detections, detection)
		}

		// Update tracker
		trackedObjects := tracker.Update(detections, 1, nil)

		// Compare tracked_objects (returned by update - initialized, active objects)
		if err := compareTrackedObjects(
			stepIdx,
			step.FrameID,
			step.Outputs.TrackedObjects,
			trackedObjects,
			"tracked_objects",
			tolerance,
		); err != nil {
			t.Fatal(err)
		}

		// Compare all_objects (all internal objects including initializing)
		// Sort by initializing_id to ensure consistent ordering for comparison
		allObjects := make([]*TrackedObject, len(tracker.TrackedObjects))
		copy(allObjects, tracker.TrackedObjects)
		sort.Slice(allObjects, func(i, j int) bool {
			iID := -1
			jID := -1
			if allObjects[i].InitializingID != nil {
				iID = *allObjects[i].InitializingID
			}
			if allObjects[j].InitializingID != nil {
				jID = *allObjects[j].InitializingID
			}
			return iID < jID
		})
		if err := compareTrackedObjects(
			stepIdx,
			step.FrameID,
			step.Outputs.AllObjects,
			allObjects,
			"all_objects",
			tolerance,
		); err != nil {
			t.Fatal(err)
		}
	}

	t.Logf("Fixture test '%s' passed: %d steps verified", scenario, len(fixture.Steps))
}

// ============================================================================
// Test Cases
// ============================================================================

func TestFixture_Small(t *testing.T) {
	runFixtureTest(t, "small")
}

func TestFixture_Medium(t *testing.T) {
	runFixtureTest(t, "medium")
}
