package bits

import (
	"reflect"
	"testing"
)

func TestNewPortRange(t *testing.T) {
	// Valid range creation
	portRange, err := NewPortRange(1000, 20)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if portRange.startPort != 1000 {
		t.Errorf("Expected start port 1000, got %d", portRange.startPort)
	}
	if portRange.size != 20 {
		t.Errorf("Expected size 20, got %d", portRange.size)
	}

	// Invalid range creation
	_, err = NewPortRange(-1, 10)
	if err == nil {
		t.Fatalf("Expected error for negative start port, got nil")
	}

	_, err = NewPortRange(1000, 0)
	if err == nil {
		t.Fatalf("Expected error for zero size, got nil")
	}
}

func TestAllocateNextPorts(t *testing.T) {
	// Test case 1: Normal allocation with enough available ports
	pr, err := NewPortRange(8000, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	allocated, err := pr.AllocateNextPorts(3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []int{8000, 8001, 8002}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated ports %v, got %v", expected, allocated)
	}

	// Test case 2: Insufficient available ports
	_, err = pr.AllocateNextPorts(8) // Requesting 8 ports when only 7 are available
	if err == nil {
		t.Fatal("Expected error due to insufficient available ports")
	}

	// Test case 3: Request for only one port
	pr.bitField.ClearBit(5) // Clear bit at index 5 to simulate a free port
	allocated, err = pr.AllocateNextPorts(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{8005}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated port %v, got %v", expected, allocated)
	}

	// Test case 4: Edge case with a range of size 1
	prSingle, err := NewPortRange(8080, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	allocated, err = prSingle.AllocateNextPorts(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{8080}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated port %v, got %v", expected, allocated)
	}
	if !prSingle.bitField.IsSet(0) {
		t.Error("Expected bit 0 to be set in bitfield of size 1")
	}

	// Test case 5: Out of range request for more ports than available
	_, err = prSingle.AllocateNextPorts(2) // Only 1 port is available in prSingle
	if err == nil {
		t.Fatal("Expected error due to request exceeding port range size")
	}

	// Test case 6: Allocate all remaining ports
	pr, err = NewPortRange(8000, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	allocated, err = pr.AllocateNextPorts(5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{8000, 8001, 8002, 8003, 8004}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated ports %v, got %v", expected, allocated)
	}

	// Test case 7: Non-consecutive allocation
	pr, err = NewPortRange(8000, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	pr.bitField.SetBit(1) // Mark port 8001 as allocated
	pr.bitField.SetBit(3) // Mark port 8003 as allocated
	allocated, err = pr.AllocateNextPorts(3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{8000, 8002, 8004}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated ports %v, got %v", expected, allocated)
	}
	if !pr.bitField.IsSet(1) || !pr.bitField.IsSet(3) {
		t.Errorf("Expected bits at positions 1 and 3 to remain set")
	}
}

func TestReleasePorts(t *testing.T) {
	portRange, _ := NewPortRange(1000, 20)

	// Allocate and then release specific ports
	_, _ = portRange.AllocateNextPorts(3) // Allocate first 3 ports
	err := portRange.ReleasePorts([]int{1000, 1002})
	if err != nil {
		t.Fatalf("Unexpected error releasing ports: %v", err)
	}

	// Check if the released ports are now available
	isAllocated, _ := portRange.IsPortAllocated(1000)
	if isAllocated {
		t.Errorf("Expected port 1000 to be released")
	}

	isAllocated, _ = portRange.IsPortAllocated(1002)
	if isAllocated {
		t.Errorf("Expected port 1002 to be released")
	}

	// Verify that port 1001 remains allocated
	isAllocated, _ = portRange.IsPortAllocated(1001)
	if !isAllocated {
		t.Errorf("Expected port 1001 to be allocated")
	}

	// Try releasing a port that is out of range
	err = portRange.ReleasePorts([]int{1050})
	if err == nil {
		t.Fatalf("Expected error when releasing out-of-range port")
	}
}

func TestIsPortAllocated(t *testing.T) {
	portRange, _ := NewPortRange(1000, 20)

	// Initially all ports should be unallocated
	isAllocated, _ := portRange.IsPortAllocated(1005)
	if isAllocated {
		t.Errorf("Expected port 1005 to be unallocated")
	}

	// Allocate a specific port and check
	_ = portRange.SetPortAllocated(1005)
	isAllocated, _ = portRange.IsPortAllocated(1005)
	if !isAllocated {
		t.Errorf("Expected port 1005 to be allocated")
	}

	// Check for a port out of range
	_, err := portRange.IsPortAllocated(1050)
	if err == nil {
		t.Fatalf("Expected error for checking out-of-range port")
	}
}

func TestSetPortAllocated(t *testing.T) {
	portRange, _ := NewPortRange(1000, 20)

	// Allocate a specific port
	err := portRange.SetPortAllocated(1003)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if the specific port is allocated
	isAllocated, _ := portRange.IsPortAllocated(1003)
	if !isAllocated {
		t.Errorf("Expected port 1003 to be allocated")
	}

	// Try to allocate a port out of range
	err = portRange.SetPortAllocated(1050)
	if err == nil {
		t.Fatalf("Expected error for out-of-range port allocation")
	}
}

func TestAllocateAndReleaseCombination(t *testing.T) {
	portRange, _ := NewPortRange(1000, 20)

	// Allocate multiple ports and release some of them
	_, _ = portRange.AllocateNextPorts(5) // Allocate first 5 ports
	err := portRange.ReleasePorts([]int{1000, 1001, 1004})
	if err != nil {
		t.Fatalf("Unexpected error releasing ports: %v", err)
	}

	// Allocate again and verify that released ports are reused
	allocatedPorts, err := portRange.AllocateNextPorts(2)
	if err != nil {
		t.Fatalf("Unexpected error allocating ports: %v", err)
	}
	expectedPorts := []int{1000, 1001}
	for i, port := range allocatedPorts {
		if port != expectedPorts[i] {
			t.Errorf("Expected port %d, got %d", expectedPorts[i], port)
		}
	}
}
