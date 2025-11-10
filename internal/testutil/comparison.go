package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"gocv.io/x/gocv"
)

// ImageSimilarity compares two images and returns similarity ratio (0.0 to 1.0).
// Uses pixel-by-pixel comparison with tolerance for anti-aliasing differences.
func ImageSimilarity(img1, img2 *gocv.Mat, pixelTolerance int) float64 {
	if img1.Rows() != img2.Rows() || img1.Cols() != img2.Cols() {
		return 0.0
	}

	totalPixels := img1.Rows() * img1.Cols() * img1.Channels()
	matchingPixels := 0

	for y := 0; y < img1.Rows(); y++ {
		for x := 0; x < img1.Cols(); x++ {
			pixel1 := img1.GetVecbAt(y, x)
			pixel2 := img2.GetVecbAt(y, x)

			channelMatches := 0
			for c := 0; c < img1.Channels(); c++ {
				diff := int(pixel1[c]) - int(pixel2[c])
				if diff < 0 {
					diff = -diff
				}
				if diff <= pixelTolerance {
					channelMatches++
				}
			}

			if channelMatches == img1.Channels() {
				matchingPixels += img1.Channels()
			}
		}
	}

	return float64(matchingPixels) / float64(totalPixels)
}

// CompareToGoldenImage compares generated image to golden reference.
// similarity is the minimum required similarity ratio (0.0 to 1.0).
func CompareToGoldenImage(t *testing.T, actual *gocv.Mat, goldenPath string, similarity float64) {
	t.Helper()

	// Load golden image
	golden := gocv.IMRead(goldenPath, gocv.IMReadColor)
	if golden.Empty() {
		t.Fatalf("Failed to load golden image: %s", goldenPath)
	}
	defer golden.Close()

	// Compare dimensions
	if actual.Rows() != golden.Rows() || actual.Cols() != golden.Cols() {
		t.Errorf("Image dimensions mismatch: got %dx%d, want %dx%d",
			actual.Rows(), actual.Cols(), golden.Rows(), golden.Cols())
		return
	}

	// Compare pixels with tolerance for anti-aliasing (allow ~5 pixel value difference)
	pixelTolerance := 5
	actualSimilarity := ImageSimilarity(actual, &golden, pixelTolerance)

	if actualSimilarity < similarity {
		t.Errorf("Image similarity %.2f%% below threshold %.2f%%",
			actualSimilarity*100, similarity*100)

		// Optionally save diff for debugging
		diffPath := goldenPath + ".diff.png"
		diff := gocv.NewMat()
		defer diff.Close()
		gocv.AbsDiff(*actual, golden, &diff)
		gocv.IMWrite(diffPath, diff)
		t.Logf("Saved diff to: %s", diffPath)
	}
}

// SaveGoldenImage saves an image as a golden reference (for generating golden data).
func SaveGoldenImage(path string, img *gocv.Mat) error {
	if !gocv.IMWrite(path, *img) {
		return fmt.Errorf("failed to write image to %s", path)
	}
	return nil
}

// CompareJSON compares two JSON files with float tolerance.
func CompareJSON(t *testing.T, actualPath, goldenPath string, floatTolerance float64) {
	t.Helper()

	// Load actual JSON
	actualData, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("Failed to read actual JSON: %v", err)
	}

	// Load golden JSON
	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden JSON: %v", err)
	}

	// Parse JSON
	var actual, golden interface{}
	if err := json.Unmarshal(actualData, &actual); err != nil {
		t.Fatalf("Failed to parse actual JSON: %v", err)
	}
	if err := json.Unmarshal(goldenData, &golden); err != nil {
		t.Fatalf("Failed to parse golden JSON: %v", err)
	}

	// Deep compare with float tolerance
	if !jsonEqual(actual, golden, floatTolerance) {
		t.Errorf("JSON data mismatch")
		t.Logf("Actual JSON: %s", string(actualData))
		t.Logf("Golden JSON: %s", string(goldenData))
	}
}

// jsonEqual recursively compares JSON structures with float tolerance.
func jsonEqual(a, b interface{}, tolerance float64) bool {
	switch av := a.(type) {
	case float64:
		bv, ok := b.(float64)
		if !ok {
			return false
		}
		return AlmostEqual(av, bv, tolerance)

	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !jsonEqual(v, bv[k], tolerance) {
				return false
			}
		}
		return true

	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonEqual(av[i], bv[i], tolerance) {
				return false
			}
		}
		return true

	default:
		return a == b
	}
}
