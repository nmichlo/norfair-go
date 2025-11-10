package drawing

import (
	"image"

	colorpkg "github.com/nmichlo/norfair-go/color"
	"gocv.io/x/gocv"
)

// DrawBoxes draws bounding boxes for Detections or TrackedObjects.
//
func DrawBoxes(
	frame *gocv.Mat,
	drawables []interface{},
	color interface{},
	thickness *int,
	drawLabels bool,
	textSize *float64,
	drawIDs bool,
	textColor interface{},
	textThickness *int,
	drawBox bool,
	drawScores bool,
) *gocv.Mat {
	// Set defaults
	if color == nil {
		color = "by_id"
	}
	if thickness == nil {
		maxDim := max(frame.Rows(), frame.Cols())
		t := int(maxDim / 500)
		thickness = &t
	}

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

		// Determine object color
		objColor := resolveColor(color, d, palette)

		// Convert points to int
		rows, cols := d.Points.Dims()
		if rows != 2 || cols != 2 {
			continue // Skip if not a bounding box (must have 2 points)
		}

		x0 := int(d.Points.At(0, 0))
		y0 := int(d.Points.At(0, 1))
		x1 := int(d.Points.At(1, 0))
		y1 := int(d.Points.At(1, 1))

		// Draw box
		if drawBox {
			pt1 := image.Point{X: x0, Y: y0}
			pt2 := image.Point{X: x1, Y: y1}
			drawer.Rectangle(frame, pt1, pt2, objColor, *thickness)
		}

		// Build text
		text := BuildText(d, drawLabels, drawIDs, drawScores)

		// Draw text if not empty
		if text != "" {
			// Determine text color
			var objTextColor Color
			if parsedTextColor != nil {
				objTextColor = *parsedTextColor
			} else {
				objTextColor = objColor
			}

			// Calculate text anchor:
			// top-left of bbox, compensating for box thickness
			textAnchor := image.Point{
				X: x0 - *thickness/2,
				Y: y0 - *thickness/2 - 1,
			}

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
				textAnchor,
				finalTextSize,
				objTextColor,
				finalTextThickness,
				true,           // shadow
				colorpkg.Black, // shadowColor
				2,              // shadowOffset
			)
		}
	}

	return frame
}
