package norfairgo

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gonum.org/v1/gonum/mat"

	"github.com/nmichlo/norfair-go/internal/motmetrics"
	"github.com/nmichlo/norfair-go/internal/scipy"
)

// =============================================================================
// InformationFile - Parse MOTChallenge seqinfo.ini files
// =============================================================================

// InformationFile parses MOTChallenge seqinfo.ini files to extract metadata.
type InformationFile struct {
	path  string
	lines []string
}

// NewInformationFile creates a new InformationFile by reading the given file path.
func NewInformationFile(filePath string) (*InformationFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open information file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read information file: %w", err)
	}

	return &InformationFile{
		path:  filePath,
		lines: lines,
	}, nil
}

// Search finds a variable in the information file and returns its value.
//
// Parameters:
//   - variableName: The key to search for (e.g., "seqLength", "frameRate")
//
// Returns: Value as interface{} (int if numeric, string otherwise) or error if not found
//
// The format expected is "key=value" on each line.
func (inf *InformationFile) Search(variableName string) (interface{}, error) {
	for _, line := range inf.lines {
		if strings.HasPrefix(line, variableName) {
			// Find the '=' separator
			equalIndex := strings.Index(line, "=")
			if equalIndex == -1 {
				continue
			}

			// Extract value after '='
			result := strings.TrimSpace(line[equalIndex+1:])

			// Try to convert to int
			if intVal, err := strconv.Atoi(result); err == nil {
				return intVal, nil
			}

			// Return as string
			return result, nil
		}
	}

	return nil, fmt.Errorf("couldn't find '%s' in %s", variableName, inf.path)
}

// SearchInt is a convenience method that returns the value as an int.
func (inf *InformationFile) SearchInt(variableName string) (int, error) {
	val, err := inf.Search(variableName)
	if err != nil {
		return 0, err
	}

	if intVal, ok := val.(int); ok {
		return intVal, nil
	}

	return 0, fmt.Errorf("value for '%s' is not an integer", variableName)
}

// SearchString is a convenience method that returns the value as a string.
func (inf *InformationFile) SearchString(variableName string) (string, error) {
	val, err := inf.Search(variableName)
	if err != nil {
		return "", err
	}

	if strVal, ok := val.(string); ok {
		return strVal, nil
	}

	// If it's an int, convert to string
	if intVal, ok := val.(int); ok {
		return strconv.Itoa(intVal), nil
	}

	return "", fmt.Errorf("value for '%s' cannot be converted to string", variableName)
}

// =============================================================================
// PredictionsTextFile - Write tracker predictions to MOTChallenge format
// =============================================================================

// PredictionsTextFile generates a text file with tracked objects in MOTChallenge format.
//
// The output format is CSV with columns:
// frame,id,bb_left,bb_top,bb_width,bb_height,-1,-1,-1,-1
type PredictionsTextFile struct {
	length      int
	textFile    *os.File
	frameNumber int
}

// NewPredictionsTextFile creates a new PredictionsTextFile for writing tracking results.
//
// Parameters:
//   - inputPath: Path to the sequence being processed
//   - savePath: Directory where predictions/ folder will be created
//   - informationFile: Optional InformationFile (if nil, will load from inputPath/seqinfo.ini)
//
// Returns: PredictionsTextFile instance or error
func NewPredictionsTextFile(inputPath string, savePath string, informationFile *InformationFile) (*PredictionsTextFile, error) {
	if savePath == "" {
		savePath = "."
	}

	// Extract sequence name from input path
	fileName := filepath.Base(inputPath)

	// Load information file if not provided
	if informationFile == nil {
		seqinfoPath := filepath.Join(inputPath, "seqinfo.ini")
		var err error
		informationFile, err = NewInformationFile(seqinfoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load information file: %w", err)
		}
	}

	// Get sequence length
	length, err := informationFile.SearchInt("seqLength")
	if err != nil {
		return nil, fmt.Errorf("failed to get seqLength: %w", err)
	}

	// Create predictions folder
	predictionsFolder := filepath.Join(savePath, "predictions")
	if err := os.MkdirAll(predictionsFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create predictions folder: %w", err)
	}

	// Open output file
	outFileName := filepath.Join(predictionsFolder, fileName+".txt")
	textFile, err := os.Create(outFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return &PredictionsTextFile{
		length:      length,
		textFile:    textFile,
		frameNumber: 1,
	}, nil
}

// Update writes tracked object information for the current frame.
//
// Parameters:
//   - predictions: List of TrackedObject instances
//   - frameNumber: Optional frame number (if nil, uses auto-incremented counter)
//
// Format: frame_number,id,bb_left,bb_top,bb_width,bb_height,-1,-1,-1,-1
func (ptf *PredictionsTextFile) Update(predictions []*TrackedObject, frameNumber *int) error {
	// Use provided frame number or auto-increment
	frame := ptf.frameNumber
	if frameNumber != nil {
		frame = *frameNumber
	}

	// Write each prediction as CSV row
	for _, obj := range predictions {
		if obj.ID == nil {
			continue // Skip objects without IDs
		}

		// Extract bounding box coordinates
		// Python: obj.estimate[0, 0], obj.estimate[0, 1], obj.estimate[1, 0], obj.estimate[1, 1]
		bbLeft := obj.Estimate.At(0, 0)
		bbTop := obj.Estimate.At(0, 1)
		bbWidth := obj.Estimate.At(1, 0) - obj.Estimate.At(0, 0)
		bbHeight := obj.Estimate.At(1, 1) - obj.Estimate.At(0, 1)

		// Format: frame,id,bb_left,bb_top,bb_width,bb_height,-1,-1,-1,-1
		line := fmt.Sprintf("%d,%d,%f,%f,%f,%f,-1,-1,-1,-1\n",
			frame, *obj.ID, bbLeft, bbTop, bbWidth, bbHeight)

		if _, err := ptf.textFile.WriteString(line); err != nil {
			return fmt.Errorf("failed to write prediction: %w", err)
		}
	}

	// Auto-increment frame number
	ptf.frameNumber++

	// Auto-close when sequence complete
	if ptf.frameNumber > ptf.length {
		if err := ptf.textFile.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
		ptf.textFile = nil // Set to nil to prevent double close
	}

	return nil
}

// Close closes the output file (useful for manual cleanup).
// Safe to call multiple times (idempotent).
func (ptf *PredictionsTextFile) Close() error {
	if ptf.textFile != nil {
		err := ptf.textFile.Close()
		ptf.textFile = nil // Set to nil to prevent double close
		return err
	}
	return nil
}

// =============================================================================
// DetectionFileParser - Load MOTChallenge format detections/ground truth
// =============================================================================

// DetectionFileParser loads detections from MOTChallenge format text files.
//
// The input format is CSV with columns:
// frame,id,bb_left,bb_top,bb_width,bb_height,conf,x,y,z
//
// This converts to corner format:
// frame,id,x_min,y_min,x_max,y_max,conf,x,y,z
type DetectionFileParser struct {
	frameNumber      int
	matrixDetections [][]float64    // All detections (N x 10 matrix)
	length           int            // Sequence length
	sortedByFrame    [][]*Detection // Pre-indexed detections by frame
}

// NewDetectionFileParser creates a new DetectionFileParser.
//
// Parameters:
//   - inputPath: Path to sequence directory
//   - informationFile: Optional InformationFile (if nil, will load from inputPath/seqinfo.ini)
//
// Returns: DetectionFileParser instance or error
func NewDetectionFileParser(inputPath string, informationFile *InformationFile) (*DetectionFileParser, error) {
	// Load detections CSV file
	detectionsPath := filepath.Join(inputPath, "det/det.txt")
	file, err := os.Open(detectionsPath)
	if err != nil {
		// Try ground truth path as fallback
		detectionsPath = filepath.Join(inputPath, "gt/gt.txt")
		file, err = os.Open(detectionsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open detections file: %w", err)
		}
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Convert to float64 matrix
	matrixDetections := make([][]float64, len(records))
	for i, record := range records {
		row := make([]float64, len(record))
		for j, val := range record {
			row[j], _ = strconv.ParseFloat(val, 64)
		}
		matrixDetections[i] = row
	}

	// Sort by frame number (column 0)
	// Python: row_order = np.argsort(self.matrix_detections[:, 0])
	sortByFrame(matrixDetections)

	// Convert width/height to corner format
	// Python: self.matrix_detections[:, 4] = self.matrix_detections[:, 2] + self.matrix_detections[:, 4]
	// Python: self.matrix_detections[:, 5] = self.matrix_detections[:, 3] + self.matrix_detections[:, 5]
	for i := range matrixDetections {
		if len(matrixDetections[i]) >= 6 {
			matrixDetections[i][4] = matrixDetections[i][2] + matrixDetections[i][4] // x_max = x + width
			matrixDetections[i][5] = matrixDetections[i][3] + matrixDetections[i][5] // y_max = y + height
		}
	}

	// Load information file if not provided
	if informationFile == nil {
		seqinfoPath := filepath.Join(inputPath, "seqinfo.ini")
		informationFile, err = NewInformationFile(seqinfoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load information file: %w", err)
		}
	}

	// Get sequence length
	length, err := informationFile.SearchInt("seqLength")
	if err != nil {
		return nil, fmt.Errorf("failed to get seqLength: %w", err)
	}

	// Create parser instance
	parser := &DetectionFileParser{
		frameNumber:      1,
		matrixDetections: matrixDetections,
		length:           length,
		sortedByFrame:    make([][]*Detection, length),
	}

	// Pre-index detections by frame
	for frameNum := 1; frameNum <= length; frameNum++ {
		parser.sortedByFrame[frameNum-1] = parser.getDetectionsFromFrame(frameNum)
	}

	return parser, nil
}

// getDetectionsFromFrame returns a list of Detection objects for the given frame.
func (dfp *DetectionFileParser) getDetectionsFromFrame(frameNumber int) []*Detection {
	var detections []*Detection

	// Find rows where column 0 == frameNumber
	for _, row := range dfp.matrixDetections {
		if len(row) >= 7 && int(row[0]) == frameNumber {
			// Extract bounding box corners
			// Python: points = np.array([[det[2], det[3]], [det[4], det[5]]])
			points := mat.NewDense(2, 2, []float64{
				row[2], row[3], // x_min, y_min
				row[4], row[5], // x_max, y_max
			})

			// Extract confidence
			conf := row[6]

			// Create Detection with scores for both corners
			// Python: Detection(points, np.array([conf, conf]))
			detection := &Detection{
				Points: points,
				Scores: []float64{conf, conf},
			}

			detections = append(detections, detection)
		}
	}

	return detections
}

// Detections returns a channel that iterates through detections frame by frame.
//
// This implements the iterator protocol using Go channels (matches video.go pattern).
func (dfp *DetectionFileParser) Detections() <-chan []*Detection {
	ch := make(chan []*Detection)
	go func() {
		defer close(ch)
		for frame := 1; frame <= dfp.length; frame++ {
			ch <- dfp.sortedByFrame[frame-1]
		}
	}()
	return ch
}

// Length returns the sequence length.
func (dfp *DetectionFileParser) Length() int {
	return dfp.length
}

// =============================================================================
// Helper Functions
// =============================================================================

// sortByFrame sorts a 2D matrix by the first column (frame number).
func sortByFrame(matrix [][]float64) {
	// Simple bubble sort (good enough for sorting by frame)
	for i := 0; i < len(matrix); i++ {
		for j := i + 1; j < len(matrix); j++ {
			if matrix[i][0] > matrix[j][0] {
				matrix[i], matrix[j] = matrix[j], matrix[i]
			}
		}
	}
}

// Note: IoU distance computation moved to internal/motmetrics package

// =============================================================================
// Hungarian Algorithm - Optimal Assignment Matching
// =============================================================================

// hungarianMatching performs optimal assignment matching with threshold filtering.
//
// This wraps scipy.LinearSumAssignment to match py-motmetrics behavior for
// MOTChallenge evaluation.
//
// Parameters:
//   - distanceMatrix: [][]float64 of shape [numGT, numPred]
//   - threshold: maximum distance for valid match (default 0.5 for IoU ≥ 0.5)
//
// Returns:
//   - matches: [][2]int, each element is [gtIdx, predIdx]
//   - unmatchedGT: []int, indices of unmatched ground truth objects
//   - unmatchedPred: []int, indices of unmatched predictions
func hungarianMatching(distanceMatrix [][]float64, threshold float64) ([][2]int, []int, []int) {
	// Use scipy.LinearSumAssignment for optimal matching
	assignments, unmatchedRows, unmatchedCols := scipy.LinearSumAssignment(distanceMatrix, threshold)

	// Convert scipy.Assignment to [][2]int format
	var matches [][2]int
	if len(assignments) > 0 {
		matches = make([][2]int, len(assignments))
		for i, assign := range assignments {
			matches[i] = [2]int{assign.RowIdx, assign.ColIdx}
		}
	}

	return matches, unmatchedRows, unmatchedCols
}

// Note: TrackLifecycle and MOTAccumulator moved to internal/motmetrics package

// =============================================================================
// MetricsDataFrame - DataFrame-like Structure for Metrics
// =============================================================================

// MetricsRow represents metrics for a single video or aggregate summary.
//
// This matches a single row in pandas DataFrame returned by py-motmetrics.
type MetricsRow struct {
	VideoName string // Name of video sequence (or "OVERALL" for aggregate)

	// Primary MOTChallenge metrics
	MOTA float64 // Multi-Object Tracking Accuracy (range: -∞ to 1.0)
	MOTP float64 // Multi-Object Tracking Precision (average IoU distance)

	// Event counts
	NumMatches        int // True positives
	NumFalsePositives int // False positives
	NumMisses         int // False negatives (missed detections)
	NumSwitches       int // ID switches
	NumObjects        int // Total ground truth objects

	// Derived metrics
	Precision float64 // TP / (TP + FP)
	Recall    float64 // TP / (TP + FN) = TP / NumObjects

	// Extended MOTChallenge metrics (Phase 2)
	NumFragmentations int     // Track fragmentations
	MT                float64 // Mostly Tracked (% of GT tracks covered >= 80%)
	ML                float64 // Mostly Lost (% of GT tracks covered <= 20%)
	PT                float64 // Partially Tracked (% of GT tracks 20% < covered < 80%)

	// ID metrics (Phase 2.3)
	IDP  float64 // ID Precision
	IDR  float64 // ID Recall
	IDF1 float64 // ID F1-Score
}

// MetricsDataFrame holds metrics for multiple videos in a DataFrame-like structure.
//
// Matches pandas.DataFrame returned by py-motmetrics compute_many()
// Rows are indexed by video name, with optional "OVERALL" aggregate row.
type MetricsDataFrame struct {
	Rows []MetricsRow // Per-video metrics + optional OVERALL
}

// NewMetricsDataFrame creates an empty DataFrame.
func NewMetricsDataFrame() *MetricsDataFrame {
	return &MetricsDataFrame{
		Rows: make([]MetricsRow, 0),
	}
}

// AddRow adds a metrics row to the DataFrame.
func (df *MetricsDataFrame) AddRow(row MetricsRow) {
	df.Rows = append(df.Rows, row)
}

// GetRow retrieves a row by video name.
//
// Returns: MetricsRow and true if found, zero value and false otherwise.
func (df *MetricsDataFrame) GetRow(videoName string) (MetricsRow, bool) {
	for _, row := range df.Rows {
		if row.VideoName == videoName {
			return row, true
		}
	}
	return MetricsRow{}, false
}

// metricExtractors maps metric names to functions that extract values from MetricsRow
var metricExtractors = map[string]func(MetricsRow) float64{
	"MOTA":              func(r MetricsRow) float64 { return r.MOTA },
	"MOTP":              func(r MetricsRow) float64 { return r.MOTP },
	"Precision":         func(r MetricsRow) float64 { return r.Precision },
	"Recall":            func(r MetricsRow) float64 { return r.Recall },
	"MT":                func(r MetricsRow) float64 { return r.MT },
	"ML":                func(r MetricsRow) float64 { return r.ML },
	"PT":                func(r MetricsRow) float64 { return r.PT },
	"IDP":               func(r MetricsRow) float64 { return r.IDP },
	"IDR":               func(r MetricsRow) float64 { return r.IDR },
	"IDF1":              func(r MetricsRow) float64 { return r.IDF1 },
	"NumMatches":        func(r MetricsRow) float64 { return float64(r.NumMatches) },
	"NumFalsePositives": func(r MetricsRow) float64 { return float64(r.NumFalsePositives) },
	"NumMisses":         func(r MetricsRow) float64 { return float64(r.NumMisses) },
	"NumSwitches":       func(r MetricsRow) float64 { return float64(r.NumSwitches) },
	"NumObjects":        func(r MetricsRow) float64 { return float64(r.NumObjects) },
	"NumFragmentations": func(r MetricsRow) float64 { return float64(r.NumFragmentations) },
}

// Get retrieves a specific metric value for a video.
//
// Parameters:
//   - videoName: Name of video (or "OVERALL")
//   - metricName: Name of metric field (e.g., "MOTA", "MOTP", "IDF1")
//
// Returns: Metric value as float64, true if found, 0.0 and false otherwise.
func (df *MetricsDataFrame) Get(videoName, metricName string) (float64, bool) {
	row, found := df.GetRow(videoName)
	if !found {
		return 0.0, false
	}

	extractor, exists := metricExtractors[metricName]
	if !exists {
		return 0.0, false
	}

	return extractor(row), true
}

// =============================================================================
// Accumulators - Multi-Video Accumulator Manager
// =============================================================================

// Accumulators manages multiple MOTAccumulator instances for multi-video evaluation.
//
// This is thread-safe for concurrent accumulation across different videos.
type Accumulators struct {
	accumulators map[string]*motmetrics.MOTAccumulator // map[videoName]*accumulator
	mu           sync.Mutex                            // Thread-safety for concurrent updates
}

// NewAccumulators creates a new multi-video accumulator manager.
//
// Returns: Initialized Accumulators instance
func NewAccumulators() *Accumulators {
	return &Accumulators{
		accumulators: make(map[string]*motmetrics.MOTAccumulator),
	}
}

// CreateAccumulator creates a new accumulator for a video sequence.
//
// Parameters:
//   - videoName: Name of the video sequence
//
// Returns: Error if accumulator already exists for this video
func (a *Accumulators) CreateAccumulator(videoName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.accumulators[videoName]; exists {
		return fmt.Errorf("accumulator for video '%s' already exists", videoName)
	}

	a.accumulators[videoName] = motmetrics.NewMOTAccumulator(videoName)
	return nil
}

// Update processes a frame for a specific video.
//
// Parameters:
//   - gtBBoxes: ground truth bounding boxes
//   - gtIDs: ground truth object IDs
//   - predBBoxes: predicted bounding boxes
//   - predIDs: tracker object IDs
//   - videoName: video sequence name
//   - threshold: IoU distance threshold (default 0.5)
//
// Returns: Error if accumulator doesn't exist
func (a *Accumulators) Update(gtBBoxes [][]float64, gtIDs []int, predBBoxes [][]float64, predIDs []int, videoName string, threshold float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	acc, exists := a.accumulators[videoName]
	if !exists {
		return fmt.Errorf("accumulator for video '%s' not found, call CreateAccumulator first", videoName)
	}

	acc.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, threshold, hungarianMatching)
	return nil
}

// Metrics contains computed MOTChallenge metrics for evaluation output.
//
// This matches the output format of py-motmetrics compute_many().
type Metrics struct {
	// Primary MOTChallenge metrics
	MOTA float64 // Multi-Object Tracking Accuracy (range: -∞ to 1.0)
	MOTP float64 // Multi-Object Tracking Precision (average IoU distance)

	// Event counts
	NumMatches        int // True positives
	NumFalsePositives int // False positives
	NumMisses         int // False negatives (missed detections)
	NumSwitches       int // ID switches
	NumObjects        int // Total ground truth objects

	// Derived metrics
	Precision float64 // TP / (TP + FP)
	Recall    float64 // TP / (TP + FN) = TP / NumObjects

	// Extended MOTChallenge metrics
	NumFragmentations int     // Track fragmentations
	MT                float64 // Mostly Tracked (% of GT tracks covered >= 80%)
	ML                float64 // Mostly Lost (% of GT tracks covered <= 20%)
	PT                float64 // Partially Tracked (% of GT tracks 20% < covered < 80%)
	MTCount           int     // Mostly Tracked count (matches py-motmetrics output)
	MLCount           int     // Mostly Lost count (matches py-motmetrics output)
	PTCount           int     // Partially Tracked count (matches py-motmetrics output)
	NumTracks         int     // Total number of unique ground truth tracks

	// ID metrics (Phase 2.3)
	IDP  float64 // ID Precision
	IDR  float64 // ID Recall
	IDF1 float64 // ID F1-Score
}

// ComputeMetrics aggregates all accumulators and computes final metrics.
//
// Returns: Metrics struct with computed values, or error
//
// Edge cases:
//   - MOTA when numObjects == 0 → return 0.0 (not NaN)
//   - MOTP when numMatches == 0 → return NaN
func (a *Accumulators) ComputeMetrics() (*Metrics, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Aggregate across all videos
	totalMatches := 0
	totalFP := 0
	totalFN := 0
	totalSwitches := 0
	totalObjects := 0
	totalDistance := 0.0

	// Extended metrics aggregation
	totalMT := 0
	totalML := 0
	totalPT := 0
	totalFragmentations := 0
	totalTracks := 0

	for _, acc := range a.accumulators {
		totalMatches += acc.NumMatches
		totalFP += acc.NumFalsePositives
		totalFN += acc.NumMisses
		totalSwitches += acc.NumSwitches
		totalObjects += acc.NumObjects
		totalDistance += acc.TotalDistance

		// Compute extended metrics for this accumulator
		mt, ml, pt, frag := acc.ComputeExtendedMetrics()
		totalMT += mt
		totalML += ml
		totalPT += pt
		totalFragmentations += frag
		totalTracks += len(acc.TrackLifecycles)
	}

	// Compute MOTA
	// Formula: MOTA = 1 - (FP + FN + IDS) / GT
	var mota float64
	if totalObjects == 0 {
		mota = 0.0 // Edge case: no ground truth objects
	} else {
		mota = 1.0 - float64(totalFP+totalFN+totalSwitches)/float64(totalObjects)
	}

	// Compute MOTP
	// Formula: MOTP = sum(distances) / num_matches
	var motp float64
	if totalMatches == 0 {
		motp = math.NaN() // Edge case: no matches
	} else {
		motp = totalDistance / float64(totalMatches)
	}

	// Compute Precision and Recall
	var precision, recall float64
	if totalMatches+totalFP == 0 {
		precision = 0.0
	} else {
		precision = float64(totalMatches) / float64(totalMatches+totalFP)
	}
	if totalObjects == 0 {
		recall = 0.0
	} else {
		recall = float64(totalMatches) / float64(totalObjects)
	}

	// Compute MT/ML/PT percentages
	var mtPercent, mlPercent, ptPercent float64
	if totalTracks == 0 {
		mtPercent = 0.0
		mlPercent = 0.0
		ptPercent = 0.0
	} else {
		mtPercent = float64(totalMT) / float64(totalTracks) * 100.0
		mlPercent = float64(totalML) / float64(totalTracks) * 100.0
		ptPercent = float64(totalPT) / float64(totalTracks) * 100.0
	}

	return &Metrics{
		MOTA:              mota,
		MOTP:              motp,
		NumMatches:        totalMatches,
		NumFalsePositives: totalFP,
		NumMisses:         totalFN,
		NumSwitches:       totalSwitches,
		NumObjects:        totalObjects,
		Precision:         precision,
		Recall:            recall,
		NumFragmentations: totalFragmentations,
		MT:                mtPercent,
		ML:                mlPercent,
		PT:                ptPercent,
		MTCount:           totalMT,
		MLCount:           totalML,
		PTCount:           totalPT,
		NumTracks:         totalTracks,
		IDP:               0.0, // Phase 2.3
		IDR:               0.0, // Phase 2.3
		IDF1:              0.0, // Phase 2.3
	}, nil
}

// PrintMetrics prints a formatted summary of computed metrics.
//
// Returns: Error if metric computation fails
func (a *Accumulators) PrintMetrics() error {
	metrics, err := a.ComputeMetrics()
	if err != nil {
		return err
	}

	fmt.Println("MOT Metrics Summary")
	fmt.Println("==================")
	fmt.Printf("MOTA:              %.6f\n", metrics.MOTA)
	if math.IsNaN(metrics.MOTP) {
		fmt.Printf("MOTP:              NaN (no matches)\n")
	} else {
		fmt.Printf("MOTP:              %.6f\n", metrics.MOTP)
	}
	fmt.Printf("Precision:         %.6f\n", metrics.Precision)
	fmt.Printf("Recall:            %.6f\n", metrics.Recall)
	fmt.Println("------------------")
	fmt.Printf("Matches (TP):      %d\n", metrics.NumMatches)
	fmt.Printf("False Positives:   %d\n", metrics.NumFalsePositives)
	fmt.Printf("Misses (FN):       %d\n", metrics.NumMisses)
	fmt.Printf("ID Switches:       %d\n", metrics.NumSwitches)
	fmt.Printf("Total GT Objects:  %d\n", metrics.NumObjects)

	return nil
}

// SaveMetrics exports metrics to a CSV file.
//
// Parameters:
//   - filePath: Path to output CSV file
//
// Returns: Error if file creation or metric computation fails
func (a *Accumulators) SaveMetrics(filePath string) error {
	metrics, err := a.ComputeMetrics()
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create metrics file: %w", err)
	}
	defer file.Close()

	// CSV header
	fmt.Fprintf(file, "MOTA,MOTP,Precision,Recall,Matches,FalsePositives,Misses,Switches,Objects\n")

	// CSV data
	if math.IsNaN(metrics.MOTP) {
		fmt.Fprintf(file, "%.6f,NaN,%.6f,%.6f,%d,%d,%d,%d,%d\n",
			metrics.MOTA, metrics.Precision, metrics.Recall,
			metrics.NumMatches, metrics.NumFalsePositives, metrics.NumMisses,
			metrics.NumSwitches, metrics.NumObjects)
	} else {
		fmt.Fprintf(file, "%.6f,%.6f,%.6f,%.6f,%d,%d,%d,%d,%d\n",
			metrics.MOTA, metrics.MOTP, metrics.Precision, metrics.Recall,
			metrics.NumMatches, metrics.NumFalsePositives, metrics.NumMisses,
			metrics.NumSwitches, metrics.NumObjects)
	}

	return nil
}

// Reset clears all accumulators.
func (a *Accumulators) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.accumulators = make(map[string]*motmetrics.MOTAccumulator)
}

// =============================================================================
// Module-Level Functions - MOTChallenge Batch Evaluation
// =============================================================================

// MOTChallengeData holds parsed MOTChallenge format data for a single video.
//
// Data is organized by frame number (1-indexed) for efficient frame-by-frame access.
type MOTChallengeData struct {
	VideoName string
	Frames    map[int]*MOTChallengeFrame // map[frameID]*frame
}

// MOTChallengeFrame holds all detections/tracks for a single frame.
type MOTChallengeFrame struct {
	FrameID int
	BBoxes  [][]float64 // [x_min, y_min, x_max, y_max]
	IDs     []int
}

// LoadMotchallenge loads MOTChallenge format CSV file into structured data.
//
// Parameters:
//   - csvPath: Path to MOTChallenge CSV file (gt.txt or predictions.txt)
//
// Returns: MOTChallengeData with frames organized by frame number
//
// CSV Format: frame,id,bb_left,bb_top,bb_width,bb_height,conf,x,y,z
func LoadMotchallenge(csvPath string) (*MOTChallengeData, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open MOTChallenge file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	data := &MOTChallengeData{
		VideoName: filepath.Base(filepath.Dir(csvPath)), // Extract video name from path
		Frames:    make(map[int]*MOTChallengeFrame),
	}

	for _, record := range records {
		if len(record) < 6 {
			continue // Skip invalid rows
		}

		// Parse fields
		frameID, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		id, err := strconv.Atoi(record[1])
		if err != nil {
			continue
		}
		bbLeft, _ := strconv.ParseFloat(record[2], 64)
		bbTop, _ := strconv.ParseFloat(record[3], 64)
		bbWidth, _ := strconv.ParseFloat(record[4], 64)
		bbHeight, _ := strconv.ParseFloat(record[5], 64)

		// Convert to corner format [x_min, y_min, x_max, y_max]
		bbox := []float64{
			bbLeft,
			bbTop,
			bbLeft + bbWidth,
			bbTop + bbHeight,
		}

		// Get or create frame
		frame, exists := data.Frames[frameID]
		if !exists {
			frame = &MOTChallengeFrame{
				FrameID: frameID,
				BBoxes:  make([][]float64, 0),
				IDs:     make([]int, 0),
			}
			data.Frames[frameID] = frame
		}

		// Add detection to frame
		frame.BBoxes = append(frame.BBoxes, bbox)
		frame.IDs = append(frame.IDs, id)
	}

	return data, nil
}

// CompareDataframes performs MOTChallenge evaluation on loaded GT and predictions.
//
// Parameters:
//   - gt: Ground truth MOTChallenge data
//   - predictions: Tracker predictions MOTChallenge data
//   - distanceFunc: Distance function name ("iou", "euclidean", etc.)
//   - threshold: Distance threshold for valid matches (default 0.5 for IoU)
//
// Returns: Populated Accumulators with all frames processed
func CompareDataframes(gt, predictions *MOTChallengeData, distanceFunc string, threshold float64) (*Accumulators, error) {
	// Only IoU distance supported for now (Phase 3)
	if distanceFunc != "iou" && distanceFunc != "" {
		return nil, fmt.Errorf("unsupported distance function: %s (only 'iou' supported)", distanceFunc)
	}

	accumulators := NewAccumulators()
	videoName := gt.VideoName
	if err := accumulators.CreateAccumulator(videoName); err != nil {
		return nil, err
	}

	// Determine frame range (union of GT and prediction frames)
	allFrameIDs := make(map[int]bool)
	for frameID := range gt.Frames {
		allFrameIDs[frameID] = true
	}
	for frameID := range predictions.Frames {
		allFrameIDs[frameID] = true
	}

	// Convert to sorted slice
	frameIDs := make([]int, 0, len(allFrameIDs))
	for frameID := range allFrameIDs {
		frameIDs = append(frameIDs, frameID)
	}
	// Sort for deterministic processing
	for i := 0; i < len(frameIDs); i++ {
		for j := i + 1; j < len(frameIDs); j++ {
			if frameIDs[i] > frameIDs[j] {
				frameIDs[i], frameIDs[j] = frameIDs[j], frameIDs[i]
			}
		}
	}

	// Process each frame
	for _, frameID := range frameIDs {
		gtFrame := gt.Frames[frameID]
		predFrame := predictions.Frames[frameID]

		var gtBBoxes [][]float64
		var gtIDs []int
		var predBBoxes [][]float64
		var predIDs []int

		if gtFrame != nil {
			gtBBoxes = gtFrame.BBoxes
			gtIDs = gtFrame.IDs
		}
		if predFrame != nil {
			predBBoxes = predFrame.BBoxes
			predIDs = predFrame.IDs
		}

		// Update accumulator for this frame
		if err := accumulators.Update(gtBBoxes, gtIDs, predBBoxes, predIDs, videoName, threshold); err != nil {
			return nil, err
		}
	}

	return accumulators, nil
}

// EvalMotChallenge performs complete MOTChallenge evaluation from file paths.
//
// Parameters:
//   - gtPath: Path to ground truth CSV file (e.g., "gt/gt.txt")
//   - predPath: Path to predictions CSV file (e.g., "predictions.txt")
//   - metricsToCompute: List of metric names to compute (nil = all metrics)
//
// Returns: Metrics struct with computed values
//
// This is the primary user-facing function for MOTChallenge evaluation.
func EvalMotChallenge(gtPath, predPath string, metricsToCompute []string) (*Metrics, error) {
	// Load ground truth
	gt, err := LoadMotchallenge(gtPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ground truth: %w", err)
	}

	// Load predictions
	predictions, err := LoadMotchallenge(predPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load predictions: %w", err)
	}

	// Compare and accumulate
	accumulators, err := CompareDataframes(gt, predictions, "iou", 0.5)
	if err != nil {
		return nil, fmt.Errorf("failed to compare dataframes: %w", err)
	}

	// Compute metrics
	metrics, err := accumulators.ComputeMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to compute metrics: %w", err)
	}

	// TODO Phase 4: Add metrics filtering based on metricsToCompute parameter

	return metrics, nil
}
