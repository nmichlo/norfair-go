package drawing

import (
	"fmt"
	"image"
	"math"
	"math/rand"

	colorpkg "github.com/nmichlo/norfair-go/pkg/color"
	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"
)

// DrawPoints draws the points included in a list of Detections or TrackedObjects.
func DrawPoints(
	frame *gocv.Mat,
	drawables []interface{}, // []Detection or []TrackedObject
	radius *int,
	thickness *int,
	color interface{}, // ColorLike: string, Color, or strategy
	drawLabels bool,
	textSize *float64,
	drawIDs bool,
	drawPoints bool,
	textThickness *int,
	textColor interface{},
	hideDeadPoints bool,
	drawScores bool,
) *gocv.Mat {
	// Early return if no drawables
	if drawables == nil || len(drawables) == 0 {
		return frame
	}

	// Parse text color if provided
	var parsedTextColor *Color
	if textColor != nil {
		c := resolveDirectColor(textColor)
		parsedTextColor = &c
	}

	// Set defaults
	if color == nil {
		color = "by_id"
	}
	if thickness == nil {
		t := -1
		thickness = &t
	}
	if radius == nil {
		maxDim := max(frame.Rows(), frame.Cols())
		r := int(math.Round(math.Max(float64(maxDim)*0.002, 1)))
		radius = &r
	}

	drawer := NewDrawer()
	palette := NewPalette(nil) // default tab10

	// Process each drawable
	for _, obj := range drawables {
		// Convert to Drawable if needed
		var d *Drawable
		switch o := obj.(type) {
		case *Drawable:
			d = o
		default:
			// Try to create from Detection or TrackedObject
			var err error
			d, err = createDrawableFromInterface(obj)
			if err != nil {
				continue // Skip invalid objects
			}
		}

		// Skip if all points are dead and hideDeadPoints is true
		if hideDeadPoints && !hasLivePoints(d.LivePoints) {
			continue
		}

		// Determine object color
		objColor := resolveColor(color, d, palette)

		// Determine text color
		var objTextColor Color
		if parsedTextColor != nil {
			objTextColor = *parsedTextColor
		} else {
			objTextColor = objColor
		}

		// Draw points (circles)
		if drawPoints {
			rows, _ := d.Points.Dims()
			for i := 0; i < rows; i++ {
				live := d.LivePoints[i]
				if live || !hideDeadPoints {
					x := int(d.Points.At(i, 0))
					y := int(d.Points.At(i, 1))
					point := image.Point{X: x, Y: y}

					drawer.Circle(frame, point, *radius, *thickness, objColor)
				}
			}
		}

		// Draw text
		if drawLabels || drawIDs || drawScores {
			// Calculate position: mean of live points minus radius
			livePoints := filterLivePoints(d.Points, d.LivePoints)
			if livePoints != nil && livePoints.RawMatrix().Rows > 0 {
				centroidX, centroidY := Centroid(livePoints)
				position := image.Point{
					X: centroidX - *radius,
					Y: centroidY - *radius,
				}

				// Build text
				text := BuildText(d, drawLabels, drawIDs, drawScores)

				// Determine text thickness
				var finalTextThickness int
				if textThickness != nil {
					finalTextThickness = *textThickness
				} else {
					finalTextThickness = 0 // Auto-scale
				}

				// Determine text size
				var finalTextSize float64
				if textSize != nil {
					finalTextSize = *textSize
				} else {
					finalTextSize = 0 // Auto-scale
				}

				drawer.Text(
					frame,
					text,
					position,
					finalTextSize,
					objTextColor,
					finalTextThickness,
					true,           // shadow
					colorpkg.Black, // shadowColor
					2,              // shadowOffset
				)
			}
		}
	}

	return frame
}

// resolveColor determines the color based on the strategy or direct value.
func resolveColor(colorStrategy interface{}, drawable *Drawable, palette *Palette) Color {
	switch strategy := colorStrategy.(type) {
	case string:
		switch strategy {
		case "by_id":
			if drawable.ID != nil {
				return palette.ChooseColor(*drawable.ID)
			}
			return palette.ChooseColor(nil) // Use default color
		case "by_label":
			if drawable.Label != nil {
				return palette.ChooseColor(*drawable.Label)
			}
			return palette.ChooseColor(nil) // Use default color
		case "random":
			// Random color each time (using random float)
			return palette.ChooseColor(rand.Float64())
		default:
			// Try to parse as hex string or color name
			c, err := ParseColorName(strategy)
			if err != nil {
				// Try hex
				c2, err2 := HexToBGR(strategy)
				if err2 != nil {
					// Default to white if parsing fails
					return colorpkg.White
				}
				return c2
			}
			return c
		}
	case Color:
		return strategy
	default:
		// Unknown type, return white
		return colorpkg.White
	}
}

// resolveDirectColor resolves a direct color value (not a strategy).
func resolveDirectColor(colorValue interface{}) Color {
	switch c := colorValue.(type) {
	case string:
		// Try color name first
		col, err := ParseColorName(c)
		if err != nil {
			// Try hex
			col2, err2 := HexToBGR(c)
			if err2 != nil {
				return colorpkg.White // Default
			}
			return col2
		}
		return col
	case Color:
		return c
	default:
		return colorpkg.White
	}
}

// hasLivePoints checks if any point is live.
func hasLivePoints(livePoints []bool) bool {
	for _, live := range livePoints {
		if live {
			return true
		}
	}
	return false
}

// filterLivePoints returns a new matrix containing only the live points.
func filterLivePoints(points *mat.Dense, livePoints []bool) *mat.Dense {
	rows, cols := points.Dims()
	if rows == 0 {
		return nil
	}

	// Count live points
	liveCount := 0
	for _, live := range livePoints {
		if live {
			liveCount++
		}
	}

	if liveCount == 0 {
		return nil
	}

	// Extract live points
	liveData := make([]float64, liveCount*cols)
	idx := 0
	for i := 0; i < rows; i++ {
		if livePoints[i] {
			for j := 0; j < cols; j++ {
				liveData[idx*cols+j] = points.At(i, j)
			}
			idx++
		}
	}

	return mat.NewDense(liveCount, cols, liveData)
}

// createDrawableFromInterface attempts to create a Drawable from Detection or TrackedObject.
func createDrawableFromInterface(obj interface{}) (*Drawable, error) {
	// Try Detection interface
	if det, ok := obj.(DetectionLike); ok {
		return NewDrawableFromDetection(det)
	}

	// Try TrackedObject interface
	if tracked, ok := obj.(TrackedObjectLike); ok {
		return NewDrawableFromTrackedObject(tracked)
	}

	// Unknown type
	return nil, fmt.Errorf("object is not a Detection or TrackedObject")
}
