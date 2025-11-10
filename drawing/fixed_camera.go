package drawing

import (
	"image"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go"
)

// FixedCamera stabilizes video based on camera motion by drawing frames on a larger canvas.
// As the camera moves, the frame moves in the opposite direction, stabilizing objects.
//
// WARNING: This only works with TranslationTransformation. Using HomographyTransformation
// will result in unexpected behavior.
//
// WARNING: If using other drawers, always apply this one last. Using other drawers on the
// scaled-up frame will not work as expected.
//
// NOTE: Sometimes the camera moves so far that the result won't fit in the scaled-up frame.
// In this case, a warning will be logged and frames will be cropped to avoid errors.
type FixedCamera struct {
	scale             float64   // Scale factor for background (default 2.0)
	background        *gocv.Mat // Larger frame canvas (lazy init)
	attenuationFactor float64   // 1.0 - attenuation (controls fade speed)
}

// NewFixedCamera creates a new FixedCamera for video stabilization.
//
// Parameters:
//   - scale: Resulting video resolution will be scale * (H, W). Use larger scale if camera moves significantly (default 2.0)
//   - attenuation: Controls how fast older frames fade to black (default 0.05)
func NewFixedCamera(scale, attenuation float64) *FixedCamera {
	// Set defaults if not provided
	if scale <= 0 {
		scale = 2.0
	}
	if attenuation < 0 {
		attenuation = 0.05
	}

	return &FixedCamera{
		scale:             scale,
		background:        nil, // Lazy init
		attenuationFactor: 1.0 - attenuation,
	}
}

// AdjustFrame renders the scaled-up stabilized frame.
//
// Parameters:
//   - frame: Input frame to draw
//   - coordTransform: Coordinate transformation from MotionEstimator
//
// Returns: New Mat with stabilized view (larger canvas)
func (fc *FixedCamera) AdjustFrame(frame *gocv.Mat, coordTransform norfairgo.CoordinateTransformation) gocv.Mat {
	frameHeight := frame.Rows()
	frameWidth := frame.Cols()

	// Initialize background if necessary (lazy init)
	if fc.background == nil {
		// Calculate scaled size
		scaledWidth := int(float64(frameWidth)*fc.scale + 0.5)   // Round
		scaledHeight := int(float64(frameHeight)*fc.scale + 0.5) // Round

		// Create zero-filled background
		bg := gocv.NewMatWithSize(scaledHeight, scaledWidth, frame.Type())
		bg.SetTo(gocv.NewScalar(0, 0, 0, 0))
		fc.background = &bg
	} else {
		// Fade existing background by attenuation factor
		fc.background.MultiplyFloat(float32(fc.attenuationFactor))
	}

	// Calculate top_left anchor point (where to draw frame on background)
	// Aim to draw in center of background, but transformations will move this point
	backgroundHeight := fc.background.Rows()
	backgroundWidth := fc.background.Cols()

	// Center of background minus center of frame
	topLeftY := backgroundHeight/2 - frameHeight/2
	topLeftX := backgroundWidth/2 - frameWidth/2

	// Transform using coordinate transformation
	// Python: coord_transformation.rel_to_abs(top_left[::-1]).round().astype(int)[::-1]
	// [::-1] reverses array (y,x) → (x,y) for transformation, then back to (y,x)
	topLeftPoint := mat.NewDense(1, 2, []float64{float64(topLeftX), float64(topLeftY)})
	transformedPoint := coordTransform.RelToAbs(topLeftPoint)

	// Round and extract coordinates
	topLeftY = int(transformedPoint.At(0, 1) + 0.5)
	topLeftX = int(transformedPoint.At(0, 0) + 0.5)

	// Define box of background that will be updated
	backgroundY0 := topLeftY
	backgroundY1 := topLeftY + frameHeight
	backgroundX0 := topLeftX
	backgroundX1 := topLeftX + frameWidth

	// Define box of frame that will be used
	frameY0 := 0
	frameY1 := frameHeight
	frameX0 := 0
	frameX1 := frameWidth

	// Check if scale is sufficient to cover movement
	// If not, crop the frame to avoid errors
	if backgroundY0 < 0 || backgroundX0 < 0 ||
		backgroundY1 > backgroundHeight || backgroundX1 > backgroundWidth {
		norfairgo.WarnOnce("moving_camera_scale is not enough to cover the range of camera movement, frame will be cropped")

		// Crop left or top of frame if necessary
		if backgroundY0 < 0 {
			frameY0 = -backgroundY0
		}
		if backgroundX0 < 0 {
			frameX0 = -backgroundX0
		}

		// Crop right or bottom of frame if necessary
		// frameY1 = max(min(background_size_y - background_y0, background_y1 - background_y0), 0)
		frameY1 = maxInt(minInt(backgroundHeight-backgroundY0, backgroundY1-backgroundY0), 0)
		frameX1 = maxInt(minInt(backgroundWidth-backgroundX0, backgroundX1-backgroundX0), 0)

		// Handle cases where limits of background become negative
		backgroundY0 = maxInt(backgroundY0, 0)
		backgroundX0 = maxInt(backgroundX0, 0)
		backgroundY1 = maxInt(backgroundY1, 0)
		backgroundX1 = maxInt(backgroundX1, 0)
	}

	// Copy frame region onto background region
	// Python: background[bg_y0:bg_y1, bg_x0:bg_x1, :] = frame[fr_y0:fr_y1, fr_x0:fr_x1, :]
	frameRegion := frame.Region(image.Rect(
		frameX0,
		frameY0,
		frameX1,
		frameY1,
	))
	defer frameRegion.Close()

	backgroundRegion := fc.background.Region(image.Rect(
		backgroundX0,
		backgroundY0,
		backgroundX1,
		backgroundY1,
	))
	defer backgroundRegion.Close()

	// Copy frame to background
	frameRegion.CopyTo(&backgroundRegion)

	// Return the background (not a new Mat, return existing reference)
	return *fc.background
}

// Close releases the background Mat.
// This should be called when the FixedCamera is no longer needed.
func (fc *FixedCamera) Close() {
	if fc.background != nil {
		fc.background.Close()
		fc.background = nil
	}
}

// Getter methods for validation/testing

// Scale returns the scale factor.
func (fc *FixedCamera) Scale() float64 {
	return fc.scale
}

// AttenuationFactor returns the attenuation factor (1.0 - attenuation).
func (fc *FixedCamera) AttenuationFactor() float64 {
	return fc.attenuationFactor
}

// Background returns the background Mat (may be nil if not initialized).
func (fc *FixedCamera) Background() *gocv.Mat {
	return fc.background
}

// Helper function for min (maxInt is already defined in drawer.go)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
