// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: BSD-3-Clause
//
// This file contains a Go port of numpy.linspace
// Original source: https://github.com/numpy/numpy/blob/main/numpy/core/function_base.py
//
// Original Copyright (c) 2005-2024, NumPy Developers
// Original License: BSD-3-Clause
//
// See LICENSE file in this directory and THIRD_PARTY_LICENSES.md in repository root.

package numpy

// Linspace generates n evenly spaced values between start and end (inclusive).
//
// This is a Go port of numpy.linspace which returns evenly spaced numbers over
// a specified interval.
//
// Parameters:
//   - start: Starting value of the sequence
//   - end: End value of the sequence
//   - n: Number of samples to generate (must be >= 2)
//
// Returns:
//   - Slice of n evenly spaced float64 values
//
// Reference: https://github.com/numpy/numpy/blob/main/numpy/core/function_base.py#L23
func Linspace(start, end float64, n int) []float64 {
	if n < 2 {
		if n == 1 {
			return []float64{start}
		}
		return []float64{}
	}

	result := make([]float64, n)
	step := (end - start) / float64(n-1)

	for i := 0; i < n; i++ {
		result[i] = start + float64(i)*step
	}

	// Ensure endpoint is exact (avoid floating point drift)
	result[n-1] = end

	return result
}
