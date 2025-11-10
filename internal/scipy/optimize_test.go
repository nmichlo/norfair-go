// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause

package scipy

import (
	"testing"
)

func TestLinearSumAssignment_BasicSquare(t *testing.T) {
	// Simple 3x3 cost matrix
	costMatrix := [][]float64{
		{1, 2, 3},
		{2, 4, 6},
		{3, 6, 9},
	}

	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 10.0)

	// Should have 3 assignments
	if len(assignments) != 3 {
		t.Errorf("Expected 3 assignments, got %d", len(assignments))
	}

	// Check no unmatched
	if len(unmatchedRows) != 0 || len(unmatchedCols) != 0 {
		t.Errorf("Expected no unmatched, got %d rows and %d cols", len(unmatchedRows), len(unmatchedCols))
	}

	// Verify all rows and cols are matched exactly once
	matchedRows := make(map[int]bool)
	matchedCols := make(map[int]bool)
	for _, a := range assignments {
		if matchedRows[a.RowIdx] {
			t.Errorf("Row %d matched multiple times", a.RowIdx)
		}
		if matchedCols[a.ColIdx] {
			t.Errorf("Col %d matched multiple times", a.ColIdx)
		}
		matchedRows[a.RowIdx] = true
		matchedCols[a.ColIdx] = true
	}
}

func TestLinearSumAssignment_CostThreshold(t *testing.T) {
	costMatrix := [][]float64{
		{1, 2, 10},
		{2, 1, 11},
		{10, 11, 1},
	}

	// With maxCost=5, high cost assignments should be rejected
	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 5.0)

	// Verify all accepted assignments are below threshold
	for _, a := range assignments {
		cost := costMatrix[a.RowIdx][a.ColIdx]
		if cost > 5.0 {
			t.Errorf("Assignment (%d, %d) has cost %v which exceeds maxCost 5.0", a.RowIdx, a.ColIdx, cost)
		}
	}

	// Total matched + unmatched should equal matrix dimensions
	totalMatched := len(assignments)
	if totalMatched+len(unmatchedRows) < 3 || totalMatched+len(unmatchedCols) < 3 {
		t.Errorf("Missing rows or cols in results")
	}
}

func TestLinearSumAssignment_RectangularMoreRows(t *testing.T) {
	// 4 rows x 2 cols
	costMatrix := [][]float64{
		{1, 5},
		{3, 2},
		{4, 6},
		{2, 3},
	}

	assignments, unmatchedRows, _ := LinearSumAssignment(costMatrix, 10.0)

	// At most 2 assignments (limited by number of cols)
	if len(assignments) > 2 {
		t.Errorf("Expected at most 2 assignments, got %d", len(assignments))
	}

	// Should have at least 2 unmatched rows
	if len(unmatchedRows) < 2 {
		t.Errorf("Expected at least 2 unmatched rows, got %d", len(unmatchedRows))
	}

	// Verify unmatched rows are valid
	for _, r := range unmatchedRows {
		if r < 0 || r >= 4 {
			t.Errorf("Invalid unmatched row index: %d", r)
		}
	}
}

func TestLinearSumAssignment_RectangularMoreCols(t *testing.T) {
	// 2 rows x 4 cols
	costMatrix := [][]float64{
		{1, 5, 3, 4},
		{2, 3, 6, 2},
	}

	assignments, _, unmatchedCols := LinearSumAssignment(costMatrix, 10.0)

	// At most 2 assignments (limited by number of rows)
	if len(assignments) > 2 {
		t.Errorf("Expected at most 2 assignments, got %d", len(assignments))
	}

	// Should have at least 2 unmatched cols
	if len(unmatchedCols) < 2 {
		t.Errorf("Expected at least 2 unmatched cols, got %d", len(unmatchedCols))
	}

	// Verify unmatched cols are valid
	for _, c := range unmatchedCols {
		if c < 0 || c >= 4 {
			t.Errorf("Invalid unmatched col index: %d", c)
		}
	}
}

func TestLinearSumAssignment_EmptyMatrix(t *testing.T) {
	costMatrix := [][]float64{}

	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 10.0)

	if assignments != nil {
		t.Errorf("Expected nil assignments for empty matrix, got %v", assignments)
	}
	if unmatchedRows != nil {
		t.Errorf("Expected nil unmatchedRows for empty matrix, got %v", unmatchedRows)
	}
	if unmatchedCols != nil {
		t.Errorf("Expected nil unmatchedCols for empty matrix, got %v", unmatchedCols)
	}
}

func TestLinearSumAssignment_EmptyColumns(t *testing.T) {
	// 3 rows but 0 cols
	costMatrix := [][]float64{
		{},
		{},
		{},
	}

	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 10.0)

	// All rows should be unmatched
	if len(unmatchedRows) != 3 {
		t.Errorf("Expected 3 unmatched rows, got %d", len(unmatchedRows))
	}

	// No assignments
	if assignments != nil {
		t.Errorf("Expected nil assignments, got %v", assignments)
	}

	// No unmatched cols (there are no cols)
	if unmatchedCols != nil {
		t.Errorf("Expected nil unmatchedCols, got %v", unmatchedCols)
	}
}

func TestLinearSumAssignment_AllRejectedByThreshold(t *testing.T) {
	costMatrix := [][]float64{
		{10, 11, 12},
		{13, 14, 15},
		{16, 17, 18},
	}

	// maxCost is very low, all assignments should be rejected
	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 5.0)

	// All rows and cols should be unmatched
	if len(unmatchedRows) != 3 {
		t.Errorf("Expected 3 unmatched rows, got %d", len(unmatchedRows))
	}
	if len(unmatchedCols) != 3 {
		t.Errorf("Expected 3 unmatched cols, got %d", len(unmatchedCols))
	}
	if len(assignments) != 0 {
		t.Errorf("Expected 0 assignments, got %d", len(assignments))
	}
}

func TestLinearSumAssignment_OptimalMatching(t *testing.T) {
	// This cost matrix has an obvious optimal matching:
	// Row 0 -> Col 0 (cost 1)
	// Row 1 -> Col 1 (cost 1)
	// Row 2 -> Col 2 (cost 1)
	// Total cost: 3
	costMatrix := [][]float64{
		{1, 10, 10},
		{10, 1, 10},
		{10, 10, 1},
	}

	assignments, _, _ := LinearSumAssignment(costMatrix, 10.0)

	// Should have 3 assignments
	if len(assignments) != 3 {
		t.Errorf("Expected 3 assignments, got %d", len(assignments))
	}

	// Verify we got the optimal assignments (diagonal)
	totalCost := 0.0
	for _, a := range assignments {
		totalCost += costMatrix[a.RowIdx][a.ColIdx]
	}

	if totalCost != 3.0 {
		t.Errorf("Expected total cost 3.0, got %v", totalCost)
	}
}

func TestLinearSumAssignment_SingleElement(t *testing.T) {
	costMatrix := [][]float64{
		{5},
	}

	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 10.0)

	// Should have 1 assignment
	if len(assignments) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignments))
	}

	if len(unmatchedRows) != 0 || len(unmatchedCols) != 0 {
		t.Errorf("Expected no unmatched, got %d rows and %d cols", len(unmatchedRows), len(unmatchedCols))
	}

	if assignments[0].RowIdx != 0 || assignments[0].ColIdx != 0 {
		t.Errorf("Expected assignment (0, 0), got (%d, %d)", assignments[0].RowIdx, assignments[0].ColIdx)
	}
}

func TestLinearSumAssignment_PartialMatching(t *testing.T) {
	// Some low costs, some high costs
	costMatrix := [][]float64{
		{1, 100, 100},
		{100, 2, 100},
		{100, 100, 100},
	}

	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 50.0)

	// Should have 2 assignments (row 0->col 0, row 1->col 1)
	// Row 2 and col 2 should be unmatched
	if len(assignments) != 2 {
		t.Errorf("Expected 2 assignments, got %d", len(assignments))
	}

	if len(unmatchedRows) != 1 || len(unmatchedCols) != 1 {
		t.Errorf("Expected 1 unmatched row and 1 unmatched col, got %d rows and %d cols",
			len(unmatchedRows), len(unmatchedCols))
	}

	// Verify row 2 and col 2 are unmatched
	if unmatchedRows[0] != 2 {
		t.Errorf("Expected unmatched row 2, got %d", unmatchedRows[0])
	}
	if unmatchedCols[0] != 2 {
		t.Errorf("Expected unmatched col 2, got %d", unmatchedCols[0])
	}
}

func TestLinearSumAssignment_ZeroCosts(t *testing.T) {
	// All zero costs
	costMatrix := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	assignments, unmatchedRows, unmatchedCols := LinearSumAssignment(costMatrix, 1.0)

	// Should have 3 assignments (any matching is optimal with zero costs)
	if len(assignments) != 3 {
		t.Errorf("Expected 3 assignments, got %d", len(assignments))
	}

	if len(unmatchedRows) != 0 || len(unmatchedCols) != 0 {
		t.Errorf("Expected no unmatched, got %d rows and %d cols", len(unmatchedRows), len(unmatchedCols))
	}
}

func TestLinearSumAssignment_max_helper(t *testing.T) {
	// Test the max helper function
	if max(5, 3) != 5 {
		t.Errorf("max(5, 3) should be 5")
	}
	if max(2, 7) != 7 {
		t.Errorf("max(2, 7) should be 7")
	}
	if max(4, 4) != 4 {
		t.Errorf("max(4, 4) should be 4")
	}
}
