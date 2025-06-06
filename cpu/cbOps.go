package cpu

import (
	"fmt"
	"go-boy/mmap"
	"os"
)

func (cpu *Cpu) handleCB() (cycles uint64) {
	cpu.PC++
	data, numReadBytes := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += numReadBytes

	instr := data >> 6
	switch instr {

	//RES
	case 0:
		return 0
	case 1:
		//Bit
		// bitIndex := (data >> 3) & 0x07
		// operand := data & 0x07
		return 0
	case 2:
		//RES
		bitIndex := (data >> 3) & 0x07
		operand := data & 0x07
		return cpu.cbResetRegBit(bitIndex, Reg8(operand))
	case 3:
		//SET
		bitIndex := (data >> 3) & 0x07
		operand := data & 0x07
		return cpu.cbSetRegBit(bitIndex, Reg8(operand))

	//END SET

	// case 0x7c:
	// 	return cpu.cbSetZeroToComplementRegBit(&cpu.H, 7)
	// case 0x11:
	// 	return cpu.cbRegRotateLeftWithCarryInBit0(&cpu.C)
	default:
		cpu.PC -= 2
		fmt.Printf("ERROR at PC 0x%04x: 0xcb%02x is not a recognized instruction!\n", cpu.PC, instr)
		os.Exit(0)

		return 0
	}
}

func (cpu *Cpu) cbResetRegBit(bitIndex uint8, operand Reg8) uint64 {
	var ptr *uint8

	switch operand {
	case REG_A:
		ptr = &cpu.A
	case REG_B:
		ptr = &cpu.B
	case REG_C:
		ptr = &cpu.C
	case REG_D:
		ptr = &cpu.D
	case REG_E:
		ptr = &cpu.E
	case REG_H:
		ptr = &cpu.H
	case REG_L:
		ptr = &cpu.L
	case REG_MEM_HL:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 0, false)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	}
	mmap.SetBit(ptr, bitIndex, false)
	return 2
}

func (cpu *Cpu) cbSetRegBit(bitIndex uint8, operand Reg8) uint64 {
	var ptr *uint8

	switch operand {
	case REG_A:
		ptr = &cpu.A
	case REG_B:
		ptr = &cpu.B
	case REG_C:
		ptr = &cpu.C
	case REG_D:
		ptr = &cpu.D
	case REG_E:
		ptr = &cpu.E
	case REG_H:
		ptr = &cpu.H
	case REG_L:
		ptr = &cpu.L
	case REG_MEM_HL:
		addr, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&addr, 0, true)
		cpu.Memory.SetValue(cpu.GetHL(), addr)
		return 4
	}
	mmap.SetBit(ptr, bitIndex, true)
	return 2
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
