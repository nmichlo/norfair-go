// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause
//
// This file contains a Go port of scipy.optimize.linear_sum_assignment behavior
// Original source: https://github.com/scipy/scipy/blob/main/scipy/optimize/_linear_sum_assignment.py
//
// Original Copyright (c) 2001-2002 Enthought, Inc. 2003-2024, SciPy Developers
// Original License: BSD-3-Clause
//
// Uses go-hungarian library (MIT License) by Arthur Kushman for the underlying Hungarian algorithm.
// See LICENSE file in this directory and THIRD_PARTY_LICENSES.md in repository root.

package scipy

import (
	hungarian "github.com/arthurkushman/go-hungarian"
)

// Assignment represents a match between two indices
type Assignment struct {
	RowIdx int // First set index
	ColIdx int // Second set index
}

// LinearSumAssignment solves the linear sum assignment problem.
//
// This is a Go port replicating scipy.optimize.linear_sum_assignment behavior,
// which finds the optimal assignment between two sets to minimize total cost.
//
// Parameters:
//   - costMatrix: 2D cost matrix where cost[i][j] is the cost of assigning row i to column j
//   - maxCost: Maximum cost threshold; assignments with cost > maxCost are rejected
//
// Returns:
//   - assignments: Slice of valid assignments (row, col pairs)
//   - unmatchedRows: Indices of rows that were not matched
//   - unmatchedCols: Indices of columns that were not matched
//
// The function handles rectangular matrices by padding to square.
// It uses the Hungarian algorithm via github.com/arthurkushman/go-hungarian.
//
// Reference: https://github.com/scipy/scipy/blob/main/scipy/optimize/_linear_sum_assignment.py
func LinearSumAssignment(costMatrix [][]float64, maxCost float64) ([]Assignment, []int, []int) {
	numRows := len(costMatrix)
	if numRows == 0 {
		return nil, nil, nil
	}
	numCols := len(costMatrix[0])
	if numCols == 0 {
		unmatchedRows := make([]int, numRows)
		for i := range unmatchedRows {
			unmatchedRows[i] = i
		}
		return nil, unmatchedRows, nil
	}

	// Pad to square matrix and convert cost to profit
	size := max(numRows, numCols)
	profitMatrix := make([][]float64, size)
	maxProfit := 10.0 // Constant for cost-to-profit conversion

	for i := range profitMatrix {
		profitMatrix[i] = make([]float64, size)
		for j := range profitMatrix[i] {
			if i < numRows && j < numCols {
				// Convert cost to profit: profit = maxProfit - cost
				profitMatrix[i][j] = maxProfit - costMatrix[i][j]
			} else {
				// Zero profit for dummy padding
				profitMatrix[i][j] = 0.0
			}
		}
	}

	// Solve using Hungarian algorithm (maximizes profit = minimizes cost)
	result := hungarian.SolveMax(profitMatrix)

	// Extract assignments and filter by max cost
	var assignments []Assignment
	matchedRows := make(map[int]bool)
	matchedCols := make(map[int]bool)

	for rowIdx, cols := range result {
		for colIdx, profit := range cols {
			// Convert profit back to cost
			cost := maxProfit - profit

			// Only accept if within bounds and below threshold
			if rowIdx < numRows && colIdx < numCols && cost <= maxCost {
				assignments = append(assignments, Assignment{
					RowIdx: rowIdx,
					ColIdx: colIdx,
				})
				matchedRows[rowIdx] = true
				matchedCols[colIdx] = true
			}
		}
	}

	// Find unmatched indices
	var unmatchedRows, unmatchedCols []int
	for i := 0; i < numRows; i++ {
		if !matchedRows[i] {
			unmatchedRows = append(unmatchedRows, i)
		}
	}
	for j := 0; j < numCols; j++ {
		if !matchedCols[j] {
			unmatchedCols = append(unmatchedCols, j)
		}
	}

	return assignments, unmatchedRows, unmatchedCols
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
