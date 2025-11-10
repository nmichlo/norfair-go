package drawing

import (
	"image"
	"math"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go/internal/numpy"
	"github.com/nmichlo/norfair-go"
)

// =============================================================================
// Helper Functions
// =============================================================================

// GetPointsToDrawFunc is a function that extracts points to draw from an estimate.
// It takes a matrix of estimate points and returns a slice of image Points.
type GetPointsToDrawFunc func(*mat.Dense) []image.Point

// defaultGetPointsToDraw returns the centroid of all points in the estimate.
// This is the default behavior for both Paths and AbsolutePaths when no custom function is provided.
func defaultGetPointsToDraw(estimate *mat.Dense) []image.Point {
	rows, cols := estimate.Dims()
	if rows == 0 || cols < 2 {
		return []image.Point{}
	}

	// Calculate centroid (mean of all points)
	var sumX, sumY float64
	for i := 0; i < rows; i++ {
		sumX += estimate.At(i, 0)
		sumY += estimate.At(i, 1)
	}

	centroidX := int(sumX / float64(rows))
	centroidY := int(sumY / float64(rows))

	return []image.Point{{X: centroidX, Y: centroidY}}
}

// Note: linspace moved to internal/numpy package

// =============================================================================
// Paths (for static cameras)
// =============================================================================

// Paths draws motion trails for tracked objects using relative coordinates.
// It accumulates circles on a mask that fades over time, creating a trail effect.
// This is designed for static cameras only - it will warn if used with camera motion.
//
type Paths struct {
	getPointsToDraw    GetPointsToDrawFunc
	thickness          *int
	color              *Color
	radius             *int
	attenuation        float64
	attenuationFactor  float64
	mask               *gocv.Mat // Accumulated trail (lazy init)
	drawer             *Drawer
	palette            *Palette
	warnedCameraMotion bool
}

// NewPaths creates a new Paths drawer for motion trail visualization.
//
// Parameters:
//   - getPointsToDraw: Function to extract points from estimate (nil = use centroid)
//   - thickness: Circle line thickness (nil = auto-calculate)
//   - color: Circle color (nil = palette color by ID)
//   - radius: Circle radius (nil = auto-calculate)
//   - attenuation: Fade rate in [0, 1] where 0=never fades, 1=instant fade (default 0.01)
//
func NewPaths(
	getPointsToDraw GetPointsToDrawFunc,
	thickness *int,
	color *Color,
	radius *int,
	attenuation float64,
) *Paths {
	// Set default getPointsToDraw if not provided
	if getPointsToDraw == nil {
		getPointsToDraw = defaultGetPointsToDraw
	}

	return &Paths{
		getPointsToDraw:   getPointsToDraw,
		thickness:         thickness,
		color:             color,
		radius:            radius,
		attenuation:       attenuation,
		attenuationFactor: 1.0 - attenuation,
		mask:              nil, // Lazy init
		drawer:            NewDrawer(),
		palette:           NewPalette(nil),
	}
}

// Draw updates the path visualization and returns a new frame.
// The returned frame is the input frame with paths alpha-blended on top.
//
// Parameters:
//   - frame: Input frame to draw on
//   - trackedObjects: List of TrackedObject pointers
//
// Returns: New Mat with paths drawn (caller must Close() when done)
//
func (p *Paths) Draw(frame *gocv.Mat, trackedObjects []*norfairgo.TrackedObject) gocv.Mat {
	// Lazy initialization of mask
	if p.mask == nil {
		// Calculate frame scale for auto-sizing
		frameScale := float64(frame.Rows()) / 100.0

		// Auto-calculate radius if not set
		if p.radius == nil {
			r := int(math.Max(frameScale*0.7, 1.0))
			p.radius = &r
		}

		// Auto-calculate thickness if not set
		if p.thickness == nil {
			t := int(math.Max(frameScale/7.0, 1.0))
			p.thickness = &t
		}

		// Create mask with same size as frame
		mask := gocv.NewMatWithSize(frame.Rows(), frame.Cols(), frame.Type())
		p.mask = &mask
	}

	// Fade the mask by multiplying by attenuationFactor
	p.mask.MultiplyFloat(float32(p.attenuationFactor))

	// Draw current positions for each tracked object
	for _, obj := range trackedObjects {
		// Warn once if used with camera motion (incompatible)
		if obj.AbsToRel != nil && !p.warnedCameraMotion {
			norfairgo.WarnOnce("Paths is not compatible with camera motion. Use AbsolutePaths instead.")
			p.warnedCameraMotion = true
		}

		// Select color (palette by ID if no custom color)
		var objColor Color
		if p.color != nil {
			objColor = *p.color
		} else {
			objColor = p.palette.ChooseColor(obj.GetID())
		}

		// Extract points to draw (relative coordinates)
		estimate, err := obj.GetEstimate(false)
		if err != nil {
			continue // Skip if estimate fails
		}
		pointsToDraw := p.getPointsToDraw(estimate)

		// Draw circles at each point
		for _, point := range pointsToDraw {
			p.drawer.Circle(p.mask, point, *p.radius, *p.thickness, objColor)
		}
	}

	// Alpha blend mask with frame (both weighted equally, alpha=1, beta=1, gamma=0)
	result := p.drawer.AlphaBlend(p.mask, frame, 1.0, 1.0, 0.0)
	return result
}

// Close releases the internal mask Mat.
// This should be called when the Paths drawer is no longer needed.
func (p *Paths) Close() {
	if p.mask != nil {
		p.mask.Close()
		p.mask = nil
	}
}

// =============================================================================
// AbsolutePaths (for moving cameras)
// =============================================================================

// AbsolutePaths draws motion trails for tracked objects using absolute world coordinates.
// It stores historical positions and transforms them to the current camera frame.
// This supports camera motion by using coordinate transformations.
//
type AbsolutePaths struct {
	getPointsToDraw GetPointsToDrawFunc
	thickness       *int
	color           *Color
	radius          *int
	maxHistory      int
	pastPoints      map[int][][]image.Point // Object ID -> history of absolute positions
	alphas          []float64                // Alpha values for each history step
	drawer          *Drawer
	palette         *Palette
}

// NewAbsolutePaths creates a new AbsolutePaths drawer for motion trail visualization with camera motion.
//
// Parameters:
//   - getPointsToDraw: Function to extract points from estimate (nil = use centroid)
//   - thickness: Line thickness (nil = auto-calculate)
//   - color: Line color (nil = palette color by ID)
//   - radius: Circle radius for current position (nil = auto-calculate)
//   - maxHistory: Number of past positions to store per object (default 20)
//
func NewAbsolutePaths(
	getPointsToDraw GetPointsToDrawFunc,
	thickness *int,
	color *Color,
	radius *int,
	maxHistory int,
) *AbsolutePaths {
	// Set default getPointsToDraw if not provided
	if getPointsToDraw == nil {
		getPointsToDraw = defaultGetPointsToDraw
	}

	// Set default maxHistory if not provided or invalid
	if maxHistory <= 0 {
		maxHistory = 20
	}

	// Pre-compute alpha values (linearly decreasing from 0.99 to 0.01)
	alphas := numpy.Linspace(0.99, 0.01, maxHistory)

	return &AbsolutePaths{
		getPointsToDraw: getPointsToDraw,
		thickness:       thickness,
		color:           color,
		radius:          radius,
		maxHistory:      maxHistory,
		pastPoints:      make(map[int][][]image.Point),
		alphas:          alphas,
		drawer:          NewDrawer(),
		palette:         NewPalette(nil),
	}
}

// Draw updates the absolute path visualization and returns a new frame.
// The returned frame is the input frame with paths drawn on top.
//
// Parameters:
//   - frame: Input frame to draw on
//   - trackedObjects: List of TrackedObject pointers
//   - coordTransform: Coordinate transformation (must not be nil)
//
// Returns: New Mat with paths drawn (caller must Close() when done)
//
func (ap *AbsolutePaths) Draw(
	frame *gocv.Mat,
	trackedObjects []*norfairgo.TrackedObject,
	coordTransform norfairgo.CoordinateTransformation,
) gocv.Mat {
	// Auto-calculate parameters if not set
	frameScale := float64(frame.Rows()) / 100.0

	if ap.radius == nil {
		r := int(math.Max(frameScale*0.7, 1.0))
		ap.radius = &r
	}

	if ap.thickness == nil {
		t := int(math.Max(frameScale/7.0, 1.0))
		ap.thickness = &t
	}

	// Process each tracked object
	for _, obj := range trackedObjects {
		// Skip objects with no live points
		if !norfairgo.AnyTrue(obj.LivePoints()) {
			continue
		}

		// Select color (palette by ID if no custom color)
		var objColor Color
		if ap.color != nil {
			objColor = *ap.color
		} else {
			objColor = ap.palette.ChooseColor(obj.GetID())
		}

		// Extract absolute points to draw
		absoluteEstimate, err := obj.GetEstimate(true)
		if err != nil {
			continue // Skip if estimate fails
		}
		absolutePoints := ap.getPointsToDraw(absoluteEstimate)

		// Draw current position (transform to relative first)
		relativePoints := ap.transformPointsToRelative(absolutePoints, coordTransform)
		for _, point := range relativePoints {
			ap.drawer.Circle(frame, point, *ap.radius, *ap.thickness, objColor)
		}

		// Draw path segments from history
		objID := obj.GetID()
		if objID == nil {
			continue // Skip objects without ID
		}
		objIDVal := *objID
		if history, exists := ap.pastPoints[objIDVal]; exists && len(history) > 0 {
			lastAbsolute := absolutePoints

			for i, pastAbsolute := range history {
				if i >= len(ap.alphas) {
					break
				}

				// Create overlay for this segment
				overlay := frame.Clone()

				// Transform both last and past positions to relative
				lastRelative := ap.transformPointsToRelative(lastAbsolute, coordTransform)
				pastRelative := ap.transformPointsToRelative(pastAbsolute, coordTransform)

				// Draw lines between consecutive positions
				for j := range lastRelative {
					if j < len(pastRelative) {
						ap.drawer.Line(&overlay, lastRelative[j], pastRelative[j], objColor, *ap.thickness)
					}
				}

				// Alpha blend overlay with frame
				alpha := ap.alphas[i]
				blended := ap.drawer.AlphaBlend(&overlay, frame, alpha, 1.0, 0.0)
				overlay.Close()

				// Replace frame with blended result
				frame.Close()
				*frame = blended

				// Move to next segment
				lastAbsolute = pastAbsolute
			}
		}

		// Update history: insert current at front, trim to maxHistory
		if _, exists := ap.pastPoints[objIDVal]; !exists {
			ap.pastPoints[objIDVal] = [][]image.Point{}
		}

		// Insert at front
		ap.pastPoints[objIDVal] = append([][]image.Point{absolutePoints}, ap.pastPoints[objIDVal]...)

		// Trim to maxHistory
		if len(ap.pastPoints[objIDVal]) > ap.maxHistory {
			ap.pastPoints[objIDVal] = ap.pastPoints[objIDVal][:ap.maxHistory]
		}
	}

	return *frame
}

// transformPointsToRelative transforms a slice of absolute points to relative coordinates.
func (ap *AbsolutePaths) transformPointsToRelative(
	points []image.Point,
	coordTransform norfairgo.CoordinateTransformation,
) []image.Point {
	if len(points) == 0 {
		return []image.Point{}
	}

	// Convert []image.Point to mat.Dense
	data := make([]float64, len(points)*2)
	for i, p := range points {
		data[i*2] = float64(p.X)
		data[i*2+1] = float64(p.Y)
	}
	absoluteMat := mat.NewDense(len(points), 2, data)

	// Transform using coordinate transformation
	relativeMat := coordTransform.AbsToRel(absoluteMat)

	// Convert back to []image.Point
	rows, _ := relativeMat.Dims()
	result := make([]image.Point, rows)
	for i := 0; i < rows; i++ {
		result[i] = image.Point{
			X: int(relativeMat.At(i, 0)),
			Y: int(relativeMat.At(i, 1)),
		}
	}

	return result
}
