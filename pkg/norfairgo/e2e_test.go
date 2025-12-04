package norfairgo_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go/pkg/norfairgo"
)

// =============================================================================
// JSON Structures for E2E Test Data
// =============================================================================

// ExpectedOutput represents the Python-generated reference output
type ExpectedOutput struct {
	Metadata Metadata        `json:"metadata"`
	Frames   []ExpectedFrame `json:"frames"`
	Summary  Summary         `json:"summary"`
}

// Metadata contains test configuration info
type Metadata struct {
	Scenario      string                 `json:"scenario"`
	TrackerConfig map[string]interface{} `json:"tracker_config"`
}

// ExpectedFrame represents expected state for a single frame
type ExpectedFrame struct {
	FrameID           int              `json:"frame_id"`
	NumDetections     int              `json:"num_detections"`
	TrackedObjects    []ExpectedObject `json:"tracked_objects"`
	ActiveObjectCount int              `json:"active_object_count"`
	TotalObjectCount  int              `json:"total_object_count"`
}

// ExpectedObject represents an expected tracked object
type ExpectedObject struct {
	ID               int         `json:"id"`
	Estimate         [][]float64 `json:"estimate"`
	Age              int         `json:"age"`
	HitCounter       int         `json:"hit_counter"`
	GroundTruthMatch *int        `json:"ground_truth_match,omitempty"`
}

// Summary contains final aggregate metrics
type Summary struct {
	TotalFrames                int `json:"total_frames"`
	TotalDetections            int `json:"total_detections"`
	TotalTrackedObjectsCreated int `json:"total_tracked_objects_created"`
	FinalActiveObjects         int `json:"final_active_objects"`
	CumulativeTrackedCount     int `json:"cumulative_tracked_count"`
}

// InputScenario represents the benchmark input data
type InputScenario struct {
	Seed          int          `json:"seed"`
	NumObjects    int          `json:"num_objects"`
	NumFrames     int          `json:"num_frames"`
	DetectionProb float64      `json:"detection_prob"`
	NoiseStd      float64      `json:"noise_std"`
	Frames        []InputFrame `json:"frames"`
}

// InputFrame represents a single frame of input data
type InputFrame struct {
	FrameID    int              `json:"frame_id"`
	Detections []InputDetection `json:"detections"`
}

// InputDetection represents a single detection
type InputDetection struct {
	Bbox          []float64 `json:"bbox"`
	GroundTruthID int       `json:"ground_truth_id"`
}

// =============================================================================
// Test Data Loading
// =============================================================================

func findTestdataDir(t *testing.T) string {
	// Try multiple locations relative to the test file
	candidates := []string{
		"../../testdata/e2e",
		"../../../testdata/e2e",
		"testdata/e2e",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	t.Fatalf("Could not find testdata/e2e directory. Tried: %v", candidates)
	return ""
}

func loadExpected(t *testing.T, testdataDir, scenario string) *ExpectedOutput {
	path := filepath.Join(testdataDir, fmt.Sprintf("e2e_expected_%s.json", scenario))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load expected output from %s: %v", path, err)
	}

	var expected ExpectedOutput
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatalf("Failed to parse expected output: %v", err)
	}
	return &expected
}

func loadInput(t *testing.T, testdataDir, scenario string) *InputScenario {
	path := filepath.Join(testdataDir, fmt.Sprintf("%s.json", scenario))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load input from %s: %v", path, err)
	}

	var input InputScenario
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}
	return &input
}

// =============================================================================
// Comparison Helpers
// =============================================================================

func estimatesMatch(actual *mat.Dense, expected [][]float64, tolerance float64) bool {
	rows, cols := actual.Dims()
	if rows != len(expected) {
		return false
	}
	for i := 0; i < rows; i++ {
		if cols != len(expected[i]) {
			return false
		}
		for j := 0; j < cols; j++ {
			if math.Abs(actual.At(i, j)-expected[i][j]) > tolerance {
				return false
			}
		}
	}
	return true
}

func reportDivergenceDetails(t *testing.T, frameIdx int, expected ExpectedFrame, actual []*norfairgo.TrackedObject) {
	t.Logf("=== DIVERGENCE DETAILS at frame %d ===", frameIdx)
	t.Logf("Expected %d objects, got %d", len(expected.TrackedObjects), len(actual))

	t.Logf("Expected objects:")
	for _, obj := range expected.TrackedObjects {
		t.Logf("  ID=%d, age=%d, hit_counter=%d, estimate=%v",
			obj.ID, obj.Age, obj.HitCounter, obj.Estimate)
	}

	t.Logf("Actual objects:")
	for _, obj := range actual {
		id := -1
		if obj.ID != nil {
			id = *obj.ID
		}
		// Extract estimate values
		rows, cols := obj.Estimate.Dims()
		est := make([][]float64, rows)
		for i := 0; i < rows; i++ {
			est[i] = make([]float64, cols)
			for j := 0; j < cols; j++ {
				est[i][j] = obj.Estimate.At(i, j)
			}
		}
		t.Logf("  ID=%d, age=%d, hit_counter=%d, estimate=%v",
			id, obj.Age, obj.HitCounter, est)
	}
}

// =============================================================================
// E2E Test Runner
// =============================================================================

func runE2ETest(t *testing.T, scenario string) {
	testdataDir := findTestdataDir(t)
	expected := loadExpected(t, testdataDir, scenario)
	input := loadInput(t, testdataDir, scenario)

	// Create tracker with same config as Python benchmark
	tracker, err := norfairgo.NewTracker(&norfairgo.TrackerConfig{
		DistanceFunction:    norfairgo.DistanceByName("iou"),
		DistanceThreshold:   0.5,
		HitCounterMax:       15,
		InitializationDelay: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	cumulativeTracked := 0
	firstDivergenceFrame := -1
	const tolerance = 1e-6

	for i, frameData := range input.Frames {
		if i >= len(expected.Frames) {
			t.Fatalf("Input has more frames than expected output")
		}
		expectedFrame := expected.Frames[i]

		// Convert detections to norfair format
		// Bbox is [x1, y1, x2, y2] - convert to 2x2 matrix (2 points x 2 dims)
		detections := make([]*norfairgo.Detection, 0, len(frameData.Detections))
		for _, det := range frameData.Detections {
			// Create 2x2 matrix: row 0 = top-left, row 1 = bottom-right
			points := mat.NewDense(2, 2, []float64{
				det.Bbox[0], det.Bbox[1], // top-left (x1, y1)
				det.Bbox[2], det.Bbox[3], // bottom-right (x2, y2)
			})
			detection, err := norfairgo.NewDetection(points, nil)
			if err != nil {
				t.Fatalf("Frame %d: failed to create detection: %v", i, err)
			}
			detections = append(detections, detection)
		}

		// Update tracker
		tracked := tracker.Update(detections, 1, nil)
		cumulativeTracked += len(tracked)

		// Sort tracked objects by ID for deterministic comparison
		sort.Slice(tracked, func(a, b int) bool {
			idA, idB := -1, -1
			if tracked[a].ID != nil {
				idA = *tracked[a].ID
			}
			if tracked[b].ID != nil {
				idB = *tracked[b].ID
			}
			return idA < idB
		})

		// Compare per-frame: number of active objects
		if len(tracked) != len(expectedFrame.TrackedObjects) {
			if firstDivergenceFrame < 0 {
				firstDivergenceFrame = i
				t.Errorf("FIRST DIVERGENCE at frame %d: expected %d objects, got %d",
					i, len(expectedFrame.TrackedObjects), len(tracked))
				reportDivergenceDetails(t, i, expectedFrame, tracked)
			}
			continue // Skip estimate comparison if counts don't match
		}

		// Compare object estimates (if counts match)
		for j, obj := range tracked {
			expObj := expectedFrame.TrackedObjects[j]

			// Compare ID
			actualID := -1
			if obj.ID != nil {
				actualID = *obj.ID
			}
			if actualID != expObj.ID {
				if firstDivergenceFrame < 0 {
					firstDivergenceFrame = i
					t.Errorf("FIRST DIVERGENCE at frame %d, object %d: ID mismatch (expected %d, got %d)",
						i, j, expObj.ID, actualID)
				}
			}

			// Compare estimate
			if !estimatesMatch(obj.Estimate, expObj.Estimate, tolerance) {
				if firstDivergenceFrame < 0 {
					firstDivergenceFrame = i
					t.Errorf("FIRST DIVERGENCE at frame %d, object %d: estimate mismatch", i, j)
					reportDivergenceDetails(t, i, expectedFrame, tracked)
				}
			}
		}
	}

	// Compare summary metrics
	if tracker.TotalObjectCount() != expected.Summary.TotalTrackedObjectsCreated {
		t.Errorf("Total objects created: expected %d, got %d",
			expected.Summary.TotalTrackedObjectsCreated, tracker.TotalObjectCount())
	}

	if cumulativeTracked != expected.Summary.CumulativeTrackedCount {
		t.Errorf("Cumulative tracked count: expected %d, got %d",
			expected.Summary.CumulativeTrackedCount, cumulativeTracked)
	}

	if firstDivergenceFrame >= 0 {
		t.Logf("First divergence occurred at frame %d", firstDivergenceFrame)
	} else {
		t.Logf("SUCCESS: All %d frames matched expected output", len(input.Frames))
		t.Logf("  Total objects: %d", tracker.TotalObjectCount())
		t.Logf("  Cumulative tracked: %d", cumulativeTracked)
	}
}

// =============================================================================
// Test Functions
// =============================================================================

func TestE2E_Small(t *testing.T) {
	runE2ETest(t, "small")
}

func TestE2E_Medium(t *testing.T) {
	runE2ETest(t, "medium")
}
