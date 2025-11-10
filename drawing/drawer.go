package drawing

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// Drawer provides stateless drawing primitive functions.
// All methods modify frames in-place and return the modified frame.
//
type Drawer struct{}

// NewDrawer creates a new Drawer instance.
func NewDrawer() *Drawer {
	return &Drawer{}
}

// =============================================================================
// Drawing Primitives
// =============================================================================

// Circle draws a circle on the frame.
// Auto-scales radius and thickness based on frame dimensions if not specified.
//
func (d *Drawer) Circle(frame *gocv.Mat, position image.Point, radius int, thickness int, color Color) {
	// Auto-scale radius if not specified (0 means auto)
	if radius == 0 {
		maxDim := max(frame.Rows(), frame.Cols())
		radius = maxInt(int(float64(maxDim)*0.005), 1)
	}

	// Auto-scale thickness if not specified (0 means auto, -1 means filled)
	if thickness == 0 {
		thickness = maxInt(radius-1, 1)
	}

	// Draw circle
	gocv.Circle(frame, position, radius, color.ToRGBA(), thickness)
}

// Text draws text on the frame with optional shadow for legibility.
// Auto-scales size and thickness based on frame dimensions if not specified.
//
func (d *Drawer) Text(
	frame *gocv.Mat,
	text string,
	position image.Point,
	size float64,
	color Color,
	thickness int,
	shadow bool,
	shadowColor Color,
	shadowOffset int,
) {
	// Auto-scale size if not specified (0 means auto)
	if size == 0 {
		maxDim := float64(max(frame.Rows(), frame.Cols()))
		size = math.Min(math.Max(maxDim/4000.0, 0.5), 1.5)
	}

	// Auto-scale thickness if not specified (0 means auto)
	if thickness == 0 {
		// Match Python's banker's rounding: round(0.5) = 0, round(1.5) = 2
		rounded := math.RoundToEven(size)
		thickness = int(rounded + 1)
	}

	// Adjust position based on thickness
	// Python: anchor = (position[0] + thickness // 2, position[1] - thickness // 2)
	anchor := image.Point{
		X: position.X + thickness/2,
		Y: position.Y - thickness/2,
	}

	// Draw shadow first (if enabled)
	if shadow {
		shadowPos := image.Point{
			X: anchor.X + shadowOffset,
			Y: anchor.Y + shadowOffset,
		}
		gocv.PutTextWithParams(
			frame,
			text,
			shadowPos,
			gocv.FontHersheySimplex,
			size,
			shadowColor.ToRGBA(),
			thickness,
			gocv.LineAA, // Anti-aliased text (matches Python cv2.LINE_AA)
			false,       // bottomLeftOrigin
		)
	}

	// Draw foreground text
	gocv.PutTextWithParams(
		frame,
		text,
		anchor,
		gocv.FontHersheySimplex,
		size,
		color.ToRGBA(),
		thickness,
		gocv.LineAA, // Anti-aliased text (matches Python cv2.LINE_AA)
		false,       // bottomLeftOrigin
	)
}

// Rectangle draws a rectangle on the frame.
//
func (d *Drawer) Rectangle(frame *gocv.Mat, pt1 image.Point, pt2 image.Point, color Color, thickness int) {
	if thickness == 0 {
		thickness = 1
	}

	rect := image.Rectangle{Min: pt1, Max: pt2}
	gocv.Rectangle(frame, rect, color.ToRGBA(), thickness)
}

// Line draws a line segment on the frame.
//
func (d *Drawer) Line(frame *gocv.Mat, start image.Point, end image.Point, color Color, thickness int) {
	if thickness == 0 {
		thickness = 1
	}

	gocv.Line(frame, start, end, color.ToRGBA(), thickness)
}

// Cross draws a cross marker (+ shape) on the frame.
//
func (d *Drawer) Cross(frame *gocv.Mat, center image.Point, radius int, color Color, thickness int) {
	// Vertical line
	start1 := image.Point{X: center.X, Y: center.Y - radius}
	end1 := image.Point{X: center.X, Y: center.Y + radius}
	d.Line(frame, start1, end1, color, thickness)

	// Horizontal line
	start2 := image.Point{X: center.X - radius, Y: center.Y}
	end2 := image.Point{X: center.X + radius, Y: center.Y}
	d.Line(frame, start2, end2, color, thickness)
}

// AlphaBlend performs weighted blending of two frames.
// output = alpha*frame1 + beta*frame2 + gamma
// If beta is < 0, it defaults to 1-alpha.
//
func (d *Drawer) AlphaBlend(frame1 *gocv.Mat, frame2 *gocv.Mat, alpha float64, beta float64, gamma float64) gocv.Mat {
	// Auto-calculate beta if not specified
	if beta < 0 {
		beta = 1.0 - alpha
	}

	result := gocv.NewMat()
	gocv.AddWeighted(*frame1, alpha, *frame2, beta, gamma, &result)
	return result
}

// =============================================================================
// Drawable - Unified interface for Detections and TrackedObjects
// =============================================================================

// Drawable provides a unified interface for drawing Detections and TrackedObjects.
// Extracts relevant fields for rendering regardless of source type.
//
type Drawable struct {
	Points     *mat.Dense
	ID         *int
	Label      *string
	Scores     []float64
	LivePoints []bool
}

// Detection interface (minimal required fields for drawing)
type DetectionLike interface {
	GetPoints() *mat.Dense
	GetLabel() *string
	GetScores() []float64
}

// TrackedObject interface (minimal required fields for drawing)
type TrackedObjectLike interface {
	GetEstimate(absolute bool) (*mat.Dense, error)
	GetID() *int
	GetLabel() *string
	GetLivePoints() []bool
}

// NewDrawableFromDetection creates a Drawable from a Detection-like object.
func NewDrawableFromDetection(det DetectionLike) (*Drawable, error) {
	points := det.GetPoints()
	if points == nil {
		return nil, fmt.Errorf("detection has nil points")
	}

	// All detection points are considered alive
	rows, _ := points.Dims()
	livePoints := make([]bool, rows)
	for i := range livePoints {
		livePoints[i] = true
	}

	return &Drawable{
		Points:     points,
		ID:         nil, // Detections don't have IDs
		Label:      det.GetLabel(),
		Scores:     det.GetScores(),
		LivePoints: livePoints,
	}, nil
}

// NewDrawableFromTrackedObject creates a Drawable from a TrackedObject-like object.
func NewDrawableFromTrackedObject(obj TrackedObjectLike) (*Drawable, error) {
	// Get estimate in relative coordinates (absolute=false)
	estimate, err := obj.GetEstimate(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get estimate: %w", err)
	}

	return &Drawable{
		Points:     estimate,
		ID:         obj.GetID(),
		Label:      obj.GetLabel(),
		Scores:     nil, // TrackedObjects don't track per-point scores
		LivePoints: obj.GetLivePoints(),
	}, nil
}

// NewDrawable creates a Drawable with explicit fields.
// Used when neither Detection nor TrackedObject is available.
func NewDrawable(points *mat.Dense, id *int, label *string, scores []float64, livePoints []bool) (*Drawable, error) {
	if points == nil {
		return nil, fmt.Errorf("points cannot be nil")
	}

	// If livePoints not specified, assume all points are alive
	if livePoints == nil {
		rows, _ := points.Dims()
		livePoints = make([]bool, rows)
		for i := range livePoints {
			livePoints[i] = true
		}
	}

	return &Drawable{
		Points:     points,
		ID:         id,
		Label:      label,
		Scores:     scores,
		LivePoints: livePoints,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// maxInt is an alias for max to match Python naming.
func maxInt(a, b int) int {
	return max(a, b)
}
