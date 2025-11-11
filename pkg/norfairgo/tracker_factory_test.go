package norfairgo

import (
	"sync"
	"testing"
)

// =============================================================================
// Test Basic ID Generation
// =============================================================================

// Python equivalent: tools/validate_tracker_ids/main.py::test_basic_initializing_ids()
//
//	from norfair.tracker import _TrackedObjectFactory
//
//	def test_basic_initializing_ids():
//	    factory = _TrackedObjectFactory()
//	    ids = []
//	    for i in range(5):
//	        id_val = factory.get_initializing_id()
//	        ids.append(id_val)
//	    expected = [1, 2, 3, 4, 5]
//	    assert ids == expected
//
// Test cases match tools/validate_tracker_ids/main.py::test_basic_initializing_ids()
func TestTrackedObjectFactory_GetInitializingID(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	factory := NewTrackedObjectFactory()

	// Test that initializing IDs start at 1 and increment
	id1 := factory.GetInitializingID()
	if id1 != 1 {
		t.Errorf("First initializing ID should be 1, got %d", id1)
	}

	id2 := factory.GetInitializingID()
	if id2 != 2 {
		t.Errorf("Second initializing ID should be 2, got %d", id2)
	}

	id3 := factory.GetInitializingID()
	if id3 != 3 {
		t.Errorf("Third initializing ID should be 3, got %d", id3)
	}

	// Verify counter value
	if factory.InitializingCount() != 3 {
		t.Errorf("InitializingCount should be 3, got %d", factory.InitializingCount())
	}

	// Permanent count should still be 0
	if factory.Count() != 0 {
		t.Errorf("Count should be 0, got %d", factory.Count())
	}
}

// Python equivalent: tools/validate_tracker_ids/main.py::test_basic_permanent_ids()
//
//	from norfair.tracker import _TrackedObjectFactory
//
//	def test_basic_permanent_ids():
//	    _TrackedObjectFactory.global_count = 0
//	    factory = _TrackedObjectFactory()
//	    instance_ids = []
//	    global_ids = []
//	    for i in range(5):
//	        inst_id, glob_id = factory.get_ids()
//	        instance_ids.append(inst_id)
//	        global_ids.append(glob_id)
//	    expected = [1, 2, 3, 4, 5]
//	    assert instance_ids == expected
//	    assert global_ids == expected
//
// Test cases match tools/validate_tracker_ids/main.py::test_basic_permanent_ids()
func TestTrackedObjectFactory_GetIDs(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	factory := NewTrackedObjectFactory()

	// Test that permanent IDs start at 1 and increment
	id1, globalID1 := factory.GetIDs()
	if id1 != 1 {
		t.Errorf("First instance ID should be 1, got %d", id1)
	}
	if globalID1 != 1 {
		t.Errorf("First global ID should be 1, got %d", globalID1)
	}

	id2, globalID2 := factory.GetIDs()
	if id2 != 2 {
		t.Errorf("Second instance ID should be 2, got %d", id2)
	}
	if globalID2 != 2 {
		t.Errorf("Second global ID should be 2, got %d", globalID2)
	}

	id3, globalID3 := factory.GetIDs()
	if id3 != 3 {
		t.Errorf("Third instance ID should be 3, got %d", id3)
	}
	if globalID3 != 3 {
		t.Errorf("Third global ID should be 3, got %d", globalID3)
	}

	// Verify counter values
	if factory.Count() != 3 {
		t.Errorf("Count should be 3, got %d", factory.Count())
	}
	if GlobalCount() != 3 {
		t.Errorf("GlobalCount should be 3, got %d", GlobalCount())
	}
}

// =============================================================================
// Test Global ID Uniqueness Across Multiple Factories
// =============================================================================

// Python equivalent: tools/validate_tracker_ids/main.py::test_global_id_uniqueness()
//
//	from norfair.tracker import _TrackedObjectFactory
//
//	def test_global_id_uniqueness():
//	    _TrackedObjectFactory.global_count = 0
//	    factory1 = _TrackedObjectFactory()
//	    factory2 = _TrackedObjectFactory()
//	    # Factory 1 creates 2 objects
//	    id1_1, gid1_1 = factory1.get_ids()
//	    id1_2, gid1_2 = factory1.get_ids()
//	    # Factory 2 creates 2 objects
//	    id2_1, gid2_1 = factory2.get_ids()
//	    id2_2, gid2_2 = factory2.get_ids()
//	    # Verify instance IDs restart for each factory
//	    assert id1_1 == 1 and id1_2 == 2
//	    assert id2_1 == 1 and id2_2 == 2
//	    # Verify global IDs are unique across factories
//	    assert gid1_1 == 1 and gid1_2 == 2 and gid2_1 == 3 and gid2_2 == 4
//
// Test cases match tools/validate_tracker_ids/main.py::test_global_id_uniqueness()
func TestTrackedObjectFactory_GlobalIDUniqueness(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	// Create two separate factory instances (simulating two separate trackers)
	factory1 := NewTrackedObjectFactory()
	factory2 := NewTrackedObjectFactory()

	// First factory creates an object
	id1, globalID1 := factory1.GetIDs()
	if id1 != 1 {
		t.Errorf("Factory1 first instance ID should be 1, got %d", id1)
	}
	if globalID1 != 1 {
		t.Errorf("Factory1 first global ID should be 1, got %d", globalID1)
	}

	// Second factory creates an object
	id2, globalID2 := factory2.GetIDs()
	if id2 != 1 {
		t.Errorf("Factory2 first instance ID should be 1, got %d", id2)
	}
	if globalID2 != 2 {
		t.Errorf("Factory2 first global ID should be 2, got %d", globalID2)
	}

	// First factory creates another object
	id3, globalID3 := factory1.GetIDs()
	if id3 != 2 {
		t.Errorf("Factory1 second instance ID should be 2, got %d", id3)
	}
	if globalID3 != 3 {
		t.Errorf("Factory1 second global ID should be 3, got %d", globalID3)
	}

	// Verify individual factory counts
	if factory1.Count() != 2 {
		t.Errorf("Factory1 count should be 2, got %d", factory1.Count())
	}
	if factory2.Count() != 1 {
		t.Errorf("Factory2 count should be 1, got %d", factory2.Count())
	}

	// Verify global count
	if GlobalCount() != 3 {
		t.Errorf("GlobalCount should be 3, got %d", GlobalCount())
	}
}

// =============================================================================
// Test Initializing vs Permanent ID Independence
// =============================================================================

// Python equivalent: tools/validate_tracker_ids/main.py::test_initializing_vs_permanent_independence()
//
//	from norfair.tracker import _TrackedObjectFactory
//
//	def test_initializing_vs_permanent_independence():
//	    _TrackedObjectFactory.global_count = 0
//	    factory = _TrackedObjectFactory()
//	    # Get 3 initializing IDs
//	    init_ids = []
//	    for i in range(3):
//	        init_id = factory.get_initializing_id()
//	        init_ids.append(init_id)
//	    # Now get 2 permanent IDs
//	    perm_ids = []
//	    for i in range(2):
//	        inst_id, glob_id = factory.get_ids()
//	        perm_ids.append((inst_id, glob_id))
//	    # Verify initializing IDs are [1,2,3]
//	    assert init_ids == [1, 2, 3]
//	    # Verify permanent IDs start from 1 independently
//	    assert perm_ids == [(1, 1), (2, 2)]
//
// Test cases match tools/validate_tracker_ids/main.py::test_initializing_vs_permanent_independence()
func TestTrackedObjectFactory_InitializingVsPermanentIDs(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	factory := NewTrackedObjectFactory()

	// Get some initializing IDs first
	initID1 := factory.GetInitializingID()
	initID2 := factory.GetInitializingID()
	initID3 := factory.GetInitializingID()

	if initID1 != 1 || initID2 != 2 || initID3 != 3 {
		t.Errorf("Initializing IDs should be 1,2,3, got %d,%d,%d", initID1, initID2, initID3)
	}

	// Verify initializing count is 3, but permanent count is still 0
	if factory.InitializingCount() != 3 {
		t.Errorf("InitializingCount should be 3, got %d", factory.InitializingCount())
	}
	if factory.Count() != 0 {
		t.Errorf("Count should be 0, got %d", factory.Count())
	}

	// Now get permanent IDs
	id1, globalID1 := factory.GetIDs()
	id2, globalID2 := factory.GetIDs()

	// Permanent IDs should start from 1, independent of initializing IDs
	if id1 != 1 || globalID1 != 1 {
		t.Errorf("First permanent IDs should be (1,1), got (%d,%d)", id1, globalID1)
	}
	if id2 != 2 || globalID2 != 2 {
		t.Errorf("Second permanent IDs should be (2,2), got (%d,%d)", id2, globalID2)
	}

	// Verify both counters are independent
	if factory.InitializingCount() != 3 {
		t.Errorf("InitializingCount should still be 3, got %d", factory.InitializingCount())
	}
	if factory.Count() != 2 {
		t.Errorf("Count should be 2, got %d", factory.Count())
	}
}

// Python equivalent: tools/validate_tracker_ids/main.py::test_mixed_sequence()
//
//	from norfair.tracker import _TrackedObjectFactory
//
//	def test_mixed_sequence():
//	    _TrackedObjectFactory.global_count = 0
//	    factory = _TrackedObjectFactory()
//	    sequence = []
//	    # Initializing ID
//	    init1 = factory.get_initializing_id()
//	    sequence.append(('init', init1))
//	    # Permanent ID
//	    perm1_inst, perm1_glob = factory.get_ids()
//	    sequence.append(('perm', perm1_inst, perm1_glob))
//	    # Initializing ID
//	    init2 = factory.get_initializing_id()
//	    sequence.append(('init', init2))
//	    # Initializing ID
//	    init3 = factory.get_initializing_id()
//	    sequence.append(('init', init3))
//	    # Permanent ID
//	    perm2_inst, perm2_glob = factory.get_ids()
//	    sequence.append(('perm', perm2_inst, perm2_glob))
//	    # Verify the sequence
//	    expected = [('init', 1), ('perm', 1, 1), ('init', 2), ('init', 3), ('perm', 2, 2)]
//	    assert sequence == expected
//
// Test cases match tools/validate_tracker_ids/main.py::test_mixed_sequence()
func TestTrackedObjectFactory_MixedSequence(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	factory := NewTrackedObjectFactory()

	// Build sequence of ID operations exactly as Python test

	// 1. Initializing ID
	init1 := factory.GetInitializingID()
	if init1 != 1 {
		t.Errorf("Step 1: Initializing ID should be 1, got %d", init1)
	}

	// 2. Permanent ID
	perm1Inst, perm1Glob := factory.GetIDs()
	if perm1Inst != 1 || perm1Glob != 1 {
		t.Errorf("Step 2: Permanent IDs should be (1, 1), got (%d, %d)", perm1Inst, perm1Glob)
	}

	// 3. Initializing ID
	init2 := factory.GetInitializingID()
	if init2 != 2 {
		t.Errorf("Step 3: Initializing ID should be 2, got %d", init2)
	}

	// 4. Initializing ID
	init3 := factory.GetInitializingID()
	if init3 != 3 {
		t.Errorf("Step 4: Initializing ID should be 3, got %d", init3)
	}

	// 5. Permanent ID
	perm2Inst, perm2Glob := factory.GetIDs()
	if perm2Inst != 2 || perm2Glob != 2 {
		t.Errorf("Step 5: Permanent IDs should be (2, 2), got (%d, %d)", perm2Inst, perm2Glob)
	}

	// Verify final state matches expected sequence
	// Expected sequence: init=[1,2,3], perm=[(1,1),(2,2)]
	if factory.InitializingCount() != 3 {
		t.Errorf("Expected 3 initializing IDs generated, got %d", factory.InitializingCount())
	}
	if factory.Count() != 2 {
		t.Errorf("Expected 2 permanent IDs generated, got %d", factory.Count())
	}
	if GlobalCount() != 2 {
		t.Errorf("Expected global count 2, got %d", GlobalCount())
	}
}

// =============================================================================
// Test Thread Safety
// =============================================================================

func TestTrackedObjectFactory_ConcurrentInitializingIDs(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	factory := NewTrackedObjectFactory()
	numGoroutines := 100
	idsPerGoroutine := 100

	var wg sync.WaitGroup
	idsChan := make(chan int, numGoroutines*idsPerGoroutine)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := factory.GetInitializingID()
				idsChan <- id
			}
		}()
	}

	wg.Wait()
	close(idsChan)

	// Collect all IDs
	ids := make(map[int]bool)
	for id := range idsChan {
		if ids[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		ids[id] = true
	}

	// Verify we got the expected number of unique IDs
	expectedTotal := numGoroutines * idsPerGoroutine
	if len(ids) != expectedTotal {
		t.Errorf("Expected %d unique IDs, got %d", expectedTotal, len(ids))
	}

	// Verify final counter value
	if factory.InitializingCount() != expectedTotal {
		t.Errorf("InitializingCount should be %d, got %d", expectedTotal, factory.InitializingCount())
	}
}

func TestTrackedObjectFactory_ConcurrentPermanentIDs(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	factory := NewTrackedObjectFactory()
	numGoroutines := 100
	idsPerGoroutine := 100

	var wg sync.WaitGroup
	type IDPair struct {
		instanceID int
		globalID   int
	}
	idsChan := make(chan IDPair, numGoroutines*idsPerGoroutine)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, globalID := factory.GetIDs()
				idsChan <- IDPair{id, globalID}
			}
		}()
	}

	wg.Wait()
	close(idsChan)

	// Collect all IDs
	instanceIDs := make(map[int]bool)
	globalIDs := make(map[int]bool)
	for pair := range idsChan {
		if instanceIDs[pair.instanceID] {
			t.Errorf("Duplicate instance ID found: %d", pair.instanceID)
		}
		if globalIDs[pair.globalID] {
			t.Errorf("Duplicate global ID found: %d", pair.globalID)
		}
		instanceIDs[pair.instanceID] = true
		globalIDs[pair.globalID] = true
	}

	// Verify we got the expected number of unique IDs
	expectedTotal := numGoroutines * idsPerGoroutine
	if len(instanceIDs) != expectedTotal {
		t.Errorf("Expected %d unique instance IDs, got %d", expectedTotal, len(instanceIDs))
	}
	if len(globalIDs) != expectedTotal {
		t.Errorf("Expected %d unique global IDs, got %d", expectedTotal, len(globalIDs))
	}

	// Verify final counter values
	if factory.Count() != expectedTotal {
		t.Errorf("Count should be %d, got %d", expectedTotal, factory.Count())
	}
	if GlobalCount() != expectedTotal {
		t.Errorf("GlobalCount should be %d, got %d", expectedTotal, GlobalCount())
	}
}

func TestTrackedObjectFactory_ConcurrentMultipleFactories(t *testing.T) {
	// Reset global counter before test
	ResetGlobalCount()

	numFactories := 10
	idsPerFactory := 100

	var wg sync.WaitGroup
	globalIDsChan := make(chan int, numFactories*idsPerFactory)

	// Launch concurrent factories
	for i := 0; i < numFactories; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			factory := NewTrackedObjectFactory()
			for j := 0; j < idsPerFactory; j++ {
				_, globalID := factory.GetIDs()
				globalIDsChan <- globalID
			}
		}()
	}

	wg.Wait()
	close(globalIDsChan)

	// Collect all global IDs
	globalIDs := make(map[int]bool)
	for globalID := range globalIDsChan {
		if globalIDs[globalID] {
			t.Errorf("Duplicate global ID found: %d", globalID)
		}
		globalIDs[globalID] = true
	}

	// Verify we got the expected number of unique global IDs
	expectedTotal := numFactories * idsPerFactory
	if len(globalIDs) != expectedTotal {
		t.Errorf("Expected %d unique global IDs, got %d", expectedTotal, len(globalIDs))
	}

	// Verify final global counter value
	if GlobalCount() != expectedTotal {
		t.Errorf("GlobalCount should be %d, got %d", expectedTotal, GlobalCount())
	}
}
