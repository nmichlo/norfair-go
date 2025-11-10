package norfairgo

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"gocv.io/x/gocv"
	"gopkg.in/ini.v1"
)

// Video wraps OpenCV VideoCapture and VideoWriter with progress tracking.
// Supports reading from video files or camera devices.
type Video struct {
	// Input (exactly one must be set)
	camera    *int
	inputPath *string

	// OpenCV handles
	videoCapture *gocv.VideoCapture
	videoWriter  *gocv.VideoWriter // Lazy initialization

	// Metadata
	fps        float64
	width      int
	height     int
	frameCount int

	// Output configuration
	outputPath   string
	outputFps    float64
	outputFourcc *string
	outputExt    string

	// Progress tracking
	label        string
	frameCounter int
	startTime    time.Time
	progressBar  *progressbar.ProgressBar

	// Display window
	window *gocv.Window
}

// VideoOptions configures Video creation.
type VideoOptions struct {
	// Input (exactly one must be set)
	Camera    *int
	InputPath *string

	// Output (optional)
	OutputPath   string  // File path or directory (default: ".")
	OutputFps    float64 // Framerate (default: input fps)
	OutputFourcc *string // Codec (default: auto-detect from extension)
	OutputExt    string  // Extension for auto-naming (default: "mp4")
	Label        string  // Progress bar label
}

// NewVideo creates a new Video instance.
// Exactly one of opts.Camera or opts.InputPath must be set.
func NewVideo(opts VideoOptions) (*Video, error) {
	// Validate input: exactly one of camera or inputPath must be set
	if (opts.Camera == nil && opts.InputPath == nil) || (opts.Camera != nil && opts.InputPath != nil) {
		return nil, fmt.Errorf("exactly one of Camera or InputPath must be set")
	}

	v := &Video{
		camera:       opts.Camera,
		inputPath:    opts.InputPath,
		outputPath:   opts.OutputPath,
		outputFps:    opts.OutputFps,
		outputFourcc: opts.OutputFourcc,
		outputExt:    opts.OutputExt,
		label:        opts.Label,
	}

	// Set defaults
	if v.outputPath == "" {
		v.outputPath = "."
	}
	if v.outputExt == "" {
		v.outputExt = "mp4"
	}

	// Create VideoCapture
	var err error
	if opts.Camera != nil {
		// Camera input
		v.videoCapture, err = gocv.OpenVideoCapture(*opts.Camera)
		if err != nil {
			return nil, fmt.Errorf("failed to open camera %d: %w", *opts.Camera, err)
		}
	} else {
		// File input - expand ~ for home directory
		path := *opts.InputPath
		if strings.HasPrefix(path, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, path[1:])
			}
		}

		v.videoCapture, err = gocv.OpenVideoCapture(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open video file %s: %w", path, err)
		}
	}

	// Extract metadata
	v.fps = v.videoCapture.Get(gocv.VideoCaptureFPS)
	v.width = int(v.videoCapture.Get(gocv.VideoCaptureFrameWidth))
	v.height = int(v.videoCapture.Get(gocv.VideoCaptureFrameHeight))
	v.frameCount = int(v.videoCapture.Get(gocv.VideoCaptureFrameCount))

	// Set output fps default
	if v.outputFps == 0 {
		v.outputFps = v.fps
	}

	return v, nil
}

// Frames returns a channel that yields video frames.
// The channel is closed when all frames have been read or an error occurs.
func (v *Video) Frames() <-chan gocv.Mat {
	frames := make(chan gocv.Mat)

	go func() {
		defer close(frames)
		defer v.cleanup()

		v.startTime = time.Now()
		v.frameCounter = 0

		// Setup progress bar
		v.setupProgressBar()

		for {
			frame := gocv.NewMat()
			if ok := v.videoCapture.Read(&frame); !ok {
				frame.Close()
				break
			}

			if frame.Empty() {
				frame.Close()
				break
			}

			v.frameCounter++
			v.updateProgressBar()

			frames <- frame
		}
	}()

	return frames
}

// Write writes a frame to the output video.
// VideoWriter is lazily initialized on first call.
func (v *Video) Write(frame gocv.Mat) error {
	// Lazy initialization of VideoWriter
	if v.videoWriter == nil {
		outputPath := v.GetOutputFilePath()
		codec := v.getCodecFourcc(outputPath)

		var err error
		v.videoWriter, err = gocv.VideoWriterFile(
			outputPath,
			codec,
			v.outputFps,
			frame.Cols(),
			frame.Rows(),
			true, // isColor
		)
		if err != nil {
			return fmt.Errorf("failed to create video writer: %w", err)
		}
	}

	// Write frame
	if err := v.videoWriter.Write(frame); err != nil {
		return fmt.Errorf("failed to write frame: %w", err)
	}

	return nil
}

// Show displays a frame in a window.
// Optional downsampling for slow network connections (X11 forwarding).
func (v *Video) Show(frame gocv.Mat, downsampleRatio float64) int {
	// Create window if not exists
	if v.window == nil {
		v.window = gocv.NewWindow("Output")
	}

	if downsampleRatio != 1.0 && downsampleRatio > 0 {
		// Resize frame
		newWidth := int(float64(frame.Cols()) * downsampleRatio)
		newHeight := int(float64(frame.Rows()) * downsampleRatio)
		resized := gocv.NewMat()
		defer resized.Close()
		gocv.Resize(frame, &resized, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		v.window.IMShow(resized)
	} else {
		v.window.IMShow(frame)
	}

	return v.window.WaitKey(1)
}

// GetOutputFilePath returns the output file path.
// If outputPath is a directory, generates a filename based on input.
func (v *Video) GetOutputFilePath() string {
	// Check if outputPath is a directory
	info, err := os.Stat(v.outputPath)
	if err == nil && info.IsDir() {
		// Auto-generate filename
		var baseName string
		if v.camera != nil {
			baseName = fmt.Sprintf("camera_%d_out", *v.camera)
		} else {
			// Extract filename without extension
			fileName := filepath.Base(*v.inputPath)
			ext := filepath.Ext(fileName)
			baseName = strings.TrimSuffix(fileName, ext) + "_out"
		}
		return filepath.Join(v.outputPath, baseName+"."+v.outputExt)
	}

	// outputPath is a file
	return v.outputPath
}

// getCodecFourcc returns the codec fourcc for the given filename.
// Auto-detects based on extension if not explicitly set.
func (v *Video) getCodecFourcc(filename string) string {
	// Use user-provided fourcc if set
	if v.outputFourcc != nil {
		return *v.outputFourcc
	}

	// Auto-detect from extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".avi":
		return "MJPG" // More cross-platform than XVID
	case ".mp4":
		return "mp4v"
	default:
		// Return a default, but this might fail
		return "mp4v"
	}
}

// setupProgressBar creates and configures the progress bar.
func (v *Video) setupProgressBar() {
	description := v.getProgressDescription()

	if v.camera != nil {
		// Camera: unknown length, no percentage/ETA
		v.progressBar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(description),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("fps"),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionClearOnFinish(),
		)
	} else {
		// File: known length, show percentage and ETA
		v.progressBar = progressbar.NewOptions(v.frameCount,
			progressbar.OptionSetDescription(description),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("fps"),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionClearOnFinish(),
		)
	}
}

// getProgressDescription returns the description for the progress bar.
func (v *Video) getProgressDescription() string {
	var desc string
	if v.camera != nil {
		desc = fmt.Sprintf("Camera %d", *v.camera)
	} else {
		desc = filepath.Base(*v.inputPath)
	}

	if v.label != "" {
		desc = fmt.Sprintf("%s - %s", desc, v.label)
	}

	// Abbreviate if too long (reserve 25 cols for progress bar)
	termCols, _ := GetTerminalSize(80, 24)
	maxLen := termCols - 25
	if len(desc) > maxLen && maxLen > 10 {
		// Truncate middle: "start ... end"
		start := desc[:maxLen/2-2]
		end := desc[len(desc)-(maxLen/2-3):]
		desc = start + " ... " + end
	}

	return desc
}

// updateProgressBar updates the progress bar with current progress.
func (v *Video) updateProgressBar() {
	if v.progressBar != nil {
		v.progressBar.Add(1)
	}
}

// cleanup releases resources.
func (v *Video) cleanup() {
	if v.videoWriter != nil {
		v.videoWriter.Close()
	}
	if v.videoCapture != nil {
		v.videoCapture.Close()
	}
	if v.window != nil {
		v.window.Close()
	}
}

// Close releases all resources.
// Should be called with defer after creating a Video.
func (v *Video) Close() error {
	v.cleanup()
	return nil
}

// VideoFromFrames reads image sequences from MOTChallenge-style directories.
// Expects a seqinfo.ini file with metadata and numbered image files.
type VideoFromFrames struct {
	inputPath  string
	outputPath string
	makeVideo  bool

	// Metadata from seqinfo.ini
	length int
	imExt  string
	imDir  string
	fps    int
	width  int
	height int
	name   string

	// State
	frameNumber int
	videoWriter *gocv.VideoWriter
}

// NewVideoFromFrames creates a new VideoFromFrames instance.
// Reads metadata from seqinfo.ini in the input directory.
func NewVideoFromFrames(inputPath, savePath string, makeVideo bool) (*VideoFromFrames, error) {
	vff := &VideoFromFrames{
		inputPath:   inputPath,
		outputPath:  savePath,
		makeVideo:   makeVideo,
		frameNumber: 0,
	}

	// Default output path
	if vff.outputPath == "" {
		vff.outputPath = "."
	}

	// Parse seqinfo.ini
	iniPath := filepath.Join(inputPath, "seqinfo.ini")
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load seqinfo.ini: %w", err)
	}

	section := cfg.Section("Sequence")

	// Extract metadata
	vff.length = section.Key("seqLength").MustInt(0)
	vff.fps = section.Key("frameRate").MustInt(30)
	vff.width = section.Key("imWidth").MustInt(0)
	vff.height = section.Key("imHeight").MustInt(0)
	vff.imExt = section.Key("imExt").MustString(".jpg")
	vff.imDir = section.Key("imDir").MustString("img1")
	vff.name = section.Key("name").MustString("sequence")

	if vff.length == 0 || vff.width == 0 || vff.height == 0 {
		return nil, fmt.Errorf("invalid seqinfo.ini: missing required fields")
	}

	// Create VideoWriter if requested
	if vff.makeVideo {
		// Create videos/ subdirectory
		videosDir := filepath.Join(vff.outputPath, "videos")
		if err := os.MkdirAll(videosDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create videos directory: %w", err)
		}

		// Create VideoWriter
		outputPath := filepath.Join(videosDir, vff.name+".mp4")
		vff.videoWriter, err = gocv.VideoWriterFile(
			outputPath,
			"mp4v",
			float64(vff.fps),
			vff.width,
			vff.height,
			true,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create video writer: %w", err)
		}
	}

	return vff, nil
}

// Frames returns a channel that yields frames from the image sequence.
func (vff *VideoFromFrames) Frames() <-chan gocv.Mat {
	frames := make(chan gocv.Mat)

	go func() {
		defer close(frames)

		for i := 1; i <= vff.length; i++ {
			// Frame path: {inputPath}/{imDir}/{frame:06d}{imExt}
			framePath := filepath.Join(vff.inputPath, vff.imDir, fmt.Sprintf("%06d%s", i, vff.imExt))

			// Read frame
			frame := gocv.IMRead(framePath, gocv.IMReadColor)
			if frame.Empty() {
				frame.Close()
				continue
			}

			vff.frameNumber = i
			frames <- frame
		}
	}()

	return frames
}

// Update writes a frame to the video if makeVideo is true.
func (vff *VideoFromFrames) Update(frame gocv.Mat) error {
	if vff.videoWriter != nil {
		if err := vff.videoWriter.Write(frame); err != nil {
			return fmt.Errorf("failed to write frame: %w", err)
		}
	}

	gocv.WaitKey(1)

	// Cleanup when done
	if vff.frameNumber >= vff.length {
		vff.Close()
	}

	return nil
}

// Close releases all resources.
func (vff *VideoFromFrames) Close() error {
	if vff.videoWriter != nil {
		vff.videoWriter.Close()
		vff.videoWriter = nil
	}
	return nil
}
