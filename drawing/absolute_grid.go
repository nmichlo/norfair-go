package drawing

import (
	"fmt"
	"image"
	"math"
	"sync"

	"gocv.io/x/gocv"
	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go"
)

// gridCacheKey is the key for caching grid generation results.
type gridCacheKey struct {
	size  int
	w     int
	h     int
	polar bool
}

// gridCache stores pre-computed grids (implements LRU-like caching).
var (
	gridCache      = make(map[gridCacheKey]*mat.Dense)
	gridCacheMutex sync.RWMutex
	gridCacheMax   = 4 // Match Python's @lru_cache(maxsize=4)
)

// getGrid constructs a grid of points in absolute coordinates.
//
// Points are chosen on a semi-sphere of radius 1 centered around (0, 0).
// Results are cached since the grid in absolute coordinates doesn't change.
func getGrid(size, w, h int, polar bool) *mat.Dense {
	key := gridCacheKey{size: size, w: w, h: h, polar: polar}

	// Check cache (read lock)
	gridCacheMutex.RLock()
	if cached, ok := gridCache[key]; ok {
		gridCacheMutex.RUnlock()
		return cached
	}
	gridCacheMutex.RUnlock()

	// Not in cache, compute (write lock)
	gridCacheMutex.Lock()
	defer gridCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if cached, ok := gridCache[key]; ok {
		return cached
	}

	// Generate grid
	grid := computeGrid(size, w, h, polar)

	// Add to cache (simple eviction: clear if at max)
	if len(gridCache) >= gridCacheMax {
		// Simple eviction: clear entire cache (Python LRU is more sophisticated)
		gridCache = make(map[gridCacheKey]*mat.Dense)
	}
	gridCache[key] = grid

	return grid
}

// computeGrid generates the spherical grid projection.
//
// Python implementation:
//
//	step = np.pi / size
//	start = -np.pi / 2 + step / 2
//	end = np.pi / 2
//	theta, fi = np.mgrid[start:end:step, start:end:step]
//
// Then projects onto plane using either polar or equator mode.
func computeGrid(size, w, h int, polar bool) *mat.Dense {
	// Generate angular grid: theta, phi ∈ (-π/2, π/2)
	step := math.Pi / float64(size)
	start := -math.Pi/2 + step/2
	end := math.Pi / 2

	// Count grid points
	numTheta := int(math.Ceil((end - start) / step))
	numPhi := numTheta // Square grid
	numPoints := numTheta * numPhi

	// Pre-allocate result matrix
	points := mat.NewDense(numPoints, 2, nil)

	idx := 0
	for i := 0; i < numTheta; i++ {
		theta := start + float64(i)*step
		for j := 0; j < numPhi; j++ {
			phi := start + float64(j)*step

			var x, y float64

			if polar {
				// Polar mode: view from pole
				// Points on sphere: [sin(θ)*cos(φ), sin(θ)*sin(φ), cos(θ)]
				// Project onto z=1 plane: [tan(θ)*cos(φ), tan(θ)*sin(φ), 1]
				tanTheta := math.Tan(theta)
				x = tanTheta * math.Cos(phi)
				y = tanTheta * math.Sin(phi)
			} else {
				// Equator mode: view from equator (default)
				// X = tan(φ)
				// Y = tan(θ) / cos(φ)
				x = math.Tan(phi)
				y = math.Tan(theta) / math.Cos(phi)
			}

			// Scale and center the points
			// Python: points * max(h, w) + np.array([w // 2, h // 2])
			maxDim := float64(maxInt(h, w))
			x = x*maxDim + float64(w/2)
			y = y*maxDim + float64(h/2)

			points.Set(idx, 0, x)
			points.Set(idx, 1, y)
			idx++
		}
	}

	return points
}

// DrawAbsoluteGrid draws a grid of points in absolute coordinates.
//
// Useful for debugging camera motion.
//
// The points are drawn as if the camera were in the center of a sphere and points
// are drawn in the intersection of latitude and longitude lines over the surface
// of the sphere.
//
// Parameters:
//   - frame: The OpenCV frame to draw on
//   - coordTransform: The coordinate transformation from MotionEstimator (can be nil)
//   - gridSize: How many points to draw (default 20)
//   - radius: Size of each point (default 2)
//   - thickness: Thickness of each point (default 1)
//   - color: Color of the points (default black)
//   - polar: If true, points drawn as if viewing pole; if false, viewing equator (default false)
func DrawAbsoluteGrid(
	frame *gocv.Mat,
	coordTransform norfairgo.CoordinateTransformation,
	gridSize int,
	radius int,
	thickness int,
	color *Color,
	polar bool,
) {
	if frame == nil {
		return
	}

	h := frame.Rows()
	w := frame.Cols()

	// Set defaults
	if gridSize <= 0 {
		gridSize = 20
	}
	if radius <= 0 {
		radius = 2
	}
	if thickness <= 0 {
		thickness = 1
	}
	if color == nil {
		color = &Color{B: 0, G: 0, R: 0} // Black
	}

	// Get absolute points grid (cached)
	points := getGrid(gridSize, w, h, polar)

	// Transform points to relative coordinates
	var pointsTransformed *mat.Dense
	if coordTransform == nil {
		pointsTransformed = points
	} else {
		pointsTransformed = coordTransform.AbsToRel(points)
	}

	// Filter points that are visible in frame
	// Python: (points_transformed <= np.array([w, h])).all(axis=1) & (points_transformed >= 0).all(axis=1)
	drawer := NewDrawer()

	rows, _ := pointsTransformed.Dims()
	for i := 0; i < rows; i++ {
		x := pointsTransformed.At(i, 0)
		y := pointsTransformed.At(i, 1)

		// Check if point is visible
		if x >= 0 && x <= float64(w) && y >= 0 && y <= float64(h) {
			// Draw cross at point location
			drawer.Cross(
				frame,
				image.Point{X: int(x), Y: int(y)},
				radius,
				*color,
				thickness,
			)
		}
	}
}

// DrawAbsoluteGridWithDefaults is a convenience function that uses default parameters.
func DrawAbsoluteGridWithDefaults(frame *gocv.Mat, coordTransform norfairgo.CoordinateTransformation) {
	DrawAbsoluteGrid(frame, coordTransform, 20, 2, 1, &Color{B: 0, G: 0, R: 0}, false)
}

// ClearGridCache clears the grid generation cache (useful for testing).
func ClearGridCache() {
	gridCacheMutex.Lock()
	defer gridCacheMutex.Unlock()
	gridCache = make(map[gridCacheKey]*mat.Dense)
}

// GridCacheSize returns the current size of the grid cache (useful for testing).
func GridCacheSize() int {
	gridCacheMutex.RLock()
	defer gridCacheMutex.RUnlock()
	return len(gridCache)
}

// GridCacheKey returns a cache key for testing purposes.
func GridCacheKey(size, w, h int, polar bool) string {
	return fmt.Sprintf("size=%d,w=%d,h=%d,polar=%t", size, w, h, polar)
}
