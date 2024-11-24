package bits

import (
	"reflect"
	"testing"
)

func TestSetBit(t *testing.T) {
	bf := NewBitField(64)

	bf.SetBit(5)
	if !bf.IsSet(5) {
		t.Errorf("Expected bit at position 5 to be set")
	}

	bf.SetBit(63)
	if !bf.IsSet(63) {
		t.Errorf("Expected bit at position 63 to be set")
	}

	// Check that other bits are still unset
	if bf.IsSet(0) {
		t.Errorf("Expected bit at position 0 to be unset")
	}
	if bf.IsSet(6) {
		t.Errorf("Expected bit at position 6 to be unset")
	}
}

func TestClearBit(t *testing.T) {
	bf := NewBitField(64)

	bf.SetBit(10)
	bf.ClearBit(10)
	if bf.IsSet(10) {
		t.Errorf("Expected bit at position 10 to be cleared")
	}

	// Clearing an already cleared bit should not change the state
	bf.ClearBit(10)
	if bf.IsSet(10) {
		t.Errorf("Expected bit at position 10 to remain cleared")
	}
}

func TestIsSet(t *testing.T) {
	bf := NewBitField(64)

	// Initially all bits should be unset
	if bf.IsSet(20) {
		t.Errorf("Expected bit at position 20 to be unset")
	}

	bf.SetBit(20)
	if !bf.IsSet(20) {
		t.Errorf("Expected bit at position 20 to be set")
	}
}

func TestAllocateNextAvailableBits(t *testing.T) {
	// Test case 1: Normal operation with enough bits available
	bf := NewBitField(10)
	bf.SetBit(1)
	bf.SetBit(3)
	allocated, err := bf.AllocateNextAvailableBits(3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []int{0, 2, 4}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated bits %v, got %v", expected, allocated)
	}

	// Test case 2: Insufficient available bits
	_, err = bf.AllocateNextAvailableBits(8) // 8 bits requested when only 7 are left
	if err == nil {
		t.Fatal("Expected error due to insufficient available bits")
	}

	// Test case 3: Request for only one bit
	bf.ClearBit(5)
	allocated, err = bf.AllocateNextAvailableBits(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{5}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated bit %v, got %v", expected, allocated)
	}

	// Test case 4: Edge case with bitfield size of 1
	bfSmall := NewBitField(1)
	allocated, err = bfSmall.AllocateNextAvailableBits(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{0}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated bit %v, got %v", expected, allocated)
	}
	if !bfSmall.IsSet(0) {
		t.Error("Expected bit 0 to be set in bitfield of size 1")
	}

	// Test case 5: Out of range request for more bits than the size of bitfield
	_, err = bfSmall.AllocateNextAvailableBits(2) // Only 1 bit is available in bfSmall
	if err == nil {
		t.Fatal("Expected error due to request exceeding bitfield size")
	}

	// Test case 6: Verify bits are non-consecutive
	bf = NewBitField(10)
	bf.SetBit(2)
	bf.SetBit(5)
	allocated, err = bf.AllocateNextAvailableBits(3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{0, 1, 3}
	if !reflect.DeepEqual(allocated, expected) {
		t.Errorf("Expected allocated bits %v, got %v", expected, allocated)
	}
	if !bf.IsSet(2) || !bf.IsSet(5) {
		t.Errorf("Expected bits at positions 2 and 5 to remain set")
	}
}

func TestAllocateNextAvailableBitsInvalidInput(t *testing.T) {
	bf := NewBitField(64)

	// Test invalid input (e.g., n <= 0 or n > size)
	_, err := bf.AllocateNextAvailableBits(0)
	if err == nil {
		t.Errorf("Expected error for invalid bit count (0)")
	}

	_, err = bf.AllocateNextAvailableBits(-1)
	if err == nil {
		t.Errorf("Expected error for invalid bit count (-1)")
	}

	_, err = bf.AllocateNextAvailableBits(65)
	if err == nil {
		t.Errorf("Expected error for bit count exceeding BitField size")
	}
}

func TestNextAvailableBitsInRange(t *testing.T) {
	// Test case 1: Normal operation within bounds
	bf := NewBitField(10)
	bf.SetBit(1)
	bf.SetBit(3)
	bf.SetBit(7)

	// Find the next 2 available bits in range [0, 5)
	availableBits, err := bf.AllocateAvailableBitsInRange(0, 5, 2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []int{0, 2}
	for i, pos := range availableBits {
		if pos != expected[i] {
			t.Errorf("Expected %d, got %d at index %d", expected[i], pos, i)
		}
	}

	// Test case 2: Range with fewer than required bits
	_, err = bf.AllocateAvailableBitsInRange(0, 5, 4)
	if err == nil {
		t.Fatal("Expected error due to insufficient consecutive bits")
	}

	// Test case 3: Sufficient bits in a later part of the range
	availableBits, err = bf.AllocateAvailableBitsInRange(5, 10, 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected = []int{5, 6, 8}
	for i, pos := range availableBits {
		if pos != expected[i] {
			t.Errorf("Expected %d, got %d at index %d", expected[i], pos, i)
		}
	}

	// Test case 4: Entire range but no sequence found
	bf.SetBit(0)
	bf.SetBit(2)
	bf.SetBit(4)
	_, err = bf.AllocateAvailableBitsInRange(0, 5, 2)
	if err == nil {
		t.Fatal("Expected error due to lack of available consecutive bits")
	}

	// Test case 5: Out-of-bounds range
	_, err = bf.AllocateAvailableBitsInRange(-1, 15, 2)
	if err == nil {
		t.Fatal("Expected error due to out-of-bounds range")
	}

	// Test case 6: Range where `n` is greater than available bits
	_, err = bf.AllocateAvailableBitsInRange(0, 10, 11)
	if err == nil {
		t.Fatal("Expected error due to requesting more bits than available in range")
	}

	// Test case 7: Requesting 1 available bit
	bf.ClearBit(0)
	availableBits, err = bf.AllocateAvailableBitsInRange(0, 10, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(availableBits) != 1 || availableBits[0] != 0 {
		t.Errorf("Expected single available bit at position 0, got %v", availableBits)
	}
}
