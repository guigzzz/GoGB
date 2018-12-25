package backend

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddReg(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0x3A
	c.reg[B] = 0xC6
	c.AddReg(B, false)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, HFlag, CFlag)
}

func TestAddn(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0x3C
	c.Addn(0xFF, false)
	assert.Equal(t, c.reg[A], byte(0x3B))
	assertFlagsSet(t, c.reg[F], HFlag, CFlag)
}

func TestAddHL(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0x3C
	c.ram[0] = 0x12 // (HL = 0) = 0
	c.AddHL(false)
	assert.Equal(t, c.reg[A], byte(0x4E))
	assertFlagsSet(t, c.reg[F])
}

func TestADCReg(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0xE1
	c.reg[E] = 0x0F
	c.ram[0] = 0x1E
	c.SetFlag(CFlag)

	c.AddReg(E, true)
	assert.Equal(t, c.reg[A], byte(0xF1))
	assertFlagsSet(t, c.reg[F], HFlag)
}

func TestADCN(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0xE1
	c.reg[E] = 0x0F
	c.ram[0] = 0x1E
	c.SetFlag(CFlag)

	c.Addn(0x3B, true)
	assert.Equal(t, c.reg[A], byte(0x1D))
	assertFlagsSet(t, c.reg[F], CFlag)
}

func TestADCHL(t *testing.T) {
	c := NewTestCPU()

	c.reg[A] = 0xE1
	c.reg[E] = 0x0F
	c.ram[0] = 0x1E
	c.SetFlag(CFlag)

	c.AddHL(true)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, HFlag, CFlag)
}

func TestSubReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3E
	c.reg[E] = 0x3E
	c.ram[0] = 0x40

	c.SubReg(E, false)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, NFlag)
}

func TestSubN(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3E
	c.reg[E] = 0x3E
	c.ram[0] = 0x40

	c.Subn(0x0F, false)
	assert.Equal(t, c.reg[A], byte(0x2F))
	assertFlagsSet(t, c.reg[F], HFlag, NFlag)
}

func TestSubHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3E
	c.reg[E] = 0x3E
	c.ram[0] = 0x40

	c.SubHL(false)
	assert.Equal(t, c.reg[A], byte(0xFE))
	assertFlagsSet(t, c.reg[F], NFlag, CFlag)
}

func TestSbcReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3B
	c.reg[H] = 0x2A
	c.ram[0] = 0x4F
	c.SetFlag(CFlag)

	c.SubReg(H, true)
	assert.Equal(t, c.reg[A], byte(0x10))
	assertFlagsSet(t, c.reg[F], NFlag)
}

func TestSbcN(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3B
	c.reg[H] = 0x2A
	c.ram[0] = 0x4F
	c.SetFlag(CFlag)

	c.Subn(0x3A, true)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, NFlag)
}

func TestSbcHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3B
	c.reg[H] = 0x2A
	c.ram[0x2A00] = 0x4F
	c.SetFlag(CFlag)

	c.SubHL(true)
	assert.Equal(t, c.reg[A], byte(0xEB))
	assertFlagsSet(t, c.reg[F], HFlag, NFlag, CFlag)
}

func TestAndReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x5A
	c.reg[L] = 0x3F
	c.ram[0x003F] = 0

	c.AndReg(L)
	assert.Equal(t, c.reg[A], byte(0x1A))
	assertFlagsSet(t, c.reg[F], HFlag)
}

func TestAndN(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x5A
	c.reg[L] = 0x3F
	c.ram[0x003F] = 0

	c.Andn(0x38)
	assert.Equal(t, c.reg[A], byte(0x18))
	assertFlagsSet(t, c.reg[F], HFlag)
}

func TestAndHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x5A
	c.reg[L] = 0x3F
	c.ram[0x003F] = 0

	c.AndHL()
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, HFlag)
}

func TestOrReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x5A
	c.ram[0] = 0x0F

	c.OrReg(A)
	assert.Equal(t, c.reg[A], byte(0x5A))
	assertFlagsSet(t, c.reg[F])
}

func TestOrN(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x5A
	c.ram[0] = 0x0F

	c.Orn(0x3)
	assert.Equal(t, c.reg[A], byte(0x5B))
	assertFlagsSet(t, c.reg[F])
}

func TestOrHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x5A
	c.ram[0] = 0x0F

	c.OrHL()
	assert.Equal(t, c.reg[A], byte(0x5F))
	assertFlagsSet(t, c.reg[F])
}

func TestXorReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0xFF
	c.ram[0] = 0x8A

	c.XorReg(A)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag)
}

func TestXorN(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0xFF
	c.ram[0] = 0x8A

	c.Xorn(0xF)
	assert.Equal(t, c.reg[A], byte(0xF0))
	assertFlagsSet(t, c.reg[F])
}

func TestXorHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0xFF
	c.ram[0] = 0x8A

	c.XorHL()
	assert.Equal(t, c.reg[A], byte(0x75))
	assertFlagsSet(t, c.reg[F])
}

func TestCpReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3C
	c.reg[B] = 0x2F
	c.ram[0] = 0x40

	c.CpReg(B)
	assertFlagsSet(t, c.reg[F], HFlag, NFlag)
}

func TestCpN(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3C
	c.reg[B] = 0x2F
	c.ram[0] = 0x40

	c.Cpn(0x3C)
	assertFlagsSet(t, c.reg[F], ZFlag, NFlag)
}

func TestCpHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x3C
	c.reg[B] = 0x2F
	c.ram[0] = 0x40

	c.CpHL()
	assertFlagsSet(t, c.reg[F], NFlag, CFlag)
}

func TestIncReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0xFF

	c.Inc(A)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, HFlag)
}

func TestIncHL(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0x50

	c.IncHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0x51))
	assertFlagsSet(t, c.reg[F])
}

func TestDecReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[L] = 0x01

	c.Dec(L)
	assert.Equal(t, c.reg[L], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag, NFlag)
}

func TestDecHL(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0

	c.DecHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0xFF))
	assertFlagsSet(t, c.reg[F], HFlag, NFlag)
}

func TestAddHL16(t *testing.T) {
	c := NewTestCPU()
	c.Writedouble(H, L, 0x8A23)
	c.Writedouble(B, C, 0x0605)

	c.AddHL16(c.ReadBC())
	assert.Equal(t, c.ReadHL(), uint16(0x9028))
	assertFlagsSet(t, c.reg[F], HFlag)

	c = NewTestCPU()
	c.Writedouble(H, L, 0x8A23)
	c.Writedouble(B, C, 0x0605)

	c.AddHL16(c.ReadHL())
	assert.Equal(t, c.ReadHL(), uint16(0x1446))
	assertFlagsSet(t, c.reg[F], HFlag, CFlag)
}

func TestIncAndDecRegs(t *testing.T) {
	c := NewTestCPU()
	c.Writedouble(D, E, 0x235F)

	c.IncRegs(D, E)
	assert.Equal(t, c.ReadDE(), uint16(0x2360))

	c.DecRegs(D, E)
	assert.Equal(t, c.ReadDE(), uint16(0x235F))
}

func TestSwapReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0
	c.ram[0] = 0xF0

	c.SwapReg(A)
	assert.Equal(t, c.reg[A], byte(0))
	assertFlagsSet(t, c.reg[F], ZFlag)
}

func TestSwapHL(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0
	c.ram[0] = 0xF0

	c.SwapHL()
	assert.Equal(t, c.ram[c.ReadHL()], byte(0x0F))
	assertFlagsSet(t, c.reg[F])
}

func TestBitReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x80
	c.reg[L] = 0xEF

	c.Bit(A, 7)
	assertFlagsSet(t, c.reg[F], HFlag)

	c.Bit(L, 4)
	assertFlagsSet(t, c.reg[F], ZFlag, HFlag)
}

func TestBitHL(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0xFE

	c.BitHL(0)
	assertFlagsSet(t, c.reg[F], ZFlag, HFlag)

	c.BitHL(1)
	assertFlagsSet(t, c.reg[F], HFlag)
}

func TestSetReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x80
	c.reg[L] = 0x3B

	c.Set(A, 3)
	assert.Equal(t, c.reg[A], byte(0x88))

	c.Set(L, 7)
	assert.Equal(t, c.reg[L], byte(0xBB))
}

func TestSetHL(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0

	c.SetHL(3)
	assert.Equal(t, c.ram[c.ReadHL()], byte(0x08))
}

func TestResReg(t *testing.T) {
	c := NewTestCPU()
	c.reg[A] = 0x80
	c.reg[L] = 0x3B

	c.Res(A, 7)
	assert.Equal(t, c.reg[A], byte(0))

	c.Res(L, 1)
	assert.Equal(t, c.reg[L], byte(0x39))
}

func TestResHL(t *testing.T) {
	c := NewTestCPU()
	c.ram[0] = 0xFF

	c.ResHL(3)
	assert.Equal(t, c.ram[c.ReadHL()], byte(0xF7))
}

func TestAddSPN(t *testing.T) {
	c := NewTestCPU()
	c.SP = 0xFFF8

	c.AddSP8(2)
	assert.Equal(t, c.SP, uint16(0xFFFA))
	assertFlagsSet(t, c.reg[F])
}

func assertFlagsSet(t *testing.T, actualFlag byte, expectedAssertedFlags ...Flag) {
	if len(expectedAssertedFlags) == 0 {
		assert.Equal(t, actualFlag, byte(0))
	} else {
		expectedFlag := byte(0)
		for _, f := range expectedAssertedFlags {
			expectedFlag |= byte(f)
		}
		assert.Equal(t, actualFlag, expectedFlag, fmt.Sprintf(
			"Error with flags. Got: 0b%0.8b, Expected: 0b%0.8b", actualFlag, expectedFlag))
	}
}
