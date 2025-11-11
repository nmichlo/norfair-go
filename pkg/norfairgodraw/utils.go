package norfairgodraw

import (
	"fmt"
	"strconv"

	"gonum.org/v1/gonum/mat"
)

// Centroid calculates the centroid (geometric center) of a set of 2D points.
// Returns the (x, y) coordinates as integers.
func Centroid(points *mat.Dense) (int, int) {
	rows, _ := points.Dims()

	var sumX, sumY float64
	for i := 0; i < rows; i++ {
		sumX += points.At(i, 0)
		sumY += points.At(i, 1)
	}

	centroidX := int(sumX / float64(rows))
	centroidY := int(sumY / float64(rows))

	return centroidX, centroidY
}

// BuildText creates formatted text from a Drawable's label, ID, and scores.
// Combines non-nil/non-empty values with "-" separator (no spaces).
// For scores, computes the mean and rounds to 4 decimal places.
//
// Parameters:
//   - drawable: The Drawable object to extract text from
//   - drawLabels: Whether to include the label
//   - drawIDs: Whether to include the ID
//   - drawScores: Whether to include the scores (as mean)
func BuildText(drawable *Drawable, drawLabels, drawIDs, drawScores bool) string {
	text := ""

	// Add label if requested and non-nil
	if drawLabels && drawable.Label != nil {
		text = *drawable.Label
	}

	// Add ID if requested and non-nil
	if drawIDs && drawable.ID != nil {
		if len(text) > 0 {
			text += "-"
		}
		text += fmt.Sprintf("%d", *drawable.ID)
	}

	// Add mean of scores if requested and non-nil/non-empty
	if drawScores && drawable.Scores != nil && len(drawable.Scores) > 0 {
		if len(text) > 0 {
			text += "-"
		}
		// Calculate mean of scores
		var sum float64
		for _, score := range drawable.Scores {
			sum += score
		}
		mean := sum / float64(len(drawable.Scores))
		// Round to 4 decimal places and strip trailing zeros (matches Python's str(np.round()))
		formatted := strconv.FormatFloat(mean, 'f', 4, 64)
		// Strip trailing zeros and decimal point if needed
		formatted = stripTrailingZeros(formatted)
		text += formatted
	}

	return text
}

// stripTrailingZeros removes trailing zeros from a decimal string.
// E.g., "0.9900" -> "0.99", "1.0000" -> "1"
func stripTrailingZeros(s string) string {
	// Only strip if there's a decimal point
	if len(s) == 0 {
		return s
	}

	// Find decimal point
	dotIdx := -1
	for i, c := range s {
		if c == '.' {
			dotIdx = i
			break
		}
	}

	// No decimal point, return as-is
	if dotIdx == -1 {
		return s
	}

	// Strip trailing zeros after decimal
	end := len(s) - 1
	for end > dotIdx && s[end] == '0' {
		end--
	}

	// If all zeros after decimal, remove decimal point too
	if end == dotIdx {
		return s[:dotIdx]
	}

	return s[:end+1]
}
