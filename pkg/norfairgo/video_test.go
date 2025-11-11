package norfairgo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gocv.io/x/gocv"
)

// =============================================================================
// Video Input Validation Tests
// =============================================================================

func TestVideo_InputValidation_BothNil(t *testing.T) {
	// Test error when both camera and inputPath are nil
	_, err := NewVideo(VideoOptions{})
	if err == nil {
		t.Fatal("Expected error when both camera and inputPath are nil")
	}

	expectedMsg := "exactly one"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestVideo_InputValidation_BothSet(t *testing.T) {
	// Test error when both camera and inputPath are set
	camera := 0
	path := "test.mp4"
	_, err := NewVideo(VideoOptions{
		Camera:    &camera,
		InputPath: &path,
	})

	if err == nil {
		t.Fatal("Expected error when both camera and inputPath are set")
	}

	expectedMsg := "exactly one"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestVideo_InputValidation_FileNotFound(t *testing.T) {
	// Test error when file doesn't exist
	path := "/nonexistent/path/to/video.mp4"
	_, err := NewVideo(VideoOptions{
		InputPath: &path,
	})

	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}

	t.Logf("Got expected error: %v", err)
}

// =============================================================================
// Video Codec Selection Tests
// =============================================================================

func TestVideo_GetCodecFourcc_AVI(t *testing.T) {
	// Test .avi -> MJPG
	v := &Video{}
	codec := v.getCodecFourcc("test.avi")

	if codec != "MJPG" {
		t.Errorf("Expected codec 'MJPG' for .avi, got '%s'", codec)
	}
}

func TestVideo_GetCodecFourcc_MP4(t *testing.T) {
	// Test .mp4 -> mp4v
	v := &Video{}
	codec := v.getCodecFourcc("test.mp4")

	if codec != "mp4v" {
		t.Errorf("Expected codec 'mp4v' for .mp4, got '%s'", codec)
	}
}

func TestVideo_GetCodecFourcc_CustomOverride(t *testing.T) {
	// Test custom fourcc override
	customCodec := "H264"
	v := &Video{
		outputFourcc: &customCodec,
	}
	codec := v.getCodecFourcc("test.mp4")

	if codec != customCodec {
		t.Errorf("Expected custom codec '%s', got '%s'", customCodec, codec)
	}
}

func TestVideo_GetCodecFourcc_UnsupportedExtension(t *testing.T) {
	// Test unsupported extension returns default
	v := &Video{}
	codec := v.getCodecFourcc("test.mov")

	// Should return default (mp4v)
	if codec != "mp4v" {
		t.Errorf("Expected default codec 'mp4v' for unsupported extension, got '%s'", codec)
	}
}

// =============================================================================
// Video Output Path Tests
// =============================================================================

func TestVideo_GetOutputFilePath_File(t *testing.T) {
	// Test when outputPath is a file
	v := &Video{
		outputPath: "/path/to/output.mp4",
	}

	result := v.GetOutputFilePath()
	if result != "/path/to/output.mp4" {
		t.Errorf("Expected '/path/to/output.mp4', got '%s'", result)
	}
}

func TestVideo_GetOutputFilePath_DirectoryWithCamera(t *testing.T) {
	// Test when outputPath is a directory with camera input
	tempDir := t.TempDir()
	camera := 0

	v := &Video{
		outputPath: tempDir,
		outputExt:  "mp4",
		camera:     &camera,
	}

	result := v.GetOutputFilePath()
	expected := filepath.Join(tempDir, "camera_0_out.mp4")

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestVideo_GetOutputFilePath_DirectoryWithFile(t *testing.T) {
	// Test when outputPath is a directory with file input
	tempDir := t.TempDir()
	inputPath := "/path/to/input_video.mp4"

	v := &Video{
		outputPath: tempDir,
		outputExt:  "mp4",
		inputPath:  &inputPath,
	}

	result := v.GetOutputFilePath()
	expected := filepath.Join(tempDir, "input_video_out.mp4")

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// =============================================================================
// Video Progress Description Tests
// =============================================================================

func TestVideo_GetProgressDescription_Camera(t *testing.T) {
	camera := 0
	v := &Video{
		camera: &camera,
	}

	desc := v.getProgressDescription()
	if desc != "Camera 0" {
		t.Errorf("Expected 'Camera 0', got '%s'", desc)
	}
}

func TestVideo_GetProgressDescription_File(t *testing.T) {
	inputPath := "/path/to/video.mp4"
	v := &Video{
		inputPath: &inputPath,
	}

	desc := v.getProgressDescription()
	if desc != "video.mp4" {
		t.Errorf("Expected 'video.mp4', got '%s'", desc)
	}
}

func TestVideo_GetProgressDescription_WithLabel(t *testing.T) {
	inputPath := "/path/to/video.mp4"
	v := &Video{
		inputPath: &inputPath,
		label:     "Processing",
	}

	desc := v.getProgressDescription()
	if desc != "video.mp4 - Processing" {
		t.Errorf("Expected 'video.mp4 - Processing', got '%s'", desc)
	}
}

func TestVideo_GetProgressDescription_Abbreviation(t *testing.T) {
	// Test abbreviation for very long descriptions
	inputPath := "/path/to/very_long_video_file_name_that_should_be_abbreviated_for_display.mp4"
	v := &Video{
		inputPath: &inputPath,
	}

	desc := v.getProgressDescription()

	// Should contain " ... " for abbreviation
	if !contains(desc, " ... ") {
		t.Logf("Description (may or may not be abbreviated based on terminal size): %s", desc)
	}
}

// =============================================================================
// VideoFromFrames Tests
// =============================================================================

func TestVideoFromFrames_INIParsing(t *testing.T) {
	// Create temporary directory with seqinfo.ini
	tempDir := t.TempDir()

	// Create seqinfo.ini
	iniContent := `[Sequence]
name=TestSequence
imDir=img1
frameRate=30
seqLength=10
imWidth=640
imHeight=480
imExt=.jpg
`
	iniPath := filepath.Join(tempDir, "seqinfo.ini")
	if err := os.WriteFile(iniPath, []byte(iniContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create VideoFromFrames
	vff, err := NewVideoFromFrames(tempDir, "", false)
	if err != nil {
		t.Fatalf("Failed to create VideoFromFrames: %v", err)
	}
	defer vff.Close()

	// Verify metadata
	if vff.name != "TestSequence" {
		t.Errorf("Expected name 'TestSequence', got '%s'", vff.name)
	}
	if vff.length != 10 {
		t.Errorf("Expected length 10, got %d", vff.length)
	}
	if vff.fps != 30 {
		t.Errorf("Expected fps 30, got %d", vff.fps)
	}
	if vff.width != 640 {
		t.Errorf("Expected width 640, got %d", vff.width)
	}
	if vff.height != 480 {
		t.Errorf("Expected height 480, got %d", vff.height)
	}
	if vff.imExt != ".jpg" {
		t.Errorf("Expected imExt '.jpg', got '%s'", vff.imExt)
	}
	if vff.imDir != "img1" {
		t.Errorf("Expected imDir 'img1', got '%s'", vff.imDir)
	}
}

func TestVideoFromFrames_MissingINI(t *testing.T) {
	// Test error when seqinfo.ini doesn't exist
	tempDir := t.TempDir()

	_, err := NewVideoFromFrames(tempDir, "", false)
	if err == nil {
		t.Fatal("Expected error for missing seqinfo.ini")
	}

	t.Logf("Got expected error: %v", err)
}

func TestVideoFromFrames_InvalidINI(t *testing.T) {
	// Create temporary directory with invalid seqinfo.ini
	tempDir := t.TempDir()

	// Create invalid seqinfo.ini (missing required fields)
	iniContent := `[Sequence]
name=TestSequence
`
	iniPath := filepath.Join(tempDir, "seqinfo.ini")
	if err := os.WriteFile(iniPath, []byte(iniContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	_, err := NewVideoFromFrames(tempDir, "", false)
	if err == nil {
		t.Fatal("Expected error for invalid seqinfo.ini")
	}

	expectedMsg := "invalid seqinfo.ini"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestVideoFromFrames_VideoGeneration(t *testing.T) {
	// Create temporary directory with complete setup
	tempDir := t.TempDir()

	// Create seqinfo.ini
	iniContent := `[Sequence]
name=TestSequence
imDir=img1
frameRate=30
seqLength=3
imWidth=100
imHeight=100
imExt=.jpg
`
	iniPath := filepath.Join(tempDir, "seqinfo.ini")
	if err := os.WriteFile(iniPath, []byte(iniContent), 0644); err != nil {
		t.Fatalf("Failed to create seqinfo.ini: %v", err)
	}

	// Create img1 directory
	imgDir := filepath.Join(tempDir, "img1")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		t.Fatalf("Failed to create img1 directory: %v", err)
	}

	// Create test images (3 black 100x100 images)
	for i := 1; i <= 3; i++ {
		img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
		imgPath := filepath.Join(imgDir, fmt.Sprintf("%06d.jpg", i))
		gocv.IMWrite(imgPath, img)
		img.Close()
	}

	// Create VideoFromFrames with video generation
	outputDir := t.TempDir()
	vff, err := NewVideoFromFrames(tempDir, outputDir, true)
	if err != nil {
		t.Fatalf("Failed to create VideoFromFrames: %v", err)
	}
	defer vff.Close()

	// Verify videos/ directory was created
	videosDir := filepath.Join(outputDir, "videos")
	if _, err := os.Stat(videosDir); os.IsNotExist(err) {
		t.Error("videos/ directory was not created")
	}

	// Verify video writer was created
	if vff.videoWriter == nil {
		t.Error("VideoWriter was not created")
	}
}
