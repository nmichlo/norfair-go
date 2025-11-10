// Copyright 2025 Nathan Michlo
// SPDX-License-Identifier: MIT
//
// This file contains a Go port of filterpy.kalman.KalmanFilter
// Original source: https://github.com/rlabbe/filterpy/blob/master/filterpy/kalman/kalman_filter.py
//
// Original Copyright (c) 2015 Roger R. Labbe Jr.
// Original License: MIT
//
// See LICENSE file in this directory and THIRD_PARTY_LICENSES.md in repository root.

package filterpy

import (
	"gonum.org/v1/gonum/mat"
)

// KalmanFilter implements a full Kalman filter matching filterpy's KalmanFilter.
//
// This is a Go port of the filterpy.kalman.KalmanFilter class which provides
// a complete Kalman filter implementation for state estimation.
//
// Reference: https://github.com/rlabbe/filterpy/blob/master/filterpy/kalman/kalman_filter.py
type KalmanFilter struct {
	dimX int        // State dimension (2 * dimZ for position + velocity)
	dimZ int        // Measurement dimension
	x    *mat.Dense // State vector (dimX, 1)
	P    *mat.Dense // Covariance matrix (dimX, dimX)
	F    *mat.Dense // State transition matrix (dimX, dimX)
	H    *mat.Dense // Measurement matrix (dimZ, dimX)
	R    *mat.Dense // Measurement noise covariance (dimZ, dimZ)
	Q    *mat.Dense // Process noise covariance (dimX, dimX)

	// Working matrices for computation (pre-allocated)
	xPrior *mat.Dense
	pPrior *mat.Dense
}

// NewKalmanFilter creates a new Kalman filter.
//
// Parameters:
//   - dimX: State vector dimension
//   - dimZ: Measurement vector dimension
//
// The filter is initialized with identity matrices and must be configured
// before use by setting F, H, Q, R, and initial state x and covariance P.
func NewKalmanFilter(dimX, dimZ int) *KalmanFilter {
	kf := &KalmanFilter{
		dimX:   dimX,
		dimZ:   dimZ,
		x:      mat.NewDense(dimX, 1, nil),
		P:      mat.NewDense(dimX, dimX, nil),
		F:      mat.NewDense(dimX, dimX, nil),
		H:      mat.NewDense(dimZ, dimX, nil),
		R:      mat.NewDense(dimZ, dimZ, nil),
		Q:      mat.NewDense(dimX, dimX, nil),
		xPrior: mat.NewDense(dimX, 1, nil),
		pPrior: mat.NewDense(dimX, dimX, nil),
	}

	// Initialize with identity matrices
	for i := 0; i < dimX; i++ {
		kf.F.Set(i, i, 1.0)
		kf.P.Set(i, i, 1.0)
		kf.Q.Set(i, i, 1.0)
	}
	for i := 0; i < dimZ; i++ {
		kf.H.Set(i, i, 1.0)
		kf.R.Set(i, i, 1.0)
	}

	return kf
}

// Predict performs the prediction step of the Kalman filter.
//
// Updates state: x = F @ x
// Updates covariance: P = F @ P @ F^T + Q
func (kf *KalmanFilter) Predict() {
	// x = F @ x
	kf.xPrior.Mul(kf.F, kf.x)
	kf.x.Copy(kf.xPrior)

	// P = F @ P @ F^T + Q
	var temp mat.Dense
	temp.Mul(kf.F, kf.P)
	kf.pPrior.Mul(&temp, kf.F.T())
	kf.P.Add(kf.pPrior, kf.Q)
}

// Update performs the update step of the Kalman filter.
//
// Parameters:
//   - z: Measurement vector (dimZ, 1)
//   - R: Measurement noise covariance (optional, uses default if nil)
//   - H: Measurement matrix (optional, uses default if nil)
//
// The update incorporates the measurement into the state estimate using
// the Kalman gain to optimally blend prediction and measurement.
func (kf *KalmanFilter) Update(z *mat.Dense, R, H *mat.Dense) {
	// Use provided R and H, or defaults
	rMatrix := kf.R
	if R != nil {
		rMatrix = R
	}
	hMatrix := kf.H
	if H != nil {
		hMatrix = H
	}

	// y = z - H @ x (innovation)
	var hx mat.Dense
	hx.Mul(hMatrix, kf.x)
	var y mat.Dense
	y.Sub(z, &hx)

	// S = H @ P @ H^T + R (innovation covariance)
	var temp1 mat.Dense
	temp1.Mul(hMatrix, kf.P)
	var s mat.Dense
	s.Mul(&temp1, hMatrix.T())
	s.Add(&s, rMatrix)

	// K = P @ H^T @ S^-1 (Kalman gain)
	var sInv mat.Dense
	err := sInv.Inverse(&s)
	if err != nil {
		// If S is singular, skip update
		return
	}
	var temp2 mat.Dense
	temp2.Mul(kf.P, hMatrix.T())
	var k mat.Dense
	k.Mul(&temp2, &sInv)

	// x = x + K @ y
	var kY mat.Dense
	kY.Mul(&k, &y)
	kf.x.Add(kf.x, &kY)

	// P = (I - K @ H) @ P (Joseph form for numerical stability)
	dimX := kf.dimX
	identity := mat.NewDense(dimX, dimX, nil)
	for i := 0; i < dimX; i++ {
		identity.Set(i, i, 1.0)
	}
	var kH mat.Dense
	kH.Mul(&k, hMatrix)
	var iMinusKH mat.Dense
	iMinusKH.Sub(identity, &kH)
	var newP mat.Dense
	newP.Mul(&iMinusKH, kf.P)
	kf.P.Copy(&newP)
}

// GetState returns the current state estimate.
func (kf *KalmanFilter) GetState() *mat.Dense {
	return kf.x
}

// SetState sets the current state estimate.
func (kf *KalmanFilter) SetState(x *mat.Dense) {
	kf.x.Copy(x)
}

// GetCovariance returns the current state covariance matrix.
func (kf *KalmanFilter) GetCovariance() *mat.Dense {
	return kf.P
}

// SetCovariance sets the current state covariance matrix.
func (kf *KalmanFilter) SetCovariance(P *mat.Dense) {
	kf.P.Copy(P)
}

// GetDimX returns the state dimension.
func (kf *KalmanFilter) GetDimX() int {
	return kf.dimX
}

// GetDimZ returns the measurement dimension.
func (kf *KalmanFilter) GetDimZ() int {
	return kf.dimZ
}

// GetF returns the state transition matrix.
func (kf *KalmanFilter) GetF() *mat.Dense {
	return kf.F
}

// GetH returns the measurement matrix.
func (kf *KalmanFilter) GetH() *mat.Dense {
	return kf.H
}

// GetR returns the measurement noise covariance matrix.
func (kf *KalmanFilter) GetR() *mat.Dense {
	return kf.R
}

// GetQ returns the process noise covariance matrix.
func (kf *KalmanFilter) GetQ() *mat.Dense {
	return kf.Q
}

// GetP returns the state covariance matrix (alias for GetCovariance).
func (kf *KalmanFilter) GetP() *mat.Dense {
	return kf.P
}

// GetX returns the state vector (alias for GetState).
func (kf *KalmanFilter) GetX() *mat.Dense {
	return kf.x
}
