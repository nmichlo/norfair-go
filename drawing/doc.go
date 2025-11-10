/*
Package drawing provides visualization for tracked objects.

Includes functions to draw bounding boxes, points, paths, and overlays on video
frames using gocv.

# Basic Usage

	import "github.com/nmichlo/norfair-go/drawing"

	// Draw tracked points
	drawing.DrawPoints(frame, trackedObjects,
		drawing.WithColorStrategy("by_id"),
		drawing.WithDrawIDs(true),
	)

	// Draw bounding boxes
	drawing.DrawBoxes(frame, trackedObjects,
		drawing.WithColorStrategy("by_id"),
		drawing.WithLineWidth(3),
	)

	// Draw motion trails
	paths := drawing.NewPaths(50)
	for _, obj := range trackedObjects {
		paths.UpdatePath(obj)
	}
	drawing.DrawPaths(frame, paths)

# Color Strategies

  - "by_id": Unique color per object ID
  - "by_label": Same color per label
  - "by_score": Gradient based on confidence

# Components

Drawer: Primitive drawing operations
Color: RGBA with conversion utilities
Path: Movement history tracking
*/
package drawing
