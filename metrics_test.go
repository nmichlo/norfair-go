package norfairgo

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go/internal/motmetrics"
)

// =============================================================================
// InformationFile Tests (3 tests)
// =============================================================================

func TestInformationFile_ParseValid(t *testing.T) {
	// Create temporary seqinfo.ini file
	tmpDir := t.TempDir()
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")

	content := `[Sequence]
name=MOT17-02-SDP
imDir=img1
frameRate=30
seqLength=600
imWidth=1920
imHeight=1080
imExt=.jpg
`

	if err := os.WriteFile(seqinfoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	// Test integer extraction
	seqLength, err := inf.SearchInt("seqLength")
	if err != nil {
		t.Errorf("SearchInt(seqLength) failed: %v", err)
	}
	if seqLength != 600 {
		t.Errorf("Expected seqLength=600, got %d", seqLength)
	}

	frameRate, err := inf.SearchInt("frameRate")
	if err != nil {
		t.Errorf("SearchInt(frameRate) failed: %v", err)
	}
	if frameRate != 30 {
		t.Errorf("Expected frameRate=30, got %d", frameRate)
	}

	// Test string extraction
	name, err := inf.SearchString("name")
	if err != nil {
		t.Errorf("SearchString(name) failed: %v", err)
	}
	if name != "MOT17-02-SDP" {
		t.Errorf("Expected name=MOT17-02-SDP, got %s", name)
	}

	imExt, err := inf.SearchString("imExt")
	if err != nil {
		t.Errorf("SearchString(imExt) failed: %v", err)
	}
	if imExt != ".jpg" {
		t.Errorf("Expected imExt=.jpg, got %s", imExt)
	}
}

func TestInformationFile_MissingVariable(t *testing.T) {
	tmpDir := t.TempDir()
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")

	content := `[Sequence]
name=test
frameRate=30
`

	if err := os.WriteFile(seqinfoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	// Should return error for missing variable
	_, err = inf.Search("seqLength")
	if err == nil {
		t.Error("Expected error for missing variable, got nil")
	}
	if !strings.Contains(err.Error(), "couldn't find 'seqLength'") {
		t.Errorf("Expected error message about missing variable, got: %v", err)
	}
}

func TestInformationFile_FileNotFound(t *testing.T) {
	_, err := NewInformationFile("/nonexistent/path/seqinfo.ini")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// =============================================================================
// DetectionFileParser Tests (5 tests)
// =============================================================================

func TestDetectionFileParser_LoadDetections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=3
frameRate=30
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create det/det.txt
	detDir := filepath.Join(tmpDir, "det")
	if err := os.MkdirAll(detDir, 0755); err != nil {
		t.Fatalf("Failed to create det dir: %v", err)
	}

	detPath := filepath.Join(detDir, "det.txt")
	detContent := `1,-1,100,200,50,75,0.9,-1,-1,-1
1,-1,300,400,60,80,0.95,-1,-1,-1
2,-1,110,210,50,75,0.85,-1,-1,-1
3,-1,120,220,50,75,0.8,-1,-1,-1
`
	if err := os.WriteFile(detPath, []byte(detContent), 0644); err != nil {
		t.Fatalf("Failed to create det.txt: %v", err)
	}

	// Parse detections
	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	// Verify sequence length
	if parser.Length() != 3 {
		t.Errorf("Expected length=3, got %d", parser.Length())
	}

	// Collect all detections
	allDetections := make([][]*Detection, 0)
	for detections := range parser.Detections() {
		allDetections = append(allDetections, detections)
	}

	// Verify frame count
	if len(allDetections) != 3 {
		t.Fatalf("Expected 3 frames, got %d", len(allDetections))
	}

	// Verify frame 1 (2 detections)
	if len(allDetections[0]) != 2 {
		t.Errorf("Frame 1: expected 2 detections, got %d", len(allDetections[0]))
	}

	// Verify frame 2 (1 detection)
	if len(allDetections[1]) != 1 {
		t.Errorf("Frame 2: expected 1 detection, got %d", len(allDetections[1]))
	}

	// Verify frame 3 (1 detection)
	if len(allDetections[2]) != 1 {
		t.Errorf("Frame 3: expected 1 detection, got %d", len(allDetections[2]))
	}
}

func TestDetectionFileParser_CoordinateConversion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create det.txt with known coordinates
	// Format: frame,id,bb_left,bb_top,bb_width,bb_height,conf,x,y,z
	// Input: bb_left=100, bb_top=200, bb_width=50, bb_height=75
	// Expected output: x_min=100, y_min=200, x_max=150, y_max=275
	detDir := filepath.Join(tmpDir, "det")
	if err := os.MkdirAll(detDir, 0755); err != nil {
		t.Fatalf("Failed to create det dir: %v", err)
	}

	detPath := filepath.Join(detDir, "det.txt")
	detContent := `1,-1,100,200,50,75,0.9,-1,-1,-1
`
	if err := os.WriteFile(detPath, []byte(detContent), 0644); err != nil {
		t.Fatalf("Failed to create det.txt: %v", err)
	}

	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	// Get first frame detections
	detections := <-parser.Detections()
	if len(detections) != 1 {
		t.Fatalf("Expected 1 detection, got %d", len(detections))
	}

	det := detections[0]

	// Verify points shape (2 points: top-left and bottom-right)
	rows, cols := det.Points.Dims()
	if rows != 2 || cols != 2 {
		t.Errorf("Expected points shape (2,2), got (%d,%d)", rows, cols)
	}

	// Verify coordinates (converted from width/height to corners)
	xMin := det.Points.At(0, 0)
	yMin := det.Points.At(0, 1)
	xMax := det.Points.At(1, 0)
	yMax := det.Points.At(1, 1)

	if xMin != 100 {
		t.Errorf("Expected x_min=100, got %f", xMin)
	}
	if yMin != 200 {
		t.Errorf("Expected y_min=200, got %f", yMin)
	}
	if xMax != 150 { // 100 + 50
		t.Errorf("Expected x_max=150, got %f", xMax)
	}
	if yMax != 275 { // 200 + 75
		t.Errorf("Expected y_max=275, got %f", yMax)
	}

	// Verify scores (confidence replicated for both corners)
	if len(det.Scores) != 2 {
		t.Errorf("Expected 2 scores, got %d", len(det.Scores))
	}
	if det.Scores[0] != 0.9 || det.Scores[1] != 0.9 {
		t.Errorf("Expected scores [0.9, 0.9], got %v", det.Scores)
	}
}

func TestDetectionFileParser_FallbackToGroundTruth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create gt/gt.txt (no det/det.txt)
	gtDir := filepath.Join(tmpDir, "gt")
	if err := os.MkdirAll(gtDir, 0755); err != nil {
		t.Fatalf("Failed to create gt dir: %v", err)
	}

	gtPath := filepath.Join(gtDir, "gt.txt")
	gtContent := `1,1,100,200,50,75,1,-1,-1,-1
`
	if err := os.WriteFile(gtPath, []byte(gtContent), 0644); err != nil {
		t.Fatalf("Failed to create gt.txt: %v", err)
	}

	// Should successfully load from gt.txt
	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	detections := <-parser.Detections()
	if len(detections) != 1 {
		t.Errorf("Expected 1 detection from gt.txt, got %d", len(detections))
	}
}

func TestDetectionFileParser_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create empty det.txt
	detDir := filepath.Join(tmpDir, "det")
	if err := os.MkdirAll(detDir, 0755); err != nil {
		t.Fatalf("Failed to create det dir: %v", err)
	}

	detPath := filepath.Join(detDir, "det.txt")
	if err := os.WriteFile(detPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create det.txt: %v", err)
	}

	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	// Should return empty detections
	detections := <-parser.Detections()
	if len(detections) != 0 {
		t.Errorf("Expected 0 detections, got %d", len(detections))
	}
}

func TestDetectionFileParser_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// No det.txt or gt.txt - should fail
	_, err := NewDetectionFileParser(tmpDir, nil)
	if err == nil {
		t.Error("Expected error when no detection file found, got nil")
	}
}

// =============================================================================
// PredictionsTextFile Tests (4 tests)
// =============================================================================

func TestPredictionsTextFile_WriteFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=2
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	// Create predictions file
	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}
	defer ptf.Close()

	// Create tracked object with bounding box
	// Points: [[x_min, y_min], [x_max, y_max]] = [[100, 200], [150, 275]]
	id := 1
	obj := &TrackedObject{
		ID:       &id,
		Estimate: mat.NewDense(2, 2, []float64{100, 200, 150, 275}),
	}

	// Write first frame
	if err := ptf.Update([]*TrackedObject{obj}, nil); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Write second frame (auto-increment)
	id2 := 2
	obj2 := &TrackedObject{
		ID:       &id2,
		Estimate: mat.NewDense(2, 2, []float64{110, 210, 160, 285}),
	}
	if err := ptf.Update([]*TrackedObject{obj2}, nil); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Close and read file
	if err := ptf.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Read predictions file
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	// Verify format: frame,id,bb_left,bb_top,bb_width,bb_height,-1,-1,-1,-1
	// Frame 1: 1,1,100.000000,200.000000,50.000000,75.000000,-1,-1,-1,-1
	// bb_width = 150 - 100 = 50
	// bb_height = 275 - 200 = 75
	expectedLine1 := "1,1,100.000000,200.000000,50.000000,75.000000,-1,-1,-1,-1"
	if lines[0] != expectedLine1 {
		t.Errorf("Line 1 mismatch:\nExpected: %s\nGot:      %s", expectedLine1, lines[0])
	}

	// Frame 2: 2,2,110.000000,210.000000,50.000000,75.000000,-1,-1,-1,-1
	expectedLine2 := "2,2,110.000000,210.000000,50.000000,75.000000,-1,-1,-1,-1"
	if lines[1] != expectedLine2 {
		t.Errorf("Line 2 mismatch:\nExpected: %s\nGot:      %s", expectedLine2, lines[1])
	}
}

func TestPredictionsTextFile_SkipNoID(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}
	defer ptf.Close()

	// Create tracked object without ID (initializing)
	obj := &TrackedObject{
		ID:       nil, // No ID
		Estimate: mat.NewDense(2, 2, []float64{100, 200, 150, 275}),
	}

	// Should skip this object
	if err := ptf.Update([]*TrackedObject{obj}, nil); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	ptf.Close()

	// Read predictions file
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	// Should be empty (no objects with ID)
	if strings.TrimSpace(string(content)) != "" {
		t.Errorf("Expected empty file, got: %s", string(content))
	}
}

func TestPredictionsTextFile_CustomFrameNumber(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=10
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}
	defer ptf.Close()

	id := 1
	obj := &TrackedObject{
		ID:       &id,
		Estimate: mat.NewDense(2, 2, []float64{100, 200, 150, 275}),
	}

	// Write with custom frame number
	frameNum := 5
	if err := ptf.Update([]*TrackedObject{obj}, &frameNum); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	ptf.Close()

	// Read and verify frame number
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	// Should start with "5," not "1,"
	if !strings.HasPrefix(string(content), "5,") {
		t.Errorf("Expected frame number 5, got: %s", strings.Split(string(content), ",")[0])
	}
}

func TestPredictionsTextFile_AutoClose(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini with seqLength=2
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=2
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}

	id := 1
	obj := &TrackedObject{
		ID:       &id,
		Estimate: mat.NewDense(2, 2, []float64{100, 200, 150, 275}),
	}

	// Write frame 1
	if err := ptf.Update([]*TrackedObject{obj}, nil); err != nil {
		t.Fatalf("Update frame 1 failed: %v", err)
	}

	// Write frame 2 (should trigger auto-close since frameNumber > length after increment)
	if err := ptf.Update([]*TrackedObject{obj}, nil); err != nil {
		t.Fatalf("Update frame 2 failed: %v", err)
	}

	// File should be auto-closed, manual Close should be safe
	if err := ptf.Close(); err != nil {
		t.Errorf("Close after auto-close should be safe, got error: %v", err)
	}

	// Verify file exists and has 2 lines
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
}

// =============================================================================
// Round-Trip Tests (3 tests)
// =============================================================================

func TestRoundTrip_DetectionsToTrackerToPredictions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=2
frameRate=30
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create det.txt
	detDir := filepath.Join(tmpDir, "det")
	if err := os.MkdirAll(detDir, 0755); err != nil {
		t.Fatalf("Failed to create det dir: %v", err)
	}

	detPath := filepath.Join(detDir, "det.txt")
	detContent := `1,-1,100,200,50,75,0.9,-1,-1,-1
2,-1,110,210,50,75,0.85,-1,-1,-1
`
	if err := os.WriteFile(detPath, []byte(detContent), 0644); err != nil {
		t.Fatalf("Failed to create det.txt: %v", err)
	}

	// Load detections
	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	// Simulate tracking (convert detections to tracked objects)
	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}
	defer ptf.Close()

	frameNum := 1
	for detections := range parser.Detections() {
		// Convert detections to tracked objects (mock tracking)
		trackedObjects := make([]*TrackedObject, len(detections))
		for i, det := range detections {
			id := i + 1
			trackedObjects[i] = &TrackedObject{
				ID:       &id,
				Estimate: det.Points, // Use detection points as estimate
			}
		}

		// Write predictions
		if err := ptf.Update(trackedObjects, &frameNum); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		frameNum++
	}

	ptf.Close()

	// Read predictions and verify format
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	// Verify first line format matches input
	// Input: 1,-1,100,200,50,75,0.9,-1,-1,-1
	// Output: 1,1,100.000000,200.000000,50.000000,75.000000,-1,-1,-1,-1
	if !strings.HasPrefix(lines[0], "1,1,100.") {
		t.Errorf("Line 1 format mismatch: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "2,1,110.") {
		t.Errorf("Line 2 format mismatch: %s", lines[1])
	}
}

func TestRoundTrip_CoordinatePreservation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create det.txt with specific coordinates
	detDir := filepath.Join(tmpDir, "det")
	if err := os.MkdirAll(detDir, 0755); err != nil {
		t.Fatalf("Failed to create det dir: %v", err)
	}

	detPath := filepath.Join(detDir, "det.txt")
	// bb_left=123.5, bb_top=456.7, bb_width=89.1, bb_height=123.4
	// Expected: x_max=212.6, y_max=580.1
	detContent := `1,-1,123.5,456.7,89.1,123.4,0.9,-1,-1,-1
`
	if err := os.WriteFile(detPath, []byte(detContent), 0644); err != nil {
		t.Fatalf("Failed to create det.txt: %v", err)
	}

	// Load and write back
	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}
	defer ptf.Close()

	for detections := range parser.Detections() {
		for _, det := range detections {
			id := 1
			obj := &TrackedObject{
				ID:       &id,
				Estimate: det.Points,
			}
			if err := ptf.Update([]*TrackedObject{obj}, nil); err != nil {
				t.Fatalf("Update failed: %v", err)
			}
		}
	}

	ptf.Close()

	// Read predictions
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	// Parse values
	parts := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(parts) != 10 {
		t.Fatalf("Expected 10 CSV fields, got %d", len(parts))
	}

	// Verify coordinates preserved
	// bb_left should be 123.5
	// bb_top should be 456.7
	// bb_width should be 89.1
	// bb_height should be 123.4
	bbLeft := parts[2]
	bbTop := parts[3]
	bbWidth := parts[4]
	bbHeight := parts[5]

	if !strings.HasPrefix(bbLeft, "123.5") {
		t.Errorf("bb_left mismatch: expected 123.5..., got %s", bbLeft)
	}
	if !strings.HasPrefix(bbTop, "456.7") {
		t.Errorf("bb_top mismatch: expected 456.7..., got %s", bbTop)
	}
	if !strings.HasPrefix(bbWidth, "89.1") {
		t.Errorf("bb_width mismatch: expected 89.1..., got %s", bbWidth)
	}
	if !strings.HasPrefix(bbHeight, "123.4") {
		t.Errorf("bb_height mismatch: expected 123.4..., got %s", bbHeight)
	}
}

func TestRoundTrip_MultipleObjects(t *testing.T) {
	tmpDir := t.TempDir()

	// Create seqinfo.ini
	seqinfoPath := filepath.Join(tmpDir, "seqinfo.ini")
	seqinfoContent := `[Sequence]
seqLength=1
`
	if err := os.WriteFile(seqinfoPath, []byte(seqinfoContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create det.txt with 3 objects
	detDir := filepath.Join(tmpDir, "det")
	if err := os.MkdirAll(detDir, 0755); err != nil {
		t.Fatalf("Failed to create det dir: %v", err)
	}

	detPath := filepath.Join(detDir, "det.txt")
	detContent := `1,-1,100,200,50,75,0.9,-1,-1,-1
1,-1,300,400,60,80,0.95,-1,-1,-1
1,-1,500,600,70,90,0.85,-1,-1,-1
`
	if err := os.WriteFile(detPath, []byte(detContent), 0644); err != nil {
		t.Fatalf("Failed to create det.txt: %v", err)
	}

	// Load and write back
	parser, err := NewDetectionFileParser(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewDetectionFileParser failed: %v", err)
	}

	inf, err := NewInformationFile(seqinfoPath)
	if err != nil {
		t.Fatalf("NewInformationFile failed: %v", err)
	}

	ptf, err := NewPredictionsTextFile(tmpDir, tmpDir, inf)
	if err != nil {
		t.Fatalf("NewPredictionsTextFile failed: %v", err)
	}
	defer ptf.Close()

	for detections := range parser.Detections() {
		trackedObjects := make([]*TrackedObject, len(detections))
		for i, det := range detections {
			id := i + 1
			trackedObjects[i] = &TrackedObject{
				ID:       &id,
				Estimate: det.Points,
			}
		}
		if err := ptf.Update(trackedObjects, nil); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	ptf.Close()

	// Read predictions
	predPath := filepath.Join(tmpDir, "predictions", filepath.Base(tmpDir)+".txt")
	content, err := os.ReadFile(predPath)
	if err != nil {
		t.Fatalf("Failed to read predictions file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Verify each object has unique ID
	ids := make(map[string]bool)
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			t.Errorf("Invalid line format: %s", line)
			continue
		}
		id := parts[1]
		if ids[id] {
			t.Errorf("Duplicate ID in same frame: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != 3 {
		t.Errorf("Expected 3 unique IDs, got %d", len(ids))
	}
}

// =============================================================================
// IoU Distance Tests
// =============================================================================

func TestIoUDistance_PerfectOverlap(t *testing.T) {
	// Perfect overlap: IoU = 1.0 → distance = 0.0
	box1 := []float64{100, 100, 200, 200}
	box2 := []float64{100, 100, 200, 200}

	distance := motmetrics.IouDistance(box1, box2)

	if distance != 0.0 {
		t.Errorf("Perfect overlap should have distance 0.0, got %.6f", distance)
	}
}

func TestIoUDistance_NoOverlap(t *testing.T) {
	// No overlap: IoU = 0.0 → distance = 1.0
	box1 := []float64{100, 100, 200, 200}
	box2 := []float64{300, 300, 400, 400}

	distance := motmetrics.IouDistance(box1, box2)

	if distance != 1.0 {
		t.Errorf("No overlap should have distance 1.0, got %.6f", distance)
	}
}

func TestIoUDistance_PartialOverlap(t *testing.T) {
	// Partial overlap: 50% overlap in both dimensions
	// Box1: 100x100 area (100,100 to 200,200)
	// Box2: 100x100 area (150,150 to 250,250)
	// Intersection: 50x50 = 2500
	// Union: 10000 + 10000 - 2500 = 17500
	// IoU = 2500/17500 = 0.142857...
	// Distance = 1 - 0.142857 = 0.857142...
	box1 := []float64{100, 100, 200, 200}
	box2 := []float64{150, 150, 250, 250}

	distance := motmetrics.IouDistance(box1, box2)

	expectedDistance := 1.0 - (2500.0 / 17500.0)
	if math.Abs(distance-expectedDistance) > 1e-6 {
		t.Errorf("Expected distance %.6f, got %.6f", expectedDistance, distance)
	}
}

func TestIoUDistance_InvalidBoxWrongLength(t *testing.T) {
	// Box with wrong number of elements should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for box with wrong length")
		}
	}()

	box1 := []float64{100, 100, 200} // Only 3 elements
	box2 := []float64{100, 100, 200, 200}

	_ = motmetrics.IouDistance(box1, box2)
}

func TestIoUDistance_InvalidBoxCoordinates(t *testing.T) {
	// Box with x_max <= x_min should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid box coordinates")
		}
	}()

	box1 := []float64{200, 100, 100, 200} // x_max < x_min
	box2 := []float64{100, 100, 200, 200}

	_ = motmetrics.IouDistance(box1, box2)
}

func TestComputeIoUMatrix_EmptyGT(t *testing.T) {
	// Empty GT should return empty matrix
	gtBBoxes := [][]float64{}
	predBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
	}

	matrix := motmetrics.ComputeIoUMatrix(gtBBoxes, predBBoxes)

	if len(matrix) != 0 {
		t.Errorf("Expected empty matrix for empty GT, got %d rows", len(matrix))
	}
}

func TestComputeIoUMatrix_EmptyPredictions(t *testing.T) {
	// Empty predictions should return matrix with empty rows
	gtBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
	}
	predBBoxes := [][]float64{}

	matrix := motmetrics.ComputeIoUMatrix(gtBBoxes, predBBoxes)

	if len(matrix) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(matrix))
	}
	for i, row := range matrix {
		if len(row) != 0 {
			t.Errorf("Row %d should be empty, got %d elements", i, len(row))
		}
	}
}

func TestComputeIoUMatrix_SinglePair(t *testing.T) {
	// Single GT, single prediction
	gtBBoxes := [][]float64{{100, 100, 200, 200}}
	predBBoxes := [][]float64{{100, 100, 200, 200}}

	matrix := motmetrics.ComputeIoUMatrix(gtBBoxes, predBBoxes)

	if len(matrix) != 1 || len(matrix[0]) != 1 {
		t.Errorf("Expected 1x1 matrix, got %dx%d", len(matrix), len(matrix[0]))
	}

	if matrix[0][0] != 0.0 {
		t.Errorf("Perfect overlap should have distance 0.0, got %.6f", matrix[0][0])
	}
}

func TestComputeIoUMatrix_RectangularMatrix(t *testing.T) {
	// 3 GT boxes, 5 prediction boxes → 3x5 matrix
	gtBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
		{500, 500, 600, 600},
	}
	predBBoxes := [][]float64{
		{100, 100, 200, 200},
		{150, 150, 250, 250},
		{300, 300, 400, 400},
		{350, 350, 450, 450},
		{700, 700, 800, 800},
	}

	matrix := motmetrics.ComputeIoUMatrix(gtBBoxes, predBBoxes)

	// Verify dimensions
	if len(matrix) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(matrix))
	}
	for i, row := range matrix {
		if len(row) != 5 {
			t.Errorf("Row %d: expected 5 columns, got %d", i, len(row))
		}
	}

	// Verify some specific values
	// GT[0] vs Pred[0]: perfect match → distance 0.0
	if matrix[0][0] != 0.0 {
		t.Errorf("GT[0] vs Pred[0]: expected distance 0.0, got %.6f", matrix[0][0])
	}

	// GT[1] vs Pred[2]: perfect match → distance 0.0
	if matrix[1][2] != 0.0 {
		t.Errorf("GT[1] vs Pred[2]: expected distance 0.0, got %.6f", matrix[1][2])
	}

	// GT[0] vs Pred[4]: no overlap → distance 1.0
	if matrix[0][4] != 1.0 {
		t.Errorf("GT[0] vs Pred[4]: expected distance 1.0, got %.6f", matrix[0][4])
	}
}

// =============================================================================
// Hungarian Matching Tests
// =============================================================================

func TestHungarianMatching_PerfectMatch(t *testing.T) {
	// All distances below threshold, optimal assignment
	// Matrix: 3×3 with diagonal being low cost (0.1), others high (0.9)
	distanceMatrix := [][]float64{
		{0.1, 0.9, 0.9},
		{0.9, 0.1, 0.9},
		{0.9, 0.9, 0.1},
	}
	threshold := 0.5

	matches, unmatchedGT, unmatchedPred := hungarianMatching(distanceMatrix, threshold)

	// Should match all 3 pairs optimally
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// All GT and predictions matched
	if len(unmatchedGT) != 0 {
		t.Errorf("Expected 0 unmatched GT, got %d", len(unmatchedGT))
	}
	if len(unmatchedPred) != 0 {
		t.Errorf("Expected 0 unmatched predictions, got %d", len(unmatchedPred))
	}

	// Verify optimal assignment (diagonal matches)
	expectedMatches := map[int]int{0: 0, 1: 1, 2: 2}
	for _, match := range matches {
		gtIdx, predIdx := match[0], match[1]
		if expectedMatches[gtIdx] != predIdx {
			t.Errorf("Unexpected match: GT[%d] -> Pred[%d]", gtIdx, predIdx)
		}
	}
}

func TestHungarianMatching_ThresholdFiltering(t *testing.T) {
	// Some distances above threshold, some below
	distanceMatrix := [][]float64{
		{0.1, 0.9}, // GT[0]: Pred[0] below threshold (0.1), Pred[1] above (0.9)
		{0.9, 0.2}, // GT[1]: Pred[0] above (0.9), Pred[1] below (0.2)
	}
	threshold := 0.5

	matches, unmatchedGT, unmatchedPred := hungarianMatching(distanceMatrix, threshold)

	// Should match both pairs
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	// No unmatched
	if len(unmatchedGT) != 0 {
		t.Errorf("Expected 0 unmatched GT, got %d", len(unmatchedGT))
	}
	if len(unmatchedPred) != 0 {
		t.Errorf("Expected 0 unmatched predictions, got %d", len(unmatchedPred))
	}
}

func TestHungarianMatching_NoValidMatches(t *testing.T) {
	// All distances above threshold
	distanceMatrix := [][]float64{
		{0.9, 0.9},
		{0.9, 0.9},
	}
	threshold := 0.5

	matches, unmatchedGT, unmatchedPred := hungarianMatching(distanceMatrix, threshold)

	// No matches (all above threshold)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}

	// All unmatched
	if len(unmatchedGT) != 2 {
		t.Errorf("Expected 2 unmatched GT, got %d", len(unmatchedGT))
	}
	if len(unmatchedPred) != 2 {
		t.Errorf("Expected 2 unmatched predictions, got %d", len(unmatchedPred))
	}
}

func TestHungarianMatching_RectangularMatrix(t *testing.T) {
	// 3 GT, 5 predictions (rectangular matrix)
	distanceMatrix := [][]float64{
		{0.1, 0.9, 0.9, 0.9, 0.9}, // GT[0] matches Pred[0]
		{0.9, 0.9, 0.2, 0.9, 0.9}, // GT[1] matches Pred[2]
		{0.9, 0.9, 0.9, 0.9, 0.3}, // GT[2] matches Pred[4]
	}
	threshold := 0.5

	matches, unmatchedGT, unmatchedPred := hungarianMatching(distanceMatrix, threshold)

	// Should match all 3 GT
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// No unmatched GT
	if len(unmatchedGT) != 0 {
		t.Errorf("Expected 0 unmatched GT, got %d", len(unmatchedGT))
	}

	// 2 unmatched predictions (Pred[1], Pred[3])
	if len(unmatchedPred) != 2 {
		t.Errorf("Expected 2 unmatched predictions, got %d", len(unmatchedPred))
	}

	// Verify optimal assignment
	matchMap := make(map[int]int)
	for _, match := range matches {
		matchMap[match[0]] = match[1]
	}

	if matchMap[0] != 0 {
		t.Errorf("Expected GT[0] -> Pred[0], got Pred[%d]", matchMap[0])
	}
	if matchMap[1] != 2 {
		t.Errorf("Expected GT[1] -> Pred[2], got Pred[%d]", matchMap[1])
	}
	if matchMap[2] != 4 {
		t.Errorf("Expected GT[2] -> Pred[4], got Pred[%d]", matchMap[2])
	}
}

func TestHungarianMatching_EmptyMatrix(t *testing.T) {
	// Test 1: Empty GT (0×2)
	distanceMatrix1 := [][]float64{}
	matches1, unmatchedGT1, unmatchedPred1 := hungarianMatching(distanceMatrix1, 0.5)

	if matches1 != nil || unmatchedGT1 != nil || unmatchedPred1 != nil {
		t.Errorf("Empty GT: expected all nil, got matches=%v, unmatchedGT=%v, unmatchedPred=%v",
			matches1, unmatchedGT1, unmatchedPred1)
	}

	// Test 2: Empty predictions (2×0)
	distanceMatrix2 := [][]float64{
		{},
		{},
	}
	matches2, unmatchedGT2, unmatchedPred2 := hungarianMatching(distanceMatrix2, 0.5)

	if len(matches2) != 0 {
		t.Errorf("Empty predictions: expected 0 matches, got %d", len(matches2))
	}
	if len(unmatchedGT2) != 2 {
		t.Errorf("Empty predictions: expected 2 unmatched GT, got %d", len(unmatchedGT2))
	}
	if unmatchedPred2 != nil {
		t.Errorf("Empty predictions: expected nil unmatchedPred, got %v", unmatchedPred2)
	}
}

// ==============================================================================
// MOTAccumulator Tests
// ==============================================================================

// Python equivalent: motmetrics library (py-motmetrics)
//
//	import motmetrics as mm
//
//	acc = mm.MOTAccumulator(auto_id=True)
//	# Update with empty frames
//	acc.update([], [], [])  # gt_ids, pred_ids, distance_matrix
//	# Update with no GT (all FP)
//	acc.update([], [1, 2], np.full((0, 2), np.nan))
//	# Update with no predictions (all misses)
//	acc.update([1, 2], [], np.full((2, 0), np.nan))
//
// Validation: tools/validate_metrics/main.py tests MOTAccumulator equivalence
func TestMOTAccumulator_EmptyFrames(t *testing.T) {
	acc := motmetrics.NewMOTAccumulator("test_video")

	// Test 1: No GT, no predictions
	acc.Update([][]float64{}, []int{}, [][]float64{}, []int{}, 0.5, hungarianMatching)

	if acc.NumMatches != 0 {
		t.Errorf("Empty frame: expected 0 matches, got %d", acc.NumMatches)
	}
	if acc.NumFalsePositives != 0 {
		t.Errorf("Empty frame: expected 0 FP, got %d", acc.NumFalsePositives)
	}
	if acc.NumMisses != 0 {
		t.Errorf("Empty frame: expected 0 misses, got %d", acc.NumMisses)
	}
	if acc.NumObjects != 0 {
		t.Errorf("Empty frame: expected 0 objects, got %d", acc.NumObjects)
	}

	// Test 2: No GT, only predictions (all FP)
	predBBoxes := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	predIDs := []int{1, 2}
	acc.Update([][]float64{}, []int{}, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumMatches != 0 {
		t.Errorf("No GT: expected 0 matches, got %d", acc.NumMatches)
	}
	if acc.NumFalsePositives != 2 {
		t.Errorf("No GT: expected 2 FP, got %d", acc.NumFalsePositives)
	}
	if acc.NumMisses != 0 {
		t.Errorf("No GT: expected 0 misses, got %d", acc.NumMisses)
	}

	// Test 3: Only GT, no predictions (all misses)
	acc2 := motmetrics.NewMOTAccumulator("test_video2")
	gtBBoxes := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	gtIDs := []int{1, 2}
	acc2.Update(gtBBoxes, gtIDs, [][]float64{}, []int{}, 0.5, hungarianMatching)

	if acc2.NumMatches != 0 {
		t.Errorf("No predictions: expected 0 matches, got %d", acc2.NumMatches)
	}
	if acc2.NumFalsePositives != 0 {
		t.Errorf("No predictions: expected 0 FP, got %d", acc2.NumFalsePositives)
	}
	if acc2.NumMisses != 2 {
		t.Errorf("No predictions: expected 2 misses, got %d", acc2.NumMisses)
	}
	if acc2.NumObjects != 2 {
		t.Errorf("No predictions: expected 2 objects, got %d", acc2.NumObjects)
	}
}

// Python equivalent: motmetrics library (py-motmetrics) - perfect tracking scenario
//
//	import motmetrics as mm
//	import numpy as np
//
//	acc = mm.MOTAccumulator(auto_id=True)
//	# Frame 1: GT boxes and predictions match perfectly
//	gt_bboxes = [[100, 100, 200, 200], [300, 300, 400, 400]]
//	pred_bboxes = [[100, 100, 200, 200], [300, 300, 400, 400]]
//	distances = compute_iou_distances(gt_bboxes, pred_bboxes)
//	acc.update([1, 2], [1, 2], distances)
//	# Perfect tracking: MOTA = 1.0, all matches, no FP, no misses, no switches
func TestMOTAccumulator_PerfectTracking(t *testing.T) {
	acc := motmetrics.NewMOTAccumulator("perfect_tracking")

	// Frame 1: 3 GT objects, 3 predictions, perfect matches
	gtBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
		{500, 500, 600, 600},
	}
	gtIDs := []int{1, 2, 3}
	predBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
		{500, 500, 600, 600},
	}
	predIDs := []int{10, 20, 30}

	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumMatches != 3 {
		t.Errorf("Frame 1: expected 3 matches, got %d", acc.NumMatches)
	}
	if acc.NumFalsePositives != 0 {
		t.Errorf("Frame 1: expected 0 FP, got %d", acc.NumFalsePositives)
	}
	if acc.NumMisses != 0 {
		t.Errorf("Frame 1: expected 0 misses, got %d", acc.NumMisses)
	}
	if acc.NumSwitches != 0 {
		t.Errorf("Frame 1: expected 0 switches, got %d", acc.NumSwitches)
	}
	if acc.NumObjects != 3 {
		t.Errorf("Frame 1: expected 3 objects, got %d", acc.NumObjects)
	}
	if acc.TotalDistance != 0.0 {
		t.Errorf("Frame 1: expected 0.0 total distance, got %.6f", acc.TotalDistance)
	}

	// Frame 2: Same objects, same tracker IDs (no switches)
	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumMatches != 6 {
		t.Errorf("Frame 2: expected 6 matches, got %d", acc.NumMatches)
	}
	if acc.NumSwitches != 0 {
		t.Errorf("Frame 2: expected 0 switches (consistent IDs), got %d", acc.NumSwitches)
	}
	if acc.NumObjects != 6 {
		t.Errorf("Frame 2: expected 6 objects, got %d", acc.NumObjects)
	}
}

// Python equivalent: motmetrics library (py-motmetrics) - false positives scenario
//
//	import motmetrics as mm
//
//	acc = mm.MOTAccumulator(auto_id=True)
//	# GT has 1 object, predictions have 2 objects (1 match, 1 FP)
//	gt_bboxes = [[100, 100, 200, 200]]
//	pred_bboxes = [[100, 100, 200, 200], [300, 300, 400, 400]]
//	distances = compute_iou_distances(gt_bboxes, pred_bboxes)
//	acc.update([1], [1, 2], distances)
//	# Result: 1 match, 1 false positive
func TestMOTAccumulator_FalsePositives(t *testing.T) {
	acc := motmetrics.NewMOTAccumulator("false_positives")

	// Frame 1: 2 GT objects, 4 predictions (2 extra)
	gtBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
	}
	gtIDs := []int{1, 2}
	predBBoxes := [][]float64{
		{100, 100, 200, 200}, // Matches GT[0]
		{150, 150, 250, 250}, // Extra prediction (partial overlap with GT[0])
		{300, 300, 400, 400}, // Matches GT[1]
		{700, 700, 800, 800}, // Extra prediction (no overlap)
	}
	predIDs := []int{10, 11, 20, 30}

	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", acc.NumMatches)
	}
	if acc.NumFalsePositives != 2 {
		t.Errorf("Expected 2 FP, got %d", acc.NumFalsePositives)
	}
	if acc.NumMisses != 0 {
		t.Errorf("Expected 0 misses, got %d", acc.NumMisses)
	}
	if acc.NumObjects != 2 {
		t.Errorf("Expected 2 objects, got %d", acc.NumObjects)
	}
}

// Python equivalent: motmetrics library (py-motmetrics) - misses scenario
//
//	import motmetrics as mm
//
//	acc = mm.MOTAccumulator(auto_id=True)
//	# GT has 2 objects, predictions have 1 object (1 match, 1 miss)
//	gt_bboxes = [[100, 100, 200, 200], [300, 300, 400, 400]]
//	pred_bboxes = [[100, 100, 200, 200]]
//	distances = compute_iou_distances(gt_bboxes, pred_bboxes)
//	acc.update([1, 2], [1], distances)
//	# Result: 1 match, 1 miss
func TestMOTAccumulator_Misses(t *testing.T) {
	acc := motmetrics.NewMOTAccumulator("misses")

	// Frame 1: 4 GT objects, 2 predictions (2 missed)
	gtBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
		{500, 500, 600, 600},
		{700, 700, 800, 800},
	}
	gtIDs := []int{1, 2, 3, 4}
	predBBoxes := [][]float64{
		{100, 100, 200, 200}, // Matches GT[0]
		{700, 700, 800, 800}, // Matches GT[3]
	}
	predIDs := []int{10, 40}

	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", acc.NumMatches)
	}
	if acc.NumFalsePositives != 0 {
		t.Errorf("Expected 0 FP, got %d", acc.NumFalsePositives)
	}
	if acc.NumMisses != 2 {
		t.Errorf("Expected 2 misses, got %d", acc.NumMisses)
	}
	if acc.NumObjects != 4 {
		t.Errorf("Expected 4 objects, got %d", acc.NumObjects)
	}
}

// Python equivalent: motmetrics library (py-motmetrics) - ID switches scenario
//
//	import motmetrics as mm
//
//	acc = mm.MOTAccumulator(auto_id=True)
//	# Frame 1: GT object 1 matches pred object 1
//	acc.update([1], [1], distances_frame1)
//	# Frame 2: GT object 1 now matches pred object 2 (ID switch)
//	acc.update([1], [2], distances_frame2)
//	# Result: ID switch detected when same GT object gets different pred ID
func TestMOTAccumulator_IDSwitches(t *testing.T) {
	acc := motmetrics.NewMOTAccumulator("id_switches")

	// Frame 1: GT object 1 tracked by tracker ID 10
	gtBBoxes := [][]float64{{100, 100, 200, 200}}
	gtIDs := []int{1}
	predBBoxes := [][]float64{{100, 100, 200, 200}}
	predIDs := []int{10}

	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumSwitches != 0 {
		t.Errorf("Frame 1: expected 0 switches (first appearance), got %d", acc.NumSwitches)
	}

	// Frame 2: GT object 1 now tracked by tracker ID 20 (SWITCH!)
	predIDs = []int{20}
	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumSwitches != 1 {
		t.Errorf("Frame 2: expected 1 switch, got %d", acc.NumSwitches)
	}

	// Frame 3: GT object 1 still tracked by tracker ID 20 (no switch)
	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumSwitches != 1 {
		t.Errorf("Frame 3: expected 1 switch (no new switch), got %d", acc.NumSwitches)
	}

	// Frame 4: GT object 1 switches back to tracker ID 10 (SWITCH!)
	predIDs = []int{10}
	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, 0.5, hungarianMatching)

	if acc.NumSwitches != 2 {
		t.Errorf("Frame 4: expected 2 switches, got %d", acc.NumSwitches)
	}

	if acc.NumMatches != 4 {
		t.Errorf("Expected 4 matches across 4 frames, got %d", acc.NumMatches)
	}
}

// Python equivalent: motmetrics library (py-motmetrics) - multi-frame scenario
//
//	import motmetrics as mm
//
//	acc = mm.MOTAccumulator(auto_id=True)
//	# Accumulate events across multiple frames
//	for frame_idx in range(num_frames):
//	    gt_bboxes, gt_ids = load_ground_truth(frame_idx)
//	    pred_bboxes, pred_ids = load_predictions(frame_idx)
//	    distances = compute_iou_distances(gt_bboxes, pred_bboxes)
//	    acc.update(gt_ids, pred_ids, distances)
//	# Compute summary metrics
//	mh = mm.metrics.create()
//	summary = mh.compute(acc, metrics=['mota', 'motp', 'precision', 'recall'])
func TestMOTAccumulator_MultiFrame(t *testing.T) {
	acc := motmetrics.NewMOTAccumulator("multi_frame")

	// Frame 1: 2 GT, 2 predictions, perfect match
	gtBBoxes1 := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
	}
	gtIDs1 := []int{1, 2}
	predBBoxes1 := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
	}
	predIDs1 := []int{10, 20}
	acc.Update(gtBBoxes1, gtIDs1, predBBoxes1, predIDs1, 0.5, hungarianMatching)

	// Frame 2: GT 1 moves, GT 2 stays, 1 extra prediction
	gtBBoxes2 := [][]float64{
		{150, 150, 250, 250},
		{300, 300, 400, 400},
	}
	gtIDs2 := []int{1, 2}
	predBBoxes2 := [][]float64{
		{150, 150, 250, 250},
		{300, 300, 400, 400},
		{700, 700, 800, 800}, // False positive
	}
	predIDs2 := []int{10, 20, 30}
	acc.Update(gtBBoxes2, gtIDs2, predBBoxes2, predIDs2, 0.5, hungarianMatching)

	// Frame 3: GT 1 disappears (miss), GT 2 switches tracker
	gtBBoxes3 := [][]float64{
		{300, 300, 400, 400},
	}
	gtIDs3 := []int{2}
	predBBoxes3 := [][]float64{
		{300, 300, 400, 400},
	}
	predIDs3 := []int{25} // ID switch from 20 to 25
	acc.Update(gtBBoxes3, gtIDs3, predBBoxes3, predIDs3, 0.5, hungarianMatching)

	// Frame 4: GT 1 reappears, GT 2 continues
	gtBBoxes4 := [][]float64{
		{200, 200, 300, 300},
		{300, 300, 400, 400},
	}
	gtIDs4 := []int{1, 2}
	predBBoxes4 := [][]float64{
		{200, 200, 300, 300},
		{300, 300, 400, 400},
	}
	predIDs4 := []int{10, 25}
	acc.Update(gtBBoxes4, gtIDs4, predBBoxes4, predIDs4, 0.5, hungarianMatching)

	// Frame 5: All GT lost, only predictions (all FP)
	gtBBoxes5 := [][]float64{}
	gtIDs5 := []int{}
	predBBoxes5 := [][]float64{
		{100, 100, 200, 200},
		{500, 500, 600, 600},
	}
	predIDs5 := []int{40, 50}
	acc.Update(gtBBoxes5, gtIDs5, predBBoxes5, predIDs5, 0.5, hungarianMatching)

	// Verify aggregated results
	// Matches: Frame1=2, Frame2=2, Frame3=1, Frame4=2, Frame5=0 → 7
	if acc.NumMatches != 7 {
		t.Errorf("Expected 7 matches, got %d", acc.NumMatches)
	}

	// FP: Frame1=0, Frame2=1, Frame3=0, Frame4=0, Frame5=2 → 3
	if acc.NumFalsePositives != 3 {
		t.Errorf("Expected 3 FP, got %d", acc.NumFalsePositives)
	}

	// Misses: Frame1=0, Frame2=0, Frame3=0, Frame4=0, Frame5=0 → 0
	// (Note: GT 1 missing in Frame3 is not counted as miss because we only count unmatched GT)
	if acc.NumMisses != 0 {
		t.Errorf("Expected 0 misses, got %d", acc.NumMisses)
	}

	// Switches: Frame3=1 (GT 2: 20→25) → 1
	if acc.NumSwitches != 1 {
		t.Errorf("Expected 1 switch, got %d", acc.NumSwitches)
	}

	// Objects: Frame1=2, Frame2=2, Frame3=1, Frame4=2, Frame5=0 → 7
	if acc.NumObjects != 7 {
		t.Errorf("Expected 7 objects, got %d", acc.NumObjects)
	}

	// TotalDistance: All perfect matches (distance 0.0)
	if acc.TotalDistance != 0.0 {
		t.Errorf("Expected 0.0 total distance, got %.6f", acc.TotalDistance)
	}

	// Verify frame counter
	if acc.FrameID != 5 {
		t.Errorf("Expected frame ID 5, got %d", acc.FrameID)
	}
}

// ==============================================================================
// Accumulators Tests
// ==============================================================================

func TestAccumulators_CreateAndUpdate(t *testing.T) {
	accumulators := NewAccumulators()

	// Create accumulator
	err := accumulators.CreateAccumulator("video1")
	if err != nil {
		t.Fatalf("Failed to create accumulator: %v", err)
	}

	// Attempt to create duplicate (should fail)
	err = accumulators.CreateAccumulator("video1")
	if err == nil {
		t.Errorf("Expected error when creating duplicate accumulator")
	}

	// Update accumulator
	gtBBoxes := [][]float64{{100, 100, 200, 200}}
	gtIDs := []int{1}
	predBBoxes := [][]float64{{100, 100, 200, 200}}
	predIDs := []int{10}

	err = accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video1", 0.5)
	if err != nil {
		t.Fatalf("Failed to update accumulator: %v", err)
	}

	// Update non-existent accumulator (should fail)
	err = accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video_missing", 0.5)
	if err == nil {
		t.Errorf("Expected error when updating non-existent accumulator")
	}
}

func TestAccumulators_MultipleVideos(t *testing.T) {
	accumulators := NewAccumulators()

	// Create accumulators for 3 videos
	accumulators.CreateAccumulator("video1")
	accumulators.CreateAccumulator("video2")
	accumulators.CreateAccumulator("video3")

	// Update video1: 2 matches, 0 FP, 0 misses
	gtBBoxes1 := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	gtIDs1 := []int{1, 2}
	predBBoxes1 := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	predIDs1 := []int{10, 20}
	accumulators.Update(gtBBoxes1, gtIDs1, predBBoxes1, predIDs1, "video1", 0.5)

	// Update video2: 1 match, 1 FP, 0 misses
	gtBBoxes2 := [][]float64{{100, 100, 200, 200}}
	gtIDs2 := []int{1}
	predBBoxes2 := [][]float64{{100, 100, 200, 200}, {500, 500, 600, 600}}
	predIDs2 := []int{10, 30}
	accumulators.Update(gtBBoxes2, gtIDs2, predBBoxes2, predIDs2, "video2", 0.5)

	// Update video3: 0 matches, 0 FP, 2 misses
	gtBBoxes3 := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	gtIDs3 := []int{1, 2}
	predBBoxes3 := [][]float64{}
	predIDs3 := []int{}
	accumulators.Update(gtBBoxes3, gtIDs3, predBBoxes3, predIDs3, "video3", 0.5)

	// Compute aggregated metrics
	metrics, err := accumulators.ComputeMetrics()
	if err != nil {
		t.Fatalf("Failed to compute metrics: %v", err)
	}

	// Total matches: 2 + 1 + 0 = 3
	if metrics.NumMatches != 3 {
		t.Errorf("Expected 3 matches, got %d", metrics.NumMatches)
	}

	// Total FP: 0 + 1 + 0 = 1
	if metrics.NumFalsePositives != 1 {
		t.Errorf("Expected 1 FP, got %d", metrics.NumFalsePositives)
	}

	// Total misses: 0 + 0 + 2 = 2
	if metrics.NumMisses != 2 {
		t.Errorf("Expected 2 misses, got %d", metrics.NumMisses)
	}

	// Total objects: 2 + 1 + 2 = 5
	if metrics.NumObjects != 5 {
		t.Errorf("Expected 5 objects, got %d", metrics.NumObjects)
	}
}

func TestAccumulators_ComputeMetrics_MOTA(t *testing.T) {
	accumulators := NewAccumulators()
	accumulators.CreateAccumulator("video1")

	// Frame 1: 10 GT objects
	// 8 matches, 1 FP, 2 misses, 1 ID switch
	gtBBoxes := make([][]float64, 10)
	gtIDs := make([]int, 10)
	for i := 0; i < 10; i++ {
		x := float64(i * 100)
		gtBBoxes[i] = []float64{x, 0, x + 100, 100}
		gtIDs[i] = i + 1
	}

	// Create predictions: 8 perfect matches + 1 FP
	predBBoxes := make([][]float64, 9)
	predIDs := make([]int, 9)
	for i := 0; i < 8; i++ {
		predBBoxes[i] = gtBBoxes[i]
		predIDs[i] = (i + 1) * 10
	}
	predBBoxes[8] = []float64{9000, 9000, 9100, 9100} // False positive
	predIDs[8] = 999

	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video1", 0.5)

	// Frame 2: Same GT, one ID switch
	predIDs[0] = 999 // Switch GT[0] from ID 10 to 999
	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video1", 0.5)

	metrics, err := accumulators.ComputeMetrics()
	if err != nil {
		t.Fatalf("Failed to compute metrics: %v", err)
	}

	// MOTA = 1 - (FP + FN + IDS) / GT
	// Frame 1: FP=1, FN=2, IDS=0, GT=10
	// Frame 2: FP=1, FN=2, IDS=1, GT=10
	// Total: FP=2, FN=4, IDS=1, GT=20
	// MOTA = 1 - (2 + 4 + 1) / 20 = 1 - 7/20 = 0.65
	expectedMOTA := 1.0 - (2.0+4.0+1.0)/20.0
	if math.Abs(metrics.MOTA-expectedMOTA) > 1e-6 {
		t.Errorf("Expected MOTA %.6f, got %.6f", expectedMOTA, metrics.MOTA)
	}
}

func TestAccumulators_ComputeMetrics_MOTP(t *testing.T) {
	accumulators := NewAccumulators()
	accumulators.CreateAccumulator("video1")

	// Frame 1: 3 GT objects with varying distances
	gtBBoxes := [][]float64{
		{100, 100, 200, 200},
		{300, 300, 400, 400},
		{500, 500, 600, 600},
	}
	gtIDs := []int{1, 2, 3}

	// Predictions with slight offsets
	predBBoxes := [][]float64{
		{100, 100, 200, 200}, // Perfect match (distance 0.0)
		{310, 310, 410, 410}, // Partial overlap
		{550, 550, 650, 650}, // Partial overlap
	}
	predIDs := []int{10, 20, 30}

	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video1", 0.5)

	metrics, err := accumulators.ComputeMetrics()
	if err != nil {
		t.Fatalf("Failed to compute metrics: %v", err)
	}

	// MOTP = sum(distances) / num_matches
	// GT[0] vs Pred[0]: Perfect match (distance 0.0)
	// GT[1] vs Pred[1]: IoU ≈ 0.68 (distance ≈ 0.32, below threshold 0.5, match)
	// GT[2] vs Pred[2]: IoU ≈ 0.143 (distance ≈ 0.857, above threshold 0.5, NO match)
	// Total: 2 matches
	if metrics.NumMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", metrics.NumMatches)
	}

	// MOTP should be in valid range
	if metrics.MOTP < 0.0 || metrics.MOTP > 1.0 {
		t.Errorf("Expected MOTP in range [0.0, 1.0], got %.6f", metrics.MOTP)
	}

	// Verify MOTP calculation: (0.0 + ~0.32) / 2 ≈ 0.16
	expectedMOTP := 0.16
	if math.Abs(metrics.MOTP-expectedMOTP) > 0.05 {
		t.Errorf("Expected MOTP ≈ %.6f, got %.6f", expectedMOTP, metrics.MOTP)
	}
}

func TestAccumulators_ComputeMetrics_EdgeCases(t *testing.T) {
	// Test 1: No objects at all
	accumulators1 := NewAccumulators()
	accumulators1.CreateAccumulator("empty")
	accumulators1.Update([][]float64{}, []int{}, [][]float64{}, []int{}, "empty", 0.5)

	metrics1, err := accumulators1.ComputeMetrics()
	if err != nil {
		t.Fatalf("Test 1 failed: %v", err)
	}

	if metrics1.MOTA != 0.0 {
		t.Errorf("Test 1: Expected MOTA 0.0 (no objects), got %.6f", metrics1.MOTA)
	}
	if !math.IsNaN(metrics1.MOTP) {
		t.Errorf("Test 1: Expected MOTP NaN (no matches), got %.6f", metrics1.MOTP)
	}

	// Test 2: No matches (all GT missed)
	accumulators2 := NewAccumulators()
	accumulators2.CreateAccumulator("no_matches")
	gtBBoxes := [][]float64{{100, 100, 200, 200}}
	gtIDs := []int{1}
	accumulators2.Update(gtBBoxes, gtIDs, [][]float64{}, []int{}, "no_matches", 0.5)

	metrics2, err := accumulators2.ComputeMetrics()
	if err != nil {
		t.Fatalf("Test 2 failed: %v", err)
	}

	// MOTA = 1 - (0 + 1 + 0) / 1 = 0.0
	if metrics2.MOTA != 0.0 {
		t.Errorf("Test 2: Expected MOTA 0.0, got %.6f", metrics2.MOTA)
	}
	if !math.IsNaN(metrics2.MOTP) {
		t.Errorf("Test 2: Expected MOTP NaN (no matches), got %.6f", metrics2.MOTP)
	}

	// Test 3: Precision and Recall edge cases
	accumulators3 := NewAccumulators()
	accumulators3.CreateAccumulator("precision_recall")

	// 2 GT, 3 predictions, 2 matches
	gtBBoxes3 := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	gtIDs3 := []int{1, 2}
	predBBoxes3 := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}, {500, 500, 600, 600}}
	predIDs3 := []int{10, 20, 30}
	accumulators3.Update(gtBBoxes3, gtIDs3, predBBoxes3, predIDs3, "precision_recall", 0.5)

	metrics3, err := accumulators3.ComputeMetrics()
	if err != nil {
		t.Fatalf("Test 3 failed: %v", err)
	}

	// Precision = 2 / (2 + 1) = 0.666...
	expectedPrecision := 2.0 / 3.0
	if math.Abs(metrics3.Precision-expectedPrecision) > 1e-6 {
		t.Errorf("Test 3: Expected Precision %.6f, got %.6f", expectedPrecision, metrics3.Precision)
	}

	// Recall = 2 / 2 = 1.0
	if metrics3.Recall != 1.0 {
		t.Errorf("Test 3: Expected Recall 1.0, got %.6f", metrics3.Recall)
	}
}

func TestAccumulators_PrintMetrics(t *testing.T) {
	accumulators := NewAccumulators()
	accumulators.CreateAccumulator("print_test")

	// Add some data
	gtBBoxes := [][]float64{{100, 100, 200, 200}}
	gtIDs := []int{1}
	predBBoxes := [][]float64{{100, 100, 200, 200}}
	predIDs := []int{10}
	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "print_test", 0.5)

	// Test that PrintMetrics doesn't crash
	err := accumulators.PrintMetrics()
	if err != nil {
		t.Errorf("PrintMetrics failed: %v", err)
	}
}

func TestAccumulators_SaveMetrics(t *testing.T) {
	accumulators := NewAccumulators()
	accumulators.CreateAccumulator("save_test")

	// Add some data
	gtBBoxes := [][]float64{{100, 100, 200, 200}, {300, 300, 400, 400}}
	gtIDs := []int{1, 2}
	predBBoxes := [][]float64{{100, 100, 200, 200}}
	predIDs := []int{10}
	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "save_test", 0.5)

	// Save to temporary file
	tmpFile := filepath.Join(os.TempDir(), "test_metrics.csv")
	defer os.Remove(tmpFile)

	err := accumulators.SaveMetrics(tmpFile)
	if err != nil {
		t.Fatalf("SaveMetrics failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Errorf("Metrics file was not created")
	}

	// Read and verify CSV format
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open metrics file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Check header
	if !scanner.Scan() {
		t.Fatalf("Failed to read header")
	}
	header := scanner.Text()
	expectedHeader := "MOTA,MOTP,Precision,Recall,Matches,FalsePositives,Misses,Switches,Objects"
	if header != expectedHeader {
		t.Errorf("Expected header '%s', got '%s'", expectedHeader, header)
	}

	// Check data row exists
	if !scanner.Scan() {
		t.Fatalf("Failed to read data row")
	}
	dataRow := scanner.Text()
	if len(dataRow) == 0 {
		t.Errorf("Data row is empty")
	}
}

func TestAccumulators_Reset(t *testing.T) {
	accumulators := NewAccumulators()
	accumulators.CreateAccumulator("video1")
	accumulators.CreateAccumulator("video2")

	// Add data to both
	gtBBoxes := [][]float64{{100, 100, 200, 200}}
	gtIDs := []int{1}
	predBBoxes := [][]float64{{100, 100, 200, 200}}
	predIDs := []int{10}
	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video1", 0.5)
	accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, "video2", 0.5)

	// Verify data exists
	metrics1, _ := accumulators.ComputeMetrics()
	if metrics1.NumMatches == 0 {
		t.Errorf("Expected non-zero matches before reset")
	}

	// Reset
	accumulators.Reset()

	// Verify all accumulators cleared
	metrics2, _ := accumulators.ComputeMetrics()
	if metrics2.NumMatches != 0 {
		t.Errorf("Expected 0 matches after reset, got %d", metrics2.NumMatches)
	}
	if metrics2.NumObjects != 0 {
		t.Errorf("Expected 0 objects after reset, got %d", metrics2.NumObjects)
	}

	// Verify can create new accumulators after reset
	err := accumulators.CreateAccumulator("video1")
	if err != nil {
		t.Errorf("Failed to create accumulator after reset: %v", err)
	}
}
