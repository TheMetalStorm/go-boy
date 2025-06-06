package cpu

import (
	"fmt"
	"go-boy/mmap"
	"os"
)

func (cpu *Cpu) handleCB() (cycles uint64) {
	cpu.PC++
	instr, numReadBytes := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += numReadBytes

	switch instr {

	//RES
	case 0x80:
		mmap.SetBit(&cpu.B, 0, false)
		return 2
	case 0x81:
		mmap.SetBit(&cpu.C, 0, false)
		return 2
	case 0x82:
		mmap.SetBit(&cpu.D, 0, false)
		return 2
	case 0x83:
		mmap.SetBit(&cpu.E, 0, false)
		return 2
	case 0x84:
		mmap.SetBit(&cpu.H, 0, false)
		return 2
	case 0x85:
		mmap.SetBit(&cpu.L, 0, false)
		return 2
	case 0x86:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 0, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0x87:
		mmap.SetBit(&cpu.A, 0, false)
		return 2

	case 0x88:
		mmap.SetBit(&cpu.B, 1, false)
		return 2
	case 0x89:
		mmap.SetBit(&cpu.C, 1, false)
		return 2
	case 0x8A:
		mmap.SetBit(&cpu.D, 1, false)
		return 2
	case 0x8B:
		mmap.SetBit(&cpu.E, 1, false)
		return 2
	case 0x8C:
		mmap.SetBit(&cpu.H, 1, false)
		return 2
	case 0x8D:
		mmap.SetBit(&cpu.L, 1, false)
		return 2
	case 0x8E:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 1, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0x8F:
		mmap.SetBit(&cpu.A, 1, false)
		return 2

	case 0x90:
		mmap.SetBit(&cpu.B, 2, false)
		return 2
	case 0x91:
		mmap.SetBit(&cpu.C, 2, false)
		return 2
	case 0x92:
		mmap.SetBit(&cpu.D, 2, false)
		return 2
	case 0x93:
		mmap.SetBit(&cpu.E, 2, false)
		return 2
	case 0x94:
		mmap.SetBit(&cpu.H, 2, false)
		return 2
	case 0x95:
		mmap.SetBit(&cpu.L, 2, false)
		return 2
	case 0x96:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 2, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0x97:
		mmap.SetBit(&cpu.A, 2, false)
		return 2

	case 0x98:
		mmap.SetBit(&cpu.B, 3, false)
		return 2
	case 0x99:
		mmap.SetBit(&cpu.C, 3, false)
		return 2
	case 0x9A:
		mmap.SetBit(&cpu.D, 3, false)
		return 2
	case 0x9B:
		mmap.SetBit(&cpu.E, 3, false)
		return 2
	case 0x9C:
		mmap.SetBit(&cpu.H, 3, false)
		return 2
	case 0x9D:
		mmap.SetBit(&cpu.L, 3, false)
		return 2
	case 0x9E:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 3, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0x9F:
		mmap.SetBit(&cpu.A, 3, false)
		return 2

	case 0xA0:
		mmap.SetBit(&cpu.B, 4, false)
		return 2
	case 0xA1:
		mmap.SetBit(&cpu.C, 4, false)
		return 2
	case 0xA2:
		mmap.SetBit(&cpu.D, 4, false)
		return 2
	case 0xA3:
		mmap.SetBit(&cpu.E, 4, false)
		return 2
	case 0xA4:
		mmap.SetBit(&cpu.H, 4, false)
		return 2
	case 0xA5:
		mmap.SetBit(&cpu.L, 4, false)
		return 2
	case 0xA6:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 4, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xA7:
		mmap.SetBit(&cpu.A, 4, false)
		return 2

	case 0xA8:
		mmap.SetBit(&cpu.B, 5, false)
		return 2
	case 0xA9:
		mmap.SetBit(&cpu.C, 5, false)
		return 2
	case 0xAA:
		mmap.SetBit(&cpu.D, 5, false)
		return 2
	case 0xAB:
		mmap.SetBit(&cpu.E, 5, false)
		return 2
	case 0xAC:
		mmap.SetBit(&cpu.H, 5, false)
		return 2
	case 0xAD:
		mmap.SetBit(&cpu.L, 5, false)
		return 2
	case 0xAE:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 5, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xAF:
		mmap.SetBit(&cpu.A, 5, false)
		return 2

	case 0xB0:
		mmap.SetBit(&cpu.B, 6, false)
		return 2
	case 0xB1:
		mmap.SetBit(&cpu.C, 6, false)
		return 2
	case 0xB2:
		mmap.SetBit(&cpu.D, 6, false)
		return 2
	case 0xB3:
		mmap.SetBit(&cpu.E, 6, false)
		return 2
	case 0xB4:
		mmap.SetBit(&cpu.H, 6, false)
		return 2
	case 0xB5:
		mmap.SetBit(&cpu.L, 6, false)
		return 2
	case 0xB6:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 6, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xB7:
		mmap.SetBit(&cpu.A, 6, false)
		return 2

	case 0xB8:
		mmap.SetBit(&cpu.B, 7, false)
		return 2
	case 0xB9:
		mmap.SetBit(&cpu.C, 7, false)
		return 2
	case 0xBA:
		mmap.SetBit(&cpu.D, 7, false)
		return 2
	case 0xBB:
		mmap.SetBit(&cpu.E, 7, false)
		return 2
	case 0xBC:
		mmap.SetBit(&cpu.H, 7, false)
		return 2
	case 0xBD:
		mmap.SetBit(&cpu.L, 7, false)
		return 2
	case 0xBE:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 7, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xBF:
		mmap.SetBit(&cpu.A, 7, false)
		return 2

	// END RES

	//SET
	case 0xC0:
		mmap.SetBit(&cpu.B, 0, true)
		return 2
	case 0xC1:
		mmap.SetBit(&cpu.C, 0, true)
		return 2
	case 0xC2:
		mmap.SetBit(&cpu.D, 0, true)
		return 2
	case 0xC3:
		mmap.SetBit(&cpu.E, 0, true)
		return 2
	case 0xC4:
		mmap.SetBit(&cpu.H, 0, true)
		return 2
	case 0xC5:
		mmap.SetBit(&cpu.L, 0, true)
		return 2
	case 0xC6:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 0, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xC7:
		mmap.SetBit(&cpu.A, 0, true)
		return 2

	case 0xC8:
		mmap.SetBit(&cpu.B, 1, true)
		return 2
	case 0xC9:
		mmap.SetBit(&cpu.C, 1, true)
		return 2
	case 0xCA:
		mmap.SetBit(&cpu.D, 1, true)
		return 2
	case 0xCB:
		mmap.SetBit(&cpu.E, 1, true)
		return 2
	case 0xCC:
		mmap.SetBit(&cpu.H, 1, true)
		return 2
	case 0xCD:
		mmap.SetBit(&cpu.L, 1, true)
		return 2
	case 0xCE:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 1, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xCF:
		mmap.SetBit(&cpu.A, 1, true)
		return 2

	case 0xD0:
		mmap.SetBit(&cpu.B, 2, true)
		return 2
	case 0xD1:
		mmap.SetBit(&cpu.C, 2, true)
		return 2
	case 0xD2:
		mmap.SetBit(&cpu.D, 2, true)
		return 2
	case 0xD3:
		mmap.SetBit(&cpu.E, 2, true)
		return 2
	case 0xD4:
		mmap.SetBit(&cpu.H, 2, true)
		return 2
	case 0xD5:
		mmap.SetBit(&cpu.L, 2, true)
		return 2
	case 0xD6:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 2, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xD7:
		mmap.SetBit(&cpu.A, 2, true)
		return 2

	case 0xD8:
		mmap.SetBit(&cpu.B, 3, true)
		return 2
	case 0xD9:
		mmap.SetBit(&cpu.C, 3, true)
		return 2
	case 0xDA:
		mmap.SetBit(&cpu.D, 3, true)
		return 2
	case 0xDB:
		mmap.SetBit(&cpu.E, 3, true)
		return 2
	case 0xDC:
		mmap.SetBit(&cpu.H, 3, true)
		return 2
	case 0xDD:
		mmap.SetBit(&cpu.L, 3, true)
		return 2
	case 0xDE:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 3, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xDF:
		mmap.SetBit(&cpu.A, 3, true)
		return 2

	case 0xE0:
		mmap.SetBit(&cpu.B, 4, true)
		return 2
	case 0xE1:
		mmap.SetBit(&cpu.C, 4, true)
		return 2
	case 0xE2:
		mmap.SetBit(&cpu.D, 4, true)
		return 2
	case 0xE3:
		mmap.SetBit(&cpu.E, 4, true)
		return 2
	case 0xE4:
		mmap.SetBit(&cpu.H, 4, true)
		return 2
	case 0xE5:
		mmap.SetBit(&cpu.L, 4, true)
		return 2
	case 0xE6:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 4, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xE7:
		mmap.SetBit(&cpu.A, 4, true)
		return 2

	case 0xE8:
		mmap.SetBit(&cpu.B, 5, true)
		return 2
	case 0xE9:
		mmap.SetBit(&cpu.C, 5, true)
		return 2
	case 0xEA:
		mmap.SetBit(&cpu.D, 5, true)
		return 2
	case 0xEB:
		mmap.SetBit(&cpu.E, 5, true)
		return 2
	case 0xEC:
		mmap.SetBit(&cpu.H, 5, true)
		return 2
	case 0xED:
		mmap.SetBit(&cpu.L, 5, true)
		return 2
	case 0xEE:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 5, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xEF:
		mmap.SetBit(&cpu.A, 5, true)
		return 2

	case 0xF0:
		mmap.SetBit(&cpu.B, 6, true)
		return 2
	case 0xF1:
		mmap.SetBit(&cpu.C, 6, true)
		return 2
	case 0xF2:
		mmap.SetBit(&cpu.D, 6, true)
		return 2
	case 0xF3:
		mmap.SetBit(&cpu.E, 6, true)
		return 2
	case 0xF4:
		mmap.SetBit(&cpu.H, 6, true)
		return 2
	case 0xF5:
		mmap.SetBit(&cpu.L, 6, true)
		return 2
	case 0xF6:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 6, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xF7:
		mmap.SetBit(&cpu.A, 6, true)
		return 2

	case 0xF8:
		mmap.SetBit(&cpu.B, 7, true)
		return 2
	case 0xF9:
		mmap.SetBit(&cpu.C, 7, true)
		return 2
	case 0xFA:
		mmap.SetBit(&cpu.D, 7, true)
		return 2
	case 0xFB:
		mmap.SetBit(&cpu.E, 7, true)
		return 2
	case 0xFC:
		mmap.SetBit(&cpu.H, 7, true)
		return 2
	case 0xFD:
		mmap.SetBit(&cpu.L, 7, true)
		return 2
	case 0xFE:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 7, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	case 0xFF:
		mmap.SetBit(&cpu.A, 7, true)
		return 2

	//END SET

	case 0x7c:
		return cpu.cbSetZeroToComplementRegBit(&cpu.H, 7)
	case 0x11:
		return cpu.cbRegRotateLeftWithCarryInBit0(&cpu.C)
	default:
		cpu.PC -= 2
		fmt.Printf("ERROR at PC 0x%04x: 0xcb%02x is not a recognized instruction!\n", cpu.PC, instr)
		os.Exit(0)

		return 0
	}
}

func (cpu *Cpu) cbRegRotateLeftWithCarryInBit0(regPtr *uint8) uint64 {
	var oldRegVal uint8
	var oldCarry = cpu.GetCarryFlag()
	var newCarry bool

	oldRegVal = *regPtr
	*regPtr = oldRegVal << 1
	if oldCarry == 1 {
		*regPtr |= 1
	} else {
		*regPtr &^= (1)
	}

	cpu.SetZeroFlag(*regPtr == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(false)
	newCarry = ((oldRegVal >> 7 & 1) == 1)
	cpu.SetCarryFlag(newCarry)

	return 2
}
func (cpu *Cpu) cbSetZeroToComplementRegBit(regPtr *uint8, bitPos int) uint64 {
	bit := *regPtr >> bitPos & 0x1

	if bit == 0 {
		cpu.SetZeroFlag(true)
	} else {
		cpu.SetZeroFlag(false)
	}
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(true)
	return 2
}
