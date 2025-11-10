// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause
//
// This file contains a Go port of scipy.spatial.distance.cdist
//
// 1. scipy
//   Original Source: https://github.com/scipy/scipy/blob/main/scipy/spatial/distance.py
//   Original Copyright (c) 2001-2002 Enthought, Inc. 2003-2024, SciPy Developers
//   Original License: BSD-3-Clause

package scipy

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// Cdist computes pairwise distances between two sets of vectors.
//
// This is a Go port of scipy.spatial.distance.cdist which computes distances
// between each pair of observations in XA and XB using the specified metric.
//
// Parameters:
//   - XA: First set of vectors (m x n matrix)
//   - XB: Second set of vectors (p x n matrix)
//   - metric: Distance metric name ("euclidean", "cityblock"/"manhattan", "cosine", "sqeuclidean", "chebyshev")
//
// Returns:
//   - Distance matrix of shape (m x p)
//
// Panics if XA and XB have different number of columns or if metric is unsupported.
//
// Reference: https://github.com/scipy/scipy/blob/main/scipy/spatial/distance.py#L2233
func Cdist(XA, XB *mat.Dense, metric string) *mat.Dense {
	rowsA, colsA := XA.Dims()
	rowsB, colsB := XB.Dims()

	if colsA != colsB {
		panic(fmt.Sprintf("XA and XB must have same number of columns, got %d and %d", colsA, colsB))
	}

	result := mat.NewDense(rowsA, rowsB, nil)

	switch metric {
	case "euclidean":
		for i := 0; i < rowsA; i++ {
			for j := 0; j < rowsB; j++ {
				rowA := XA.RawRowView(i)
				rowB := XB.RawRowView(j)
				var sum float64
				for k := range rowA {
					diff := rowA[k] - rowB[k]
					sum += diff * diff
				}
				result.Set(i, j, math.Sqrt(sum))
			}
		}

	case "cityblock", "manhattan":
		for i := 0; i < rowsA; i++ {
			for j := 0; j < rowsB; j++ {
				rowA := XA.RawRowView(i)
				rowB := XB.RawRowView(j)
				var sum float64
				for k := range rowA {
					sum += math.Abs(rowA[k] - rowB[k])
				}
				result.Set(i, j, sum)
			}
		}

	case "cosine":
		for i := 0; i < rowsA; i++ {
			for j := 0; j < rowsB; j++ {
				rowA := XA.RawRowView(i)
				rowB := XB.RawRowView(j)
				var dot, normA, normB float64
				for k := range rowA {
					dot += rowA[k] * rowB[k]
					normA += rowA[k] * rowA[k]
					normB += rowB[k] * rowB[k]
				}
				normA = math.Sqrt(normA)
				normB = math.Sqrt(normB)
				if normA == 0 || normB == 0 {
					result.Set(i, j, 0)
				} else {
					// Cosine distance = 1 - cosine similarity
					result.Set(i, j, 1.0-dot/(normA*normB))
				}
			}
		}

	case "sqeuclidean":
		for i := 0; i < rowsA; i++ {
			for j := 0; j < rowsB; j++ {
				rowA := XA.RawRowView(i)
				rowB := XB.RawRowView(j)
				var sum float64
				for k := range rowA {
					diff := rowA[k] - rowB[k]
					sum += diff * diff
				}
				result.Set(i, j, sum)
			}
		}

	case "chebyshev":
		for i := 0; i < rowsA; i++ {
			for j := 0; j < rowsB; j++ {
				rowA := XA.RawRowView(i)
				rowB := XB.RawRowView(j)
				var maxDiff float64
				for k := range rowA {
					diff := math.Abs(rowA[k] - rowB[k])
					if diff > maxDiff {
						maxDiff = diff
					}
				}
				result.Set(i, j, maxDiff)
			}
		}

	default:
		panic(fmt.Sprintf("unsupported metric: %s", metric))
	}

	return result
}
