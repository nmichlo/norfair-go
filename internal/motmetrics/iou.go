// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause
//
// This file contains a Go port of py-motmetrics IoU distance computation:
//
// 1. py-motmetrics
//   Original Source: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/distances.py
//   Original Copyright (c) 2017-2019 Christoph Heindl, Jack Valmadre
//   Original License: MIT

package motmetrics

import (
	"fmt"
	"math"
)

// IouDistance computes IoU-based distance between two bounding boxes.
//
// This is a Go port of py-motmetrics IoU distance computation used for
// MOTChallenge evaluation.
//
// Parameters:
//   - box1: Bounding box [x_min, y_min, x_max, y_max]
//   - box2: Bounding box [x_min, y_min, x_max, y_max]
//
// Returns: 1.0 - IoU (distance in range [0, 1])
//   - 0.0 = perfect overlap (IoU = 1.0)
//   - 1.0 = no overlap (IoU = 0.0)
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/distances.py
func IouDistance(box1, box2 []float64) float64 {
	// Validate input
	if len(box1) != 4 || len(box2) != 4 {
		panic(fmt.Sprintf("boxes must have 4 elements [x_min, y_min, x_max, y_max], got %d and %d", len(box1), len(box2)))
	}

	// Validate box1 coordinates
	if box1[2] <= box1[0] || box1[3] <= box1[1] {
		panic(fmt.Sprintf("invalid box1: x_max (%.2f) <= x_min (%.2f) or y_max (%.2f) <= y_min (%.2f)",
			box1[2], box1[0], box1[3], box1[1]))
	}

	// Validate box2 coordinates
	if box2[2] <= box2[0] || box2[3] <= box2[1] {
		panic(fmt.Sprintf("invalid box2: x_max (%.2f) <= x_min (%.2f) or y_max (%.2f) <= y_min (%.2f)",
			box2[2], box2[0], box2[3], box2[1]))
	}

	// Compute intersection rectangle
	xMinInter := math.Max(box1[0], box2[0])
	yMinInter := math.Max(box1[1], box2[1])
	xMaxInter := math.Min(box1[2], box2[2])
	yMaxInter := math.Min(box1[3], box2[3])

	// Compute intersection area
	var intersection float64
	if xMaxInter < xMinInter || yMaxInter < yMinInter {
		intersection = 0.0 // No overlap
	} else {
		intersection = (xMaxInter - xMinInter) * (yMaxInter - yMinInter)
	}

	// Compute union area
	area1 := (box1[2] - box1[0]) * (box1[3] - box1[1])
	area2 := (box2[2] - box2[0]) * (box2[3] - box2[1])
	union := area1 + area2 - intersection

	// Edge case: zero union (both boxes have zero area)
	if union == 0.0 {
		return 1.0 // Maximum distance
	}

	// Compute IoU and convert to distance
	iou := intersection / union
	return 1.0 - iou
}

// ComputeIoUMatrix computes pairwise IoU distances for all GT × prediction pairs.
//
// This is a Go port of py-motmetrics distance matrix computation used for
// MOTChallenge evaluation.
//
// Parameters:
//   - gtBBoxes: Ground truth bounding boxes, each [x_min, y_min, x_max, y_max]
//   - predBBoxes: Predicted bounding boxes, same format
//
// Returns: Distance matrix [numGT][numPred] where each element is IoU distance (1.0 - IoU)
//
// Reference: https://github.com/cheind/py-motmetrics/blob/master/motmetrics/distances.py
func ComputeIoUMatrix(gtBBoxes, predBBoxes [][]float64) [][]float64 {
	numGT := len(gtBBoxes)
	numPred := len(predBBoxes)

	matrix := make([][]float64, numGT)
	for i := range matrix {
		matrix[i] = make([]float64, numPred)
		for j := range matrix[i] {
			matrix[i][j] = IouDistance(gtBBoxes[i], predBBoxes[j])
		}
	}

	return matrix
}
