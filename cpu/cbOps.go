package cpu

import (
	"go-boy/mmap"
)

func (cpu *Cpu) handleCB() (cycles uint64) {
	cpu.PC++
	data, numReadBytes := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += numReadBytes

	instr := data >> 6
	operand := data & 0x07

	switch instr {

	//RES
	case 0:
		actualInstr := (data >> 3) & 0x1f
		switch actualInstr {
		//TODO implement all of these
		case 0:
			//rlc r8
			return cpu.cbRlc(Reg8(operand))
		case 1:
			// rrc r8
			return cpu.cbRrc(Reg8(operand))
		case 2:
			//rl r8
			return cpu.cbRl(Reg8(operand))
		case 3:
			//rr r8
			return cpu.cbRr(Reg8(operand))
		case 4:
			// sla r8
			return 0
		case 5:
			// sra r8
			return 0
		case 6:
			// swap r8
			return 0
		case 7:
			//srl r8
			return 0
		default:
			//No Opportunity for missing instructions
			return 0
		}
	case 1:
		bitIndex := (data >> 3) & 0x07
		return cpu.cbBit(bitIndex, Reg8(operand))
	case 2:
		bitIndex := (data >> 3) & 0x07
		return cpu.cbResetRegBit(bitIndex, Reg8(operand))
	case 3:
		bitIndex := (data >> 3) & 0x07
		return cpu.cbSetRegBit(bitIndex, Reg8(operand))
	default:
		//No Opportunity for missing instructions
		return 0
	}
}

// case 0x11:
//	return cpu.cbRegRotateLeftWithCarryInBit0(&cpu.C)

func (cpu *Cpu) cbBit(bitIndex uint8, operand Reg8) uint64 {

	var regData uint8

	_ = regData
	switch operand {
	case REG_A:
		regData = cpu.A
	case REG_B:
		regData = cpu.B
	case REG_C:
		regData = cpu.C
	case REG_D:
		regData = cpu.D
	case REG_E:
		regData = cpu.E
	case REG_H:
		regData = cpu.H
	case REG_L:
		regData = cpu.L
	case REG_MEM_HL:
		regData, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		bit := mmap.GetBit(regData, bitIndex)

		cpu.SetZeroFlag(!bit)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(true)

		return 3
	}
	bit := mmap.GetBit(regData, bitIndex)

	cpu.SetZeroFlag(!bit)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(true)

	return 2
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
		regVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		mmap.SetBit(&regVal, 0, true)
		cpu.Memory.SetValue(cpu.GetHL(), regVal)
		return 4
	}
	mmap.SetBit(ptr, bitIndex, true)
	return 2
}

func (cpu *Cpu) cbRlc(operand Reg8) uint64 {

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
		regVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		newCarry := mmap.GetBit(regVal, 7)
		regVal <<= 1
		mmap.SetBit(&regVal, 0, newCarry)
		cpu.Memory.SetValue(cpu.GetHL(), regVal)

		cpu.SetZeroFlag(regVal == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)
		return 4
	}

	newCarry := mmap.GetBit(*ptr, 7)
	*ptr <<= 1
	mmap.SetBit(ptr, 0, newCarry)

	cpu.SetZeroFlag(*ptr == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetCarryFlag(newCarry)

	return 2
}

func (cpu *Cpu) cbRrc(operand Reg8) uint64 {

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
		regVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		newCarry := mmap.GetBit(regVal, 0)
		regVal >>= 1
		mmap.SetBit(&regVal, 7, newCarry)
		cpu.Memory.SetValue(cpu.GetHL(), regVal)

		cpu.SetZeroFlag(regVal == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)
		return 4
	}

	newCarry := mmap.GetBit(*ptr, 0)
	*ptr >>= 1
	mmap.SetBit(ptr, 7, newCarry)

	cpu.SetZeroFlag(*ptr == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetCarryFlag(newCarry)

	return 2
}

func (cpu *Cpu) cbRl(operand Reg8) uint64 {

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
		regVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		newCarry := mmap.GetBit(regVal, 7)
		regVal <<= 1
		mmap.SetBit(&regVal, 0, cpu.GetCarryFlag() == 1)
		cpu.Memory.SetValue(cpu.GetHL(), regVal)

		cpu.SetZeroFlag(regVal == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)
		return 4
	}

	newCarry := mmap.GetBit(*ptr, 7)
	*ptr <<= 1
	mmap.SetBit(ptr, 0, cpu.GetCarryFlag() == 1)

	cpu.SetZeroFlag(*ptr == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetCarryFlag(newCarry)

	return 2
}

func (cpu *Cpu) cbRr(operand Reg8) uint64 {

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
		regVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		newCarry := mmap.GetBit(regVal, 0)
		regVal >>= 1
		mmap.SetBit(&regVal, 7, cpu.GetCarryFlag() == 1)
		cpu.Memory.SetValue(cpu.GetHL(), regVal)

		cpu.SetZeroFlag(regVal == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)
		return 4
	}

	newCarry := mmap.GetBit(*ptr, 0)
	*ptr >>= 1
	mmap.SetBit(ptr, 7, cpu.GetCarryFlag() == 1)

	cpu.SetZeroFlag(*ptr == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetCarryFlag(newCarry)

	return 2
}
