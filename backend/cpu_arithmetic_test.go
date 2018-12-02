package backend

import "testing"

func TestAddnHalfCarry(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0x8
	c.Addn(0x8, false)

	if c.reg[A] != 0x8*2 {
		t.Errorf("Test -> A: 0x8 + n: 0x8 -> Result incorrect, got: %d, want: %d.", c.reg[A], 2*0x8)
	}
	if !c.IsFlagSet(HFlag) {
		t.Errorf("Test -> A: 0x8 + n: 0x8 -> Half carry flag should be set")
	}
}

func TestAddnCarry(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0xFF
	c.Addn(0x1, false)

	if c.reg[A] != 0 {
		t.Errorf("Test -> A: 0xFF + n: 0x1 -> Result incorrect, got: %d, want: %d.", c.reg[A], 0)
	}
	if !c.IsFlagSet(CFlag) {
		t.Errorf("Test -> A: 0xFF + n: 0x1 -> Carry flag should be set")
	}
	if !c.IsFlagSet(ZFlag) {
		t.Errorf("Test -> A: 0xFF + n: 0x1 -> Zero flag should be set")
	}
}

func TestAddnCarryIn(t *testing.T) {
	c := NewTestCPU()

	c.SetFlag(CFlag)

	c.reg[A] = 0xFF
	c.Addn(0x0, true)

	if c.reg[A] != 0 {
		t.Errorf("Test -> A: 0xFF + n: 0x1 -> Result incorrect, got: %d, want: %d.", c.reg[A], 0)
	}
	if !c.IsFlagSet(CFlag) {
		t.Errorf("Test -> A: 0xFF + n: 0x1 -> Carry flag should be set")
	}
	if !c.IsFlagSet(ZFlag) {
		t.Errorf("Test -> A: 0xFF + n: 0x1 -> Zero flag should be set")
	}
}
