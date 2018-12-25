package backend

import "testing"

func TestEmulatorRunsBootromAsExpected(t *testing.T) {
	c := NewCPU()
	c.ram[0xFF44] = 144

	for c.PC != 0x100 {
		c.DecodeAndExecuteNext()
	}

	if c.ReadAF() != 0x01B0 {
		t.Error()
	}
	if c.ReadBC() != 0x0013 {
		t.Error()
	}
	if c.ReadDE() != 0x00D8 {
		t.Error()
	}
	if c.ReadHL() != 0x014D {
		t.Error()
	}

	tests := []struct {
		address       uint16
		expectedValue byte
	}{
		{0xFF05, 0x00}, //    TIMA
		{0xFF06, 0x00}, //    TMA
		{0xFF07, 0x00}, //    TAC
		// {0xFF10, 0x80}, //    NR10
		// {0xFF11, 0xBF}, //    NR11
		// {0xFF12, 0xF3}, //    NR12
		// {0xFF14, 0xBF}, //    NR14
		// {0xFF16, 0x3F}, //    NR21
		// {0xFF17, 0x00}, //    NR22
		// {0xFF19, 0xBF}, //    NR24
		// {0xFF1A, 0x7F}, //    NR30
		// {0xFF1B, 0xFF}, //    NR31
		// {0xFF1C, 0x9F}, //    NR32
		// {0xFF1E, 0xBF}, //    NR33
		// {0xFF20, 0xFF}, //    NR41
		// {0xFF21, 0x00}, //    NR42
		// {0xFF22, 0x00}, //    NR43
		// {0xFF23, 0xBF}, //    NR30
		// {0xFF24, 0x77}, //    NR50
		// {0xFF25, 0xF3}, //    NR51
		// {0xFF26, 0xF1}, //    NR52
		{0xFF40, 0x91}, //    LCDC
		{0xFF42, 0x00}, //    SCY
		{0xFF43, 0x00}, //    SCX
		{0xFF45, 0x00}, //    LYC
		{0xFF47, 0xFC}, //    BGP
		// {0xFF48, 0xFF}, //    OBP0
		// {0xFF49, 0xFF}, //    OBP1
		{0xFF4A, 0x00}, //    WY
		{0xFF4B, 0x00}, //    WX
		{0xFFFF, 0x00}, //    IE
	}

	for _, test := range tests {
		if c.ram[test.address] != test.expectedValue {
			t.Errorf("Got unexpected value @ 0x%0.4X. Expected: 0x%0.2X, Got: 0x%0.2X",
				test.address, test.expectedValue, c.ram[test.address])
		}
	}

}
