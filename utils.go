package norfairgo

import (
	"fmt"
	"image"
	"log"
	"os"
	"sync"

	"gocv.io/x/gocv"
	"golang.org/x/term"
	"gonum.org/v1/gonum/mat"
)

// ValidatePoints ensures points have shape (n_points, n_dims) where n_dims is 2 or 3.
// If points is a 1D array (single point), it reshapes to (1, n_dims).
//
func ValidatePoints(points *mat.Dense) (*mat.Dense, error) {
	rows, cols := points.Dims()

	// Handle 1D case: if input is shape (n,), reshape to (1, n)
	// In gonum, we check if rows==1 and cols>1, meaning it's a row vector that should be a single point
	if rows == 1 && (cols == 2 || cols == 3) {
		// This is already the correct shape for a single 2D or 3D point
		return points, nil
	}

	// Validate dimensions
	if cols != 2 && cols != 3 {
		return nil, fmt.Errorf(
			"invalid points shape: expected n_dims to be 2 or 3, got shape (%d, %d)",
			rows, cols,
		)
	}

	// Valid shape: (n_points, n_dims) where n_dims is 2 or 3
	return points, nil
}

// GetTerminalSize returns the terminal dimensions (columns, lines).
// If terminal size cannot be detected, returns the provided defaults.
//
func GetTerminalSize(defaultCols, defaultLines int) (cols, lines int) {
	// Try to get terminal size from various file descriptors
	// Try stdin (fd 0)
	if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		return width, height
	}

	// Try stdout (fd 1)
	if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return width, height
	}

	// Try stderr (fd 2)
	if width, height, err := term.GetSize(int(os.Stderr.Fd())); err == nil {
		return width, height
	}

	// Fallback to defaults
	return defaultCols, defaultLines
}

// GetCutout extracts a rectangular region from an image based on the bounding box of points.
// The cutout is defined by the minimum and maximum x and y coordinates in the points array.
//
func GetCutout(points *mat.Dense, img gocv.Mat) gocv.Mat {
	rows, cols := points.Dims()
	if rows == 0 || cols < 2 {
		// Return empty mat for invalid points
		return gocv.NewMat()
	}

	// Find bounding box
	minX := points.At(0, 0)
	maxX := points.At(0, 0)
	minY := points.At(0, 1)
	maxY := points.At(0, 1)

	for i := 0; i < rows; i++ {
		x := points.At(i, 0)
		y := points.At(i, 1)

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	// Convert to integer coordinates
	x1 := int(minX)
	y1 := int(minY)
	x2 := int(maxX) + 1 // +1 to include the max point
	y2 := int(maxY) + 1

	// Clamp to image bounds
	imgHeight := img.Rows()
	imgWidth := img.Cols()

	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > imgWidth {
		x2 = imgWidth
	}
	if y2 > imgHeight {
		y2 = imgHeight
	}

	// Check for valid region
	if x1 >= x2 || y1 >= y2 {
		// Return empty mat for invalid region
		return gocv.NewMat()
	}

	// Extract region using gocv.Mat.Region()
	rect := image.Rect(x1, y1, x2, y2)
	region := img.Region(rect)
	return region
}

// warnedMessages tracks which messages have been warned about (for warnOnce)
var warnedMessages sync.Map

// WarnOnce prints a warning message only once (thread-safe).
// Subsequent calls with the same message are ignored.
//
func WarnOnce(message string) {
	if _, loaded := warnedMessages.LoadOrStore(message, true); !loaded {
		log.Printf("WARNING: %s", message)
	}
}

// AnyTrue returns true if any element in the slice is true.
// Returns false for empty slices.
//
func AnyTrue(values []bool) bool {
	for _, v := range values {
		if v {
			return true
		}
	}
	return false
}
