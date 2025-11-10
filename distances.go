package norfairgo

import (
	"fmt"
	"log"
	"math"

	"github.com/nmichlo/norfair-go/internal/scipy"
	"gonum.org/v1/gonum/mat"
)

// =============================================================================
// Distance Interface
// =============================================================================

// Distance is the interface for all distance implementations
type Distance interface {
	GetDistances(objects []*TrackedObject, candidates interface{}) *mat.Dense
}

// =============================================================================
// ScalarDistance - Wraps pointwise distance functions
// =============================================================================

// ScalarDistance wraps a function that computes distance between one detection and one tracked object
type ScalarDistance struct {
	distanceFunction func(*Detection, *TrackedObject) float64
}

// NewScalarDistance creates a new ScalarDistance
func NewScalarDistance(distanceFunction func(*Detection, *TrackedObject) float64) *ScalarDistance {
	return &ScalarDistance{
		distanceFunction: distanceFunction,
	}
}

// GetDistances computes the distance matrix using scalar distance function
func (sd *ScalarDistance) GetDistances(objects []*TrackedObject, candidates interface{}) *mat.Dense {
	candList := convertCandidatesToList(candidates)
	distanceMatrix := createInfinityMatrix(len(candList), len(objects))

	if len(candList) == 0 || len(objects) == 0 {
		return distanceMatrix
	}

	for c := 0; c < len(candList); c++ {
		for o := 0; o < len(objects); o++ {
			if dist, ok := sd.computePairDistance(candList[c], objects[o]); ok {
				distanceMatrix.Set(c, o, dist)
			}
		}
	}

	return distanceMatrix
}

func convertCandidatesToList(candidates interface{}) []interface{} {
	switch v := candidates.(type) {
	case []Detection:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = &v[i]
		}
		return result
	case []*Detection:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	case []TrackedObject:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = &v[i]
		}
		return result
	case []*TrackedObject:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	default:
		panic(fmt.Sprintf("unsupported candidates type: %T", candidates))
	}
}

func createInfinityMatrix(rows, cols int) *mat.Dense {
	matrix := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			matrix.Set(i, j, math.Inf(1))
		}
	}
	return matrix
}

func (sd *ScalarDistance) computePairDistance(candidate interface{}, obj *TrackedObject) (float64, bool) {
	switch cand := candidate.(type) {
	case *Detection:
		if labelsMatch(cand.Label, obj.Label) {
			return sd.distanceFunction(cand, obj), true
		}
	case *TrackedObject:
		// Can't compute scalar distance between two TrackedObjects
		return 0, false
	}
	return 0, false
}

func labelsMatch(label1, label2 *string) bool {
	if label1 != nil && label2 != nil {
		return *label1 == *label2
	}
	if label1 != nil || label2 != nil {
		log.Printf("Warning: comparing objects with mismatched label presence")
		return false
	}
	return true
}

// =============================================================================
// Built-in Distance Functions (Scalar)
// =============================================================================

// Frobenius computes the Frobenius norm distance
func Frobenius(detection *Detection, trackedObject *TrackedObject) float64 {
	// Compute difference matrix
	rows, cols := detection.Points.Dims()
	diff := mat.NewDense(rows, cols, nil)
	diff.Sub(detection.Points, trackedObject.Estimate)

	// Compute Frobenius norm
	return mat.Norm(diff, 2)
}

// MeanEuclidean computes the mean Euclidean distance between corresponding points
func MeanEuclidean(detection *Detection, trackedObject *TrackedObject) float64 {
	rows, _ := detection.Points.Dims()

	var sum float64
	for i := 0; i < rows; i++ {
		// Get i-th point from detection and estimate
		detPoint := detection.Points.RawRowView(i)
		estPoint := trackedObject.Estimate.RawRowView(i)

		// Compute Euclidean distance for this point
		var distSq float64
		for j := range detPoint {
			diff := detPoint[j] - estPoint[j]
			distSq += diff * diff
		}
		sum += math.Sqrt(distSq)
	}

	return sum / float64(rows)
}

// MeanManhattan computes the mean Manhattan distance between corresponding points
func MeanManhattan(detection *Detection, trackedObject *TrackedObject) float64 {
	rows, _ := detection.Points.Dims()

	var sum float64
	for i := 0; i < rows; i++ {
		// Get i-th point from detection and estimate
		detPoint := detection.Points.RawRowView(i)
		estPoint := trackedObject.Estimate.RawRowView(i)

		// Compute Manhattan distance for this point
		var dist float64
		for j := range detPoint {
			dist += math.Abs(detPoint[j] - estPoint[j])
		}
		sum += dist
	}

	return sum / float64(rows)
}

// =============================================================================
// VectorizedDistance - Batch distance computation
// =============================================================================

// VectorizedDistance wraps a function that computes distances for all pairs at once
type VectorizedDistance struct {
	distanceFunction func(candidates, objects *mat.Dense) *mat.Dense
}

// NewVectorizedDistance creates a new VectorizedDistance
func NewVectorizedDistance(distanceFunction func(candidates, objects *mat.Dense) *mat.Dense) *VectorizedDistance {
	return &VectorizedDistance{
		distanceFunction: distanceFunction,
	}
}

// GetDistances computes the distance matrix using vectorized distance function
func (vd *VectorizedDistance) GetDistances(objects []*TrackedObject, candidates interface{}) *mat.Dense {
	candList := convertCandidatesToList(candidates)
	distanceMatrix := createInfinityMatrix(len(candList), len(objects))

	if len(candList) == 0 || len(objects) == 0 {
		return distanceMatrix
	}

	objectLabels := extractObjectLabels(objects)
	candidateLabels := extractCandidateLabels(candList)
	uniqueLabels := findIntersection(unique(objectLabels), unique(candidateLabels))

	for _, label := range uniqueLabels {
		vd.processLabelGroup(label, objects, candList, objectLabels, candidateLabels, distanceMatrix)
	}

	return distanceMatrix
}

func extractObjectLabels(objects []*TrackedObject) []string {
	labels := make([]string, len(objects))
	for i := range objects {
		if objects[i].Label != nil {
			labels[i] = *objects[i].Label
		} else {
			labels[i] = "None"
		}
	}
	return labels
}

func extractCandidateLabels(candList []interface{}) []string {
	labels := make([]string, len(candList))
	for i, cand := range candList {
		switch c := cand.(type) {
		case *Detection:
			if c.Label != nil {
				labels[i] = *c.Label
			} else {
				labels[i] = "None"
			}
		case *TrackedObject:
			if c.Label != nil {
				labels[i] = *c.Label
			} else {
				labels[i] = "None"
			}
		}
	}
	return labels
}

func (vd *VectorizedDistance) processLabelGroup(
	label string,
	objects []*TrackedObject,
	candList []interface{},
	objectLabels, candidateLabels []string,
	distanceMatrix *mat.Dense,
) {
	objIndices := findLabelIndices(objectLabels, label)
	candIndices := findLabelIndices(candidateLabels, label)

	if len(objIndices) == 0 || len(candIndices) == 0 {
		return
	}

	stackedObjects := stackObjectEstimates(objects, objIndices)
	stackedCandidates := stackCandidateData(candList, candIndices, stackedObjects.RawMatrix().Cols)

	distances := vd.distanceFunction(stackedCandidates, stackedObjects)
	assignDistancesToMatrix(distances, candIndices, objIndices, distanceMatrix)
}

func stackObjectEstimates(objects []*TrackedObject, indices []int) *mat.Dense {
	if len(indices) == 0 {
		return nil
	}

	firstEst := objects[indices[0]].Estimate
	rows, cols := firstEst.Dims()
	flattenedCols := rows * cols

	stacked := mat.NewDense(len(indices), flattenedCols, nil)
	for i, idx := range indices {
		flatData := flattenMatrix(objects[idx].Estimate)
		for j, val := range flatData {
			stacked.Set(i, j, val)
		}
	}
	return stacked
}

func stackCandidateData(candList []interface{}, indices []int, flattenedCols int) *mat.Dense {
	stacked := mat.NewDense(len(indices), flattenedCols, nil)
	for i, idx := range indices {
		var flatData []float64
		cand := candList[idx]

		if det, ok := cand.(*Detection); ok {
			flatData = flattenMatrix(det.Points)
		} else if obj, ok := cand.(*TrackedObject); ok {
			flatData = flattenMatrix(obj.Estimate)
		}

		for j, val := range flatData {
			stacked.Set(i, j, val)
		}
	}
	return stacked
}

func assignDistancesToMatrix(distances *mat.Dense, candIndices, objIndices []int, distanceMatrix *mat.Dense) {
	rows, cols := distances.Dims()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			distanceMatrix.Set(candIndices[i], objIndices[j], distances.At(i, j))
		}
	}
}

// =============================================================================
// ScipyDistance - cdist-style distance computation
// =============================================================================

// ScipyDistance implements scipy.spatial.distance.cdist functionality
type ScipyDistance struct {
	metric string
	VectorizedDistance
}

// NewScipyDistance creates a new ScipyDistance
func NewScipyDistance(metric string) *ScipyDistance {
	sd := &ScipyDistance{
		metric: metric,
	}
	// Set up the vectorized distance function using scipy.Cdist
	sd.VectorizedDistance.distanceFunction = func(candidates, objects *mat.Dense) *mat.Dense {
		return scipy.Cdist(candidates, objects, sd.metric)
	}
	return sd
}

// =============================================================================
// Built-in Distance Functions (Vectorized)
// =============================================================================

// IoU computes the IoU distance (1 - IoU) for bounding boxes
// Input format: [x_min, y_min, x_max, y_max]
func IoU(candidates, objects *mat.Dense) *mat.Dense {
	// Validate bboxes
	validateBboxes(candidates)
	validateBboxes(objects)

	candRows, _ := candidates.Dims()
	objRows, _ := objects.Dims()

	// Compute areas
	candAreas := boxesArea(candidates)
	objAreas := boxesArea(objects)

	// Compute pairwise IoU
	result := mat.NewDense(candRows, objRows, nil)

	for i := 0; i < candRows; i++ {
		for j := 0; j < objRows; j++ {
			// Get bboxes
			candBox := candidates.RawRowView(i)
			objBox := objects.RawRowView(j)

			// Intersection top-left
			xMin := math.Max(candBox[0], objBox[0])
			yMin := math.Max(candBox[1], objBox[1])

			// Intersection bottom-right
			xMax := math.Min(candBox[2], objBox[2])
			yMax := math.Min(candBox[3], objBox[3])

			// Intersection area
			width := math.Max(0, xMax-xMin)
			height := math.Max(0, yMax-yMin)
			intersection := width * height

			// Union area
			union := candAreas[i] + objAreas[j] - intersection

			// IoU distance = 1 - IoU
			iou := intersection / union
			result.Set(i, j, 1.0-iou)
		}
	}

	return result
}

// validateBboxes checks that bboxes have correct shape and warns on invalid bounds
func validateBboxes(bboxes *mat.Dense) {
	rows, cols := bboxes.Dims()
	if cols != 4 {
		panic(fmt.Sprintf("bboxes must have 4 columns, got %d", cols))
	}

	// Check for invalid bboxes (x_min >= x_max or y_min >= y_max)
	for i := 0; i < rows; i++ {
		row := bboxes.RawRowView(i)
		if row[0] >= row[2] || row[1] >= row[3] {
			log.Printf("Warning: bbox at row %d has invalid bounds: [%.2f, %.2f, %.2f, %.2f]",
				i, row[0], row[1], row[2], row[3])
		}
	}
}

// boxesArea computes the area of each bbox
func boxesArea(boxes *mat.Dense) []float64 {
	rows, _ := boxes.Dims()
	areas := make([]float64, rows)
	for i := 0; i < rows; i++ {
		row := boxes.RawRowView(i)
		areas[i] = (row[2] - row[0]) * (row[3] - row[1])
	}
	return areas
}

// =============================================================================
// Helper Functions
// =============================================================================

// Helper to convert mat.Dense to flattened slice
func flattenMatrix(m *mat.Dense) []float64 {
	rows, cols := m.Dims()
	result := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[i*cols+j] = m.At(i, j)
		}
	}
	return result
}

// unique returns unique strings from a slice
func unique(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, val := range s {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}
	return result
}

// findIntersection finds common elements between two slices
func findIntersection(a, b []string) []string {
	set := make(map[string]bool)
	for _, val := range a {
		set[val] = true
	}

	var result []string
	for _, val := range b {
		if set[val] {
			result = append(result, val)
		}
	}
	return result
}

// findLabelIndices finds all indices where the label matches
func findLabelIndices(labels []string, target string) []int {
	var indices []int
	for i, label := range labels {
		if label == target {
			indices = append(indices, i)
		}
	}
	return indices
}

// =============================================================================
// Factory Functions
// =============================================================================

// CreateKeypointsVotingDistance constructs a keypoint voting distance function
// configured with the thresholds.
//
// Count how many points in a detection match with a tracked_object.
// A match is considered when distance between the points is < keypointDistanceThreshold
// and the score of the last_detection of the tracked_object is > detectionThreshold.
// Notice that if multiple points are tracked, the ith point in detection can only match
// the ith point in the tracked object.
//
// Distance is 1 if no point matches and approximates 0 as more points are matched.
func CreateKeypointsVotingDistance(keypointDistanceThreshold, detectionThreshold float64) func(*Detection, *TrackedObject) float64 {
	return func(detection *Detection, trackedObject *TrackedObject) float64 {
		rows, _ := detection.Points.Dims()

		// Compute euclidean distances per row
		var matchNum int
		for i := 0; i < rows; i++ {
			detPoint := detection.Points.RawRowView(i)
			estPoint := trackedObject.Estimate.RawRowView(i)

			// Compute euclidean distance for this point
			var distSq float64
			for j := range detPoint {
				diff := detPoint[j] - estPoint[j]
				distSq += diff * diff
			}
			dist := math.Sqrt(distSq)

			// Check if this is a match
			if dist < keypointDistanceThreshold &&
				detection.Scores[i] > detectionThreshold &&
				trackedObject.LastDetection.Scores[i] > detectionThreshold {
				matchNum++
			}
		}

		return 1.0 / (1.0 + float64(matchNum))
	}
}

// CreateNormalizedMeanEuclideanDistance constructs a normalized mean euclidean distance
// function configured with the max height and width.
//
// The result distance is bound to [0, 1] where 1 indicates opposite corners of the image.
func CreateNormalizedMeanEuclideanDistance(height, width int) func(*Detection, *TrackedObject) float64 {
	fHeight := float64(height)
	fWidth := float64(width)

	return func(detection *Detection, trackedObject *TrackedObject) float64 {
		rows, _ := detection.Points.Dims()

		// Calculate normalized euclidean distances and average
		var sum float64
		for i := 0; i < rows; i++ {
			detPoint := detection.Points.RawRowView(i)
			estPoint := trackedObject.Estimate.RawRowView(i)

			// Normalize by width (x-axis, index 0) and height (y-axis, index 1)
			var distSq float64
			for j := range detPoint {
				var diff float64
				if j == 0 {
					// x-coordinate - normalize by width
					diff = (detPoint[j] - estPoint[j]) / fWidth
				} else if j == 1 {
					// y-coordinate - normalize by height
					diff = (detPoint[j] - estPoint[j]) / fHeight
				} else {
					// Other dimensions - no normalization
					diff = detPoint[j] - estPoint[j]
				}
				distSq += diff * diff
			}
			sum += math.Sqrt(distSq)
		}

		return sum / float64(rows)
	}
}

// =============================================================================
// Distance Registry
// =============================================================================

// Scalar distance function registry
var scalarDistanceFunctions = map[string]func(*Detection, *TrackedObject) float64{
	"frobenius":      Frobenius,
	"mean_manhattan": MeanManhattan,
	"mean_euclidean": MeanEuclidean,
}

// Vectorized distance function registry
var vectorizedDistanceFunctions = map[string]func(*mat.Dense, *mat.Dense) *mat.Dense{
	"iou":     IoU,
	"iou_opt": IoU, // deprecated, same as iou
}

// List of supported scipy distance metrics
var scipyDistanceMetrics = []string{
	"braycurtis", "canberra", "chebyshev", "cityblock", "correlation",
	"cosine", "dice", "euclidean", "hamming", "jaccard", "jensenshannon",
	"kulczynski1", "mahalanobis", "matching", "minkowski", "rogerstanimoto",
	"russellrao", "seuclidean", "sokalmichener", "sokalsneath", "sqeuclidean",
	"yule",
}

// GetDistanceByName selects a distance by name.
//
// Returns the corresponding Distance implementation for the given name.
// Supports scalar distances (frobenius, mean_euclidean, mean_manhattan),
// vectorized distances (iou), and scipy metrics (euclidean, manhattan, etc.).
func GetDistanceByName(name string) Distance {
	// Check scalar distances
	if fn, ok := scalarDistanceFunctions[name]; ok {
		log.Printf("Warning: You are using a scalar distance function. If you want to speed up the tracking process please consider using a vectorized distance function.")
		return NewScalarDistance(fn)
	}

	// Check vectorized distances
	if fn, ok := vectorizedDistanceFunctions[name]; ok {
		if name == "iou_opt" {
			log.Printf("Warning: iou_opt is deprecated, use iou instead")
		}
		return NewVectorizedDistance(fn)
	}

	// Check scipy distances
	for _, metric := range scipyDistanceMetrics {
		if name == metric {
			return NewScipyDistance(name)
		}
	}

	// Not found
	panic(fmt.Sprintf("Invalid distance '%s', expecting one of the supported distance names", name))
}

// DistanceByName is a convenience alias for GetDistanceByName.
// Panics if the distance name is invalid.
//
// Example:
//
//	config := &TrackerConfig{
//	    DistanceFunction: DistanceByName("iou"),
//	}
func DistanceByName(name string) Distance {
	return GetDistanceByName(name)
}
