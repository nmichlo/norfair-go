package norfairgo

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

// DetectionConfig contains optional configuration for a Detection.
// All fields are optional and can be nil/zero.
type DetectionConfig struct {
	// Scores are optional per-point confidence scores
	// Used for partial measurements in Kalman filter update
	Scores []float64

	// Data is custom user data storage
	Data interface{}

	// Label is the class label for multi-class tracking
	// Objects with different labels are never matched
	Label *string

	// Embedding is the ReID embedding for re-identification
	Embedding []float64
}

// Detection represents a detected object in a frame.
type Detection struct {
	// Points are the original detection points in relative coordinates (camera frame)
	Points *mat.Dense

	// AbsolutePoints are the transformed points in absolute coordinates (world frame)
	// Initially a copy of Points, updated by UpdateCoordinateTransformation
	AbsolutePoints *mat.Dense

	// Scores are optional per-point confidence scores (can be nil)
	// Used for partial measurements in Kalman filter update
	Scores []float64

	// Data is custom user data storage (can be nil)
	Data interface{}

	// Label is the class label for multi-class tracking (can be nil)
	// Objects with different labels are never matched
	Label *string

	// Embedding is the ReID embedding for re-identification (can be nil)
	Embedding []float64

	// Age is the age of this detection when added to past_detections
	// Set by TrackedObject when storing past detections
	Age int
}

// StringPtr returns a pointer to a string. Helper for DetectionConfig.Label.
//
// Example:
//
//	det, err := NewDetection(points, &DetectionConfig{
//	    Label: StringPtr("person"),
//	})
func StringPtr(s string) *string {
	return &s
}

// NewDetection creates a new Detection from points and optional configuration.
//
// Example - Simple (no config):
//
//	det, err := norfairgo.NewDetection(points, nil)
//
// Example - With label:
//
//	det, err := norfairgo.NewDetection(points, &norfairgo.DetectionConfig{
//	    Label: norfairgo.StringPtr("person"),
//	})
//
// Example - Full config:
//
//	det, err := norfairgo.NewDetection(points, &norfairgo.DetectionConfig{
//	    Scores:    []float64{0.9, 0.8},
//	    Label:     norfairgo.StringPtr("person"),
//	    Embedding: featureVector,
//	})
//
// Parameters:
//   - points: Detection points, shape (n_points, n_dims) where n_dims is 2 or 3
//   - config: Optional configuration (can be nil)
//
// Returns error if points have invalid shape.
func NewDetection(points *mat.Dense, config *DetectionConfig) (*Detection, error) {
	// Validate and normalize points
	validatedPoints, err := ValidatePoints(points)
	if err != nil {
		return nil, fmt.Errorf("invalid detection points: %w", err)
	}

	// Create copy for absolute points
	rows, cols := validatedPoints.Dims()
	absolutePoints := mat.NewDense(rows, cols, nil)
	absolutePoints.Copy(validatedPoints)

	// Extract config fields (or use defaults)
	var scores []float64
	var data interface{}
	var label *string
	var embedding []float64

	if config != nil {
		scores = config.Scores
		data = config.Data
		label = config.Label
		embedding = config.Embedding
	}

	return &Detection{
		Points:         validatedPoints,
		AbsolutePoints: absolutePoints,
		Scores:         scores,
		Data:           data,
		Label:          label,
		Embedding:      embedding,
		Age:            0,
	}, nil
}

// UpdateCoordinateTransformation transforms detection points to absolute coordinates.
// This is used for camera motion compensation.
//
// If coordTransform is nil, this is a no-op.
func (d *Detection) UpdateCoordinateTransformation(coordTransform CoordinateTransformation) {
	if coordTransform != nil {
		d.AbsolutePoints = coordTransform.RelToAbs(d.AbsolutePoints)
	}
}

// GetPoints returns the detection points in relative coordinates.
// Required by norfairgodraw.DetectionLike interface.
func (d *Detection) GetPoints() *mat.Dense {
	return d.Points
}

// GetLabel returns the detection label.
// Required by norfairgodraw.DetectionLike interface.
func (d *Detection) GetLabel() *string {
	return d.Label
}

// GetScores returns the detection scores.
// Required by norfairgodraw.DetectionLike interface.
func (d *Detection) GetScores() []float64 {
	return d.Scores
}
