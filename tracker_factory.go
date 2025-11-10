package norfairgo

import (
	"sync"
)

// =============================================================================
// Package-level Global Counter
// =============================================================================

var (
	// globalCount is the global counter shared across all TrackedObjectFactory instances
	globalCount int

	// globalCountMutex protects globalCount for thread-safe increments
	globalCountMutex sync.Mutex
)

// =============================================================================
// TrackedObjectFactory - Manages ID generation for TrackedObjects
// =============================================================================

// TrackedObjectFactory is responsible for generating unique IDs for tracked objects.
// It maintains both instance-level IDs (unique within a tracker) and global IDs
// (unique across all trackers).
type TrackedObjectFactory struct {
	// count is the instance-level counter for permanent IDs
	count int

	// initializingCount is the counter for temporary IDs during object initialization
	initializingCount int

	// mu protects the instance-level counters
	mu sync.Mutex
}

// NewTrackedObjectFactory creates a new TrackedObjectFactory instance.
func NewTrackedObjectFactory() *TrackedObjectFactory {
	return &TrackedObjectFactory{
		count:             0,
		initializingCount: 0,
	}
}

// GetInitializingID returns a temporary ID for objects during initialization phase.
// This ID is assigned immediately when a tracked object is created, before it
// passes the initialization delay threshold.
//
// Returns:
//   - int: A unique initializing ID (starts at 1, increments with each call)
func (f *TrackedObjectFactory) GetInitializingID() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.initializingCount++
	return f.initializingCount
}

// GetIDs returns both an instance-level ID and a global ID for a tracked object.
// This method should only be called when an object transitions from the
// initialization phase to the active tracking phase (i.e., when hit_counter
// exceeds initialization_delay).
//
// Returns:
//   - id: Instance-level ID (unique within this tracker)
//   - globalID: Global ID (unique across all trackers in the application)
func (f *TrackedObjectFactory) GetIDs() (id int, globalID int) {
	// Lock instance counter
	f.mu.Lock()
	f.count++
	instanceID := f.count
	f.mu.Unlock()

	// Lock global counter
	globalCountMutex.Lock()
	globalCount++
	gID := globalCount
	globalCountMutex.Unlock()

	return instanceID, gID
}

// Count returns the current instance-level counter value.
// This represents the total number of objects that have been fully initialized
// by this factory instance.
func (f *TrackedObjectFactory) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.count
}

// InitializingCount returns the current initializing counter value.
// This represents the total number of objects that have been created
// (including those still in initialization phase).
func (f *TrackedObjectFactory) InitializingCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.initializingCount
}

// GlobalCount returns the current global counter value.
// This represents the total number of objects that have been fully initialized
// across ALL factory instances.
func GlobalCount() int {
	globalCountMutex.Lock()
	defer globalCountMutex.Unlock()
	return globalCount
}

// ResetGlobalCount resets the global counter to zero.
// This should typically only be used in tests.
func ResetGlobalCount() {
	globalCountMutex.Lock()
	globalCount = 0
	globalCountMutex.Unlock()
}
