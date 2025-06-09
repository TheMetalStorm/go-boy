package cpu

import (
	"fmt"
	"go-boy/mmap"
	"math"
	"os"
)

// returns machine cycles it took to execute
func (cpu *Cpu) decodeExecute(instr byte) (cycles uint64) {

	switch instr {

	//nop
	case 0x00:
		cpu.PC++
		return 1

	//HALT
	//NOTE: THANKS CHATGPT
	case 0x76:
		cpu.PC++

		if !cpu.IME {
			if cpu.Memory.GetIe()&cpu.Memory.Io.GetIF() != 0 {
				// HALT BUG TRIGGERED: Skip next instruction!
				cpu.PC++
			} else {
				cpu.Halt = true
			}
		} else {
			cpu.Halt = true
		}
		return 1

	//Stop
	case 0x10:
		cpu.PC++
		read, _ := cpu.Memory.ReadByteAt(cpu.PC)
		if read == 0x00 {
			fmt.Printf("ERROR at PC 0x%04x: we do not handle Stop Instruction right now :(\n", cpu.PC)
			os.Exit(0)
			return 0
			// cpu.Stop = true
			// cpu.PC += skip
			// return 1
		}
		fmt.Printf("ERROR at PC 0x%04x: 0x%02x is not a valid stop instruction! Stop instruction should be 0x1000, but found 0x%04x\n", cpu.PC, instr, read)
		os.Exit(0)
		return 0

	//DAA
	//https://github.com/guigzzz/GoGB/blob/master/backend/cpu_arithmetic.go#L349
	case 0x27:
		cpu.PC++

		val := cpu.A
		if cpu.GetSubFlag() == 0 {
			if cpu.GetHalfCarryFlag() == 1 || (val&0x0f) > 0x09 {
				val += 0x06
			}
			if cpu.GetCarryFlag() == 1 || val > 0x9f {
				val += 0x60
				cpu.SetCarryFlag(true)
			}
		} else {
			if cpu.GetHalfCarryFlag() == 1 {
				val -= 0x6
			}

			if cpu.GetCarryFlag() == 1 {
				val -= 0x60
			}
		}
		cpu.A = val

		cpu.SetZeroFlag(val&0x99 == 0)
		cpu.SetHalfCarryFlag(false)
		return 1

	//SCF
	case 0x37:
		cpu.PC++
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(true)
		return 1

	//CCF
	case 0x3f:
		cpu.PC++
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		flag := cpu.GetCarryFlag()
		if flag == 0 {
			cpu.SetCarryFlag(true)
		} else {
			cpu.SetCarryFlag(false)
		}
		return 1

	//CPL
	case 0x2f:
		cpu.PC++

		cpu.SetSubFlag(true)
		cpu.SetHalfCarryFlag(true)
		cpu.A = ^cpu.A

		return 1
		//cb
	case 0xcb:
		return cpu.handleCB()

		//16 Load 16 Bit Imm to Reg
	case 0x01:
		return cpu.loadImm16Reg2Ptr(&cpu.B, &cpu.C)
	case 0x11:
		return cpu.loadImm16Reg2Ptr(&cpu.D, &cpu.E)
	case 0x21:
		return cpu.loadImm16Reg2Ptr(&cpu.H, &cpu.L)
	case 0x31:
		return cpu.loadImm16Reg(&cpu.SP)

	//LD (a16), SP
	case 0x08:

		cpu.PC++
		addr, skip := cpu.Memory.Read16At(cpu.PC)
		cpu.PC += skip
		lower := GetLower8(cpu.SP)
		higher := GetHigher8(cpu.SP)
		cpu.Memory.SetValue(addr, lower)
		cpu.Memory.SetValue(addr+1, higher)
		return 5

	//ADD SP, s8
	case 0xe8:
		cpu.PC++
		imm, skip := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.PC += skip

		signedImm := int8(imm)

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)

		cpu.SetHalfCarryFlag(isHalfCarryFlagAddition(uint8(cpu.SP), imm))
		cpu.SetCarryFlag(isCarryFlagAddition(uint8(cpu.SP), imm))
		cpu.SP = uint16(int16(cpu.SP) + int16(signedImm))

		return 4
	//LD HL. SP + s8

	case 0xF8:
		cpu.PC++
		imm, skip := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.PC += skip

		signedImm := int8(imm)

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)

		cpu.SetHalfCarryFlag(isHalfCarryFlagAddition(uint8(cpu.SP), imm))
		cpu.SetCarryFlag(isCarryFlagAddition(uint8(cpu.SP), imm))
		cpu.SetHL(uint16(int16(cpu.SP) + int16(signedImm)))

		return 3

	//LD SP, HL
	case 0xf9:
		cpu.PC++
		cpu.SP = cpu.GetHL()
		return 2

	// Load 8 Bit Imm to Reg

	case 0x06:
		return cpu.loadImm8IntoReg(&cpu.B)
	case 0x16:
		return cpu.loadImm8IntoReg(&cpu.D)
	case 0x26:
		return cpu.loadImm8IntoReg(&cpu.H)
	case 0x36:
		cpu.PC++
		imm, skip := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.PC += skip
		cpu.Memory.SetValue(cpu.GetHL(), imm)
		return 3

	case 0x0e:
		return cpu.loadImm8IntoReg(&cpu.C)
	case 0x1e:
		return cpu.loadImm8IntoReg(&cpu.E)
	case 0x2e:
		return cpu.loadImm8IntoReg(&cpu.L)
	case 0x3e:
		return cpu.loadImm8IntoReg(&cpu.A)

	// decrement Reg8
	case 0x05:
		return cpu.decrementReg8(&cpu.B)
	case 0x15:
		return cpu.decrementReg8(&cpu.D)
	case 0x25:
		return cpu.decrementReg8(&cpu.H)
	case 0x35:
		cpu.PC++
		oldVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		newVal := oldVal - 1
		cpu.Memory.SetValue(cpu.GetHL(), newVal)

		cpu.SetZeroFlag(newVal == 0)
		cpu.SetSubFlag(true)
		cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(oldVal, 1))

		return 3

	case 0x0d:
		return cpu.decrementReg8(&cpu.C)
	case 0x1d:
		return cpu.decrementReg8(&cpu.E)
	case 0x2d:
		return cpu.decrementReg8(&cpu.L)
	case 0x3d:
		return cpu.decrementReg8(&cpu.A)

	// increment Reg8
	case 0x04:
		return cpu.incrementReg8(&cpu.B)
	case 0x14:
		return cpu.incrementReg8(&cpu.D)
	case 0x24:
		return cpu.incrementReg8(&cpu.H)
	case 0x34:
		cpu.PC++
		oldVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		newVal := oldVal + 1
		cpu.Memory.SetValue(cpu.GetHL(), newVal)

		cpu.SetZeroFlag(newVal == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(isHalfCarryFlagAddition(oldVal, 1))

		return 3

	case 0x0c:
		return cpu.incrementReg8(&cpu.C)
	case 0x1c:
		return cpu.incrementReg8(&cpu.E)
	case 0x2c:
		return cpu.incrementReg8(&cpu.L)
	case 0x3c:
		return cpu.incrementReg8(&cpu.A)

	// increment Reg16
	case 0x03:
		return cpu.incrementReg16(REG_BC)
	case 0x13:
		return cpu.incrementReg16(REG_DE)
	case 0x23:
		return cpu.incrementReg16(REG_HL)
	case 0x33:
		return cpu.incrementReg16(REG_SP)

	// add Reg16 to HL
	case 0x09:
		return cpu.addToHL(REG_BC)
	case 0x19:
		return cpu.addToHL(REG_DE)
	case 0x29:
		return cpu.addToHL(REG_HL)
	case 0x39:
		return cpu.addToHL(REG_SP)

	// decrement Reg16
	case 0x0b:
		return cpu.decrementReg16(REG_BC)
	case 0x1b:
		return cpu.decrementReg16(REG_DE)
	case 0x2b:
		return cpu.decrementReg16(REG_HL)
	case 0x3b:
		return cpu.decrementReg16(REG_SP)

	// jp HL
	case 0xe9:
		cpu.PC = cpu.GetHL()
		return 1
	//jump
	case 0xC3:
		return cpu.jumpIf(true)
	case 0xC2:
		return cpu.jumpIf(cpu.GetZeroFlag() == 0)
	case 0xD2:
		return cpu.jumpIf(cpu.GetCarryFlag() == 0)
	case 0xCA:
		return cpu.jumpIf(cpu.GetZeroFlag() == 1)
	case 0xDA:
		return cpu.jumpIf(cpu.GetCarryFlag() == 1)

	// jumpRel
	case 0x20:
		return cpu.jumpRelIf(cpu.GetZeroFlag() == 0)
	case 0x30:
		return cpu.jumpRelIf(cpu.GetCarryFlag() == 0)
	case 0x18:
		return cpu.jumpRelIf(true)
	case 0x28:
		return cpu.jumpRelIf(cpu.GetZeroFlag() == 1)
	case 0x38:
		return cpu.jumpRelIf(cpu.GetCarryFlag() == 1)

	// call
	case 0xcd:
		return cpu.call16ImmIf(true)
	case 0xc4:
		return cpu.call16ImmIf(cpu.GetZeroFlag() == 0)
	case 0xd4:
		return cpu.call16ImmIf(cpu.GetCarryFlag() == 0)
	case 0xcc:
		return cpu.call16ImmIf(cpu.GetZeroFlag() == 1)
	case 0xdc:
		return cpu.call16ImmIf(cpu.GetCarryFlag() == 1)

	//ret
	case 0xc9:
		return cpu.ret()
	//reti
	case 0xd9:
		cycles := cpu.ret()
		cpu.IME = true
		return cycles

	//retIf
	case 0xC0:
		return cpu.retIf(cpu.GetZeroFlag() == 0)
	case 0xD0:
		return cpu.retIf(cpu.GetCarryFlag() == 0)
	case 0xC8:
		return cpu.retIf(cpu.GetZeroFlag() == 1)
	case 0xD8:
		return cpu.retIf(cpu.GetCarryFlag() == 1)

		// load reg to reg/(HL)

	// In B
	case 0x40:
		return cpu.storeValInReg(&cpu.B, cpu.B)
	case 0x41:
		return cpu.storeValInReg(&cpu.B, cpu.C)
	case 0x42:
		return cpu.storeValInReg(&cpu.B, cpu.D)
	case 0x43:
		return cpu.storeValInReg(&cpu.B, cpu.E)
	case 0x44:
		return cpu.storeValInReg(&cpu.B, cpu.H)
	case 0x45:
		return cpu.storeValInReg(&cpu.B, cpu.L)
	case 0x46:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.B, value)
		cycles++
		return cycles
	case 0x47:
		return cpu.storeValInReg(&cpu.B, cpu.A)

		//END In B

		// In C
	case 0x48:
		return cpu.storeValInReg(&cpu.C, cpu.B)
	case 0x49:
		return cpu.storeValInReg(&cpu.C, cpu.C)
	case 0x4a:
		return cpu.storeValInReg(&cpu.C, cpu.D)
	case 0x4b:
		return cpu.storeValInReg(&cpu.C, cpu.E)
	case 0x4c:
		return cpu.storeValInReg(&cpu.C, cpu.H)
	case 0x4d:
		return cpu.storeValInReg(&cpu.C, cpu.L)
	case 0x4e:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.C, value)
		cycles++
		return cycles
	case 0x4f:
		return cpu.storeValInReg(&cpu.C, cpu.A)

	//END In C

	// In D
	case 0x50:
		return cpu.storeValInReg(&cpu.D, cpu.B)
	case 0x51:
		return cpu.storeValInReg(&cpu.D, cpu.C)
	case 0x52:
		return cpu.storeValInReg(&cpu.D, cpu.D)
	case 0x53:
		return cpu.storeValInReg(&cpu.D, cpu.E)
	case 0x54:
		return cpu.storeValInReg(&cpu.D, cpu.H)
	case 0x55:
		return cpu.storeValInReg(&cpu.D, cpu.L)
	case 0x56:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.D, value)
		cycles++
		return cycles
	case 0x57:
		return cpu.storeValInReg(&cpu.D, cpu.A)

		//END In D

		// In E
	case 0x58:
		return cpu.storeValInReg(&cpu.E, cpu.B)
	case 0x59:
		return cpu.storeValInReg(&cpu.E, cpu.C)
	case 0x5a:
		return cpu.storeValInReg(&cpu.E, cpu.D)
	case 0x5b:
		return cpu.storeValInReg(&cpu.E, cpu.E)
	case 0x5c:
		return cpu.storeValInReg(&cpu.E, cpu.H)
	case 0x5d:
		return cpu.storeValInReg(&cpu.E, cpu.L)
	case 0x5e:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.E, value)
		cycles++
		return cycles
	case 0x5f:
		return cpu.storeValInReg(&cpu.E, cpu.A)

	//END In E

	// In H
	case 0x60:
		return cpu.storeValInReg(&cpu.H, cpu.B)
	case 0x61:
		return cpu.storeValInReg(&cpu.H, cpu.C)
	case 0x62:
		return cpu.storeValInReg(&cpu.H, cpu.D)
	case 0x63:
		return cpu.storeValInReg(&cpu.H, cpu.E)
	case 0x64:
		return cpu.storeValInReg(&cpu.H, cpu.H)
	case 0x65:
		return cpu.storeValInReg(&cpu.H, cpu.L)
	case 0x66:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.H, value)
		cycles++
		return cycles
	case 0x67:
		return cpu.storeValInReg(&cpu.H, cpu.A)

	//END In H

	// In L
	case 0x68:
		return cpu.storeValInReg(&cpu.L, cpu.B)
	case 0x69:
		return cpu.storeValInReg(&cpu.L, cpu.C)
	case 0x6a:
		return cpu.storeValInReg(&cpu.L, cpu.D)
	case 0x6b:
		return cpu.storeValInReg(&cpu.L, cpu.E)
	case 0x6c:
		return cpu.storeValInReg(&cpu.L, cpu.H)
	case 0x6d:
		return cpu.storeValInReg(&cpu.L, cpu.L)
	case 0x6e:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.L, value)
		cycles++
		return cycles
	case 0x6f:
		return cpu.storeValInReg(&cpu.L, cpu.A)

	//END In L

	// IN (HL)

	case 0x70:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.B)
	case 0x71:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.C)
	case 0x72:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.D)
	case 0x73:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.E)
	case 0x74:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.H)
	case 0x75:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.L)
	//NOTE: dont worry, 0x76 is Halt
	case 0x77:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.A)

	//END in (HL)

	// In A
	case 0x78:
		return cpu.storeValInReg(&cpu.A, cpu.B)
	case 0x79:
		return cpu.storeValInReg(&cpu.A, cpu.C)
	case 0x7a:
		return cpu.storeValInReg(&cpu.A, cpu.D)
	case 0x7b:
		return cpu.storeValInReg(&cpu.A, cpu.E)
	case 0x7c:
		return cpu.storeValInReg(&cpu.A, cpu.H)
	case 0x7d:
		return cpu.storeValInReg(&cpu.A, cpu.L)
	case 0x7e:
		value, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cycles := cpu.storeValInReg(&cpu.A, value)
		cycles++
		return cycles
	case 0x7f:
		return cpu.storeValInReg(&cpu.A, cpu.A)

	//END In A

	// add to A Reg
	case 0x80:
		return cpu.addToRegA(cpu.B)
	case 0x81:
		return cpu.addToRegA(cpu.C)
	case 0x82:
		return cpu.addToRegA(cpu.D)
	case 0x83:
		return cpu.addToRegA(cpu.E)
	case 0x84:
		return cpu.addToRegA(cpu.H)
	case 0x85:
		return cpu.addToRegA(cpu.L)
	case 0x86:
		oldVal := cpu.A
		addVal, skip := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.A += addVal
		cpu.PC += skip

		cpu.SetZeroFlag(cpu.A == 0)
		cpu.SetSubFlag(false)
		cpu.SetCarryFlag(isCarryFlagAddition(oldVal, addVal))
		cpu.SetHalfCarryFlag(isHalfCarryFlagAddition(oldVal, addVal))

		return 2
	case 0x87:
		return cpu.addToRegA(cpu.A)
	case 0xc6:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.addToRegA(imm)
		return 2
		// END add to A Reg

	// add with carry to A Reg
	case 0x88:
		return cpu.addWithCarryToRegA(cpu.B)
	case 0x89:
		return cpu.addWithCarryToRegA(cpu.C)
	case 0x8a:
		return cpu.addWithCarryToRegA(cpu.D)
	case 0x8b:
		return cpu.addWithCarryToRegA(cpu.E)
	case 0x8c:
		return cpu.addWithCarryToRegA(cpu.H)
	case 0x8d:
		return cpu.addWithCarryToRegA(cpu.L)
	case 0x8e:

		read, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.addWithCarryToRegA(read)

		return 2
	case 0x8f:
		return cpu.addWithCarryToRegA(cpu.A)
	case 0xce:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.addWithCarryToRegA(imm)
		return 2

	// END add with carry to A Reg

	// sub from A Reg
	case 0x90:
		return cpu.subFromRegA(cpu.B)
	case 0x91:
		return cpu.subFromRegA(cpu.C)
	case 0x92:
		return cpu.subFromRegA(cpu.D)
	case 0x93:
		return cpu.subFromRegA(cpu.E)
	case 0x94:
		return cpu.subFromRegA(cpu.H)
	case 0x95:
		return cpu.subFromRegA(cpu.L)
	case 0x96:
		oldVal := cpu.A
		subVal, skip := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.A -= subVal
		cpu.PC += skip

		cpu.SetZeroFlag(cpu.A == 0)
		cpu.SetSubFlag(true)
		cpu.SetCarryFlag(isCarryFlagSubtraction(oldVal, subVal))
		cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(oldVal, subVal))

		return 2
	case 0x97:
		return cpu.subFromRegA(cpu.A)
	case 0xd6:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.subFromRegA(imm)
		return 2
	// END sub from A Reg

	// sub with carry to A Reg
	case 0x98:
		return cpu.subWithCarryFromRegA(cpu.B)
	case 0x99:
		return cpu.subWithCarryFromRegA(cpu.C)
	case 0x9a:
		return cpu.subWithCarryFromRegA(cpu.D)
	case 0x9b:
		return cpu.subWithCarryFromRegA(cpu.E)
	case 0x9c:
		return cpu.subWithCarryFromRegA(cpu.H)
	case 0x9d:
		return cpu.subWithCarryFromRegA(cpu.L)
	case 0x9e:
		subVal, _ := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.subWithCarryFromRegA(subVal)
		return 2
	case 0x9f:
		return cpu.subWithCarryFromRegA(cpu.A)
	case 0xDE:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.subWithCarryFromRegA(imm)
		return 2

	// END sub with carry to A Reg

	// bin and with A Reg
	case 0xa0:
		return cpu.binAndWithRegA(cpu.B)
	case 0xa1:
		return cpu.binAndWithRegA(cpu.C)
	case 0xa2:
		return cpu.binAndWithRegA(cpu.D)
	case 0xa3:
		return cpu.binAndWithRegA(cpu.E)
	case 0xa4:
		return cpu.binAndWithRegA(cpu.H)
	case 0xa5:
		return cpu.binAndWithRegA(cpu.L)
	case 0xa6:
		andVal, skip := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.A &= andVal
		cpu.PC += skip

		cpu.SetZeroFlag(cpu.A == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(true)
		cpu.SetCarryFlag(false)

		return 2
	case 0xa7:
		return cpu.binAndWithRegA(cpu.A)
	case 0xE6:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.binAndWithRegA(imm)
		return 2
	// END bin and with A Reg

	// xor Wit A Reg
	case 0xa8:
		return cpu.xorWithRegA(cpu.B)
	case 0xa9:
		return cpu.xorWithRegA(cpu.C)
	case 0xaa:
		return cpu.xorWithRegA(cpu.D)
	case 0xab:
		return cpu.xorWithRegA(cpu.E)
	case 0xac:
		return cpu.xorWithRegA(cpu.H)
	case 0xad:
		return cpu.xorWithRegA(cpu.L)
	case 0xae:
		val, skip := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.A ^= val

		cpu.SetZeroFlag(cpu.A == 0)
		cpu.SetCarryFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetSubFlag(false)

		cpu.PC += skip
		return 2
	case 0xaf:
		return cpu.xorWithRegA(cpu.A)
	case 0xEE:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.xorWithRegA(imm)
		return 2
	// END xor Wit A Reg

	// bin or with A Reg
	case 0xb0:
		return cpu.binOrWithRegA(cpu.B)
	case 0xb1:
		return cpu.binOrWithRegA(cpu.C)
	case 0xb2:
		return cpu.binOrWithRegA(cpu.D)
	case 0xb3:
		return cpu.binOrWithRegA(cpu.E)
	case 0xb4:
		return cpu.binOrWithRegA(cpu.H)
	case 0xb5:
		return cpu.binOrWithRegA(cpu.L)
	case 0xb6:
		orVal, skip := cpu.Memory.ReadByteAt(cpu.GetHL())
		cpu.A |= orVal
		cpu.PC += skip

		cpu.SetZeroFlag(cpu.A == 0)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(false)

		return 2
	case 0xb7:
		return cpu.binOrWithRegA(cpu.A)
	case 0xF6:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.binOrWithRegA(imm)
		return 2
		// END bin or with A Reg

	// compare With A Reg

	case 0xb8:
		return cpu.compareWithRegA(cpu.B)
	case 0xb9:
		return cpu.compareWithRegA(cpu.C)
	case 0xba:
		return cpu.compareWithRegA(cpu.D)
	case 0xbb:
		return cpu.compareWithRegA(cpu.E)
	case 0xbc:
		return cpu.compareWithRegA(cpu.H)
	case 0xbd:
		return cpu.compareWithRegA(cpu.L)
	case 0xbe:
		compVal, skip := cpu.Memory.ReadByteAt(cpu.GetHL())

		cpu.SetZeroFlag(cpu.A == compVal)
		cpu.SetSubFlag(true)
		cpu.SetCarryFlag(isCarryFlagSubtraction(cpu.A, compVal))
		cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(cpu.A, compVal))

		cpu.PC += skip
		return 2
	case 0xbf:
		return cpu.compareWithRegA(cpu.A)
	case 0xfe:
		cpu.PC++
		imm, _ := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.compareWithRegA(imm)
		return 2
	// END compare With A Reg

	//store reg in mem

	case 0x02:
		return cpu.storeRegInMemAddr(cpu.GetBC(), cpu.A)
	case 0x12:
		return cpu.storeRegInMemAddr(cpu.GetDE(), cpu.A)
	case 0x22:
		hl := cpu.GetHL()
		mc := cpu.storeRegInMemAddr(hl, cpu.A)
		cpu.SetHL(hl + 1)
		return mc
	case 0x32:
		hl := cpu.GetHL()
		mc := cpu.storeRegInMemAddr(hl, cpu.A)
		cpu.SetHL(hl - 1)
		return mc

	case 0xe0:
		return cpu.storeRegInAfterIoImmMemAddr(cpu.A)
	case 0xe2:
		return cpu.storeRegInMemAddr(IO_START_ADDR+uint16(cpu.C), cpu.A)
	case 0xea:
		return cpu.storeRegInImmMemAddr(cpu.A)

	//store mem in reg
	case 0x0a:
		return cpu.storeMemIntoReg(cpu.GetBC(), &cpu.A)
	case 0x1a:
		return cpu.storeMemIntoReg(cpu.GetDE(), &cpu.A)
	case 0x2a:
		hl := cpu.GetHL()
		cycles := cpu.storeMemIntoReg(hl, &cpu.A)
		cpu.SetHL(hl + 1)
		return cycles
	case 0x3a:
		hl := cpu.GetHL()
		cycles := cpu.storeMemIntoReg(hl, &cpu.A)
		cpu.SetHL(hl - 1)
		return cycles

	//store imm mem in reg
	case 0xf0:
		return cpu.storeAfterIoImm8MemAddrIntoReg(&cpu.A)
	case 0xf2:
		cpu.PC++
		loadedFromMem, _ := cpu.Memory.ReadByteAt(IO_START_ADDR + uint16(cpu.C))
		cpu.A = loadedFromMem
		return 2
	case 0xfa:
		cpu.PC++
		ptr, _ := cpu.Memory.Read16At(cpu.PC)
		val, _ := cpu.Memory.ReadByteAt(ptr)
		cpu.PC += 2
		cpu.A = val
		return 4
	// push 16

	case 0xc5:
		return cpu.push16(&cpu.B, &cpu.C)
	case 0xd5:
		return cpu.push16(&cpu.D, &cpu.E)
	case 0xe5:
		return cpu.push16(&cpu.H, &cpu.L)
	case 0xf5:
		return cpu.push16(&cpu.A, &cpu.F)

	// pop 16
	case 0xc1:
		return cpu.pop16(&cpu.B, &cpu.C, false)
	case 0xd1:
		return cpu.pop16(&cpu.D, &cpu.E, false)
	case 0xe1:
		return cpu.pop16(&cpu.H, &cpu.L, false)
	case 0xf1:
		return cpu.pop16(&cpu.A, &cpu.F, true)

	// ie
	case 0xfb:
		cpu.PC++
		cpu.IME = true
		return 1
	case 0xf3:
		cpu.PC++
		cpu.IME = false
		return 1

	//RLCA
	case 0x07:
		cpu.PC++

		newCarry := mmap.GetBit(cpu.A, 7)
		cpu.A = cpu.A << 1
		mmap.SetBit(&cpu.A, 0, newCarry)

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)

		return 1
		//RLA
		//similiar to cbRegRotateLeft (other numBytes, numCycles and different flags)
	case 0x17:
		cpu.PC++

		newCarry := mmap.GetBit(cpu.A, 7)
		cpu.A <<= 1
		mmap.SetBit(&cpu.A, 0, cpu.GetCarryFlag() == 1)

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)

		return 1

	//RRCA
	case 0x0F:
		cpu.PC++

		newCarry := mmap.GetBit(cpu.A, 0)
		cpu.A >>= 1
		mmap.SetBit(&cpu.A, 7, newCarry)

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)

		return 1

	//RRA
	case 0x1f:
		cpu.PC++
		newCarry := mmap.GetBit(cpu.A, 0)

		cpu.A >>= 1
		mmap.SetBit(&cpu.A, 7, cpu.GetCarryFlag() == 1)

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		cpu.SetCarryFlag(newCarry)

		return 1

	// RST
	case 0xc7:
		return cpu.rst(0x00)
	case 0xcf:
		return cpu.rst(0x08)
	case 0xd7:
		return cpu.rst(0x10)
	case 0xdf:
		return cpu.rst(0x18)
	case 0xe7:
		return cpu.rst(0x20)
	case 0xef:
		return cpu.rst(0x28)
	case 0xf7:
		return cpu.rst(0x30)
	case 0xff:
		return cpu.rst(0x38)

	default:
		fmt.Printf("ERROR at PC 0x%04x: 0x%02x is not a recognized instruction!\n", cpu.PC, instr)
		os.Exit(0)
		return 0

	}
}

func (cpu *Cpu) rst(newPC uint8) (cycles uint64) {

	cpu.PC++

	lowerPC := GetLower8(cpu.PC)
	higherPC := GetHigher8(cpu.PC)

	cpu.SP--
	cpu.Memory.SetValue(cpu.SP, higherPC)
	cpu.SP--
	cpu.Memory.SetValue(cpu.SP, lowerPC)

	cpu.PC = uint16(newPC)

	return 4
}

func (cpu *Cpu) addToHL(reg Reg16) (cycles uint64) {
	cpu.PC++

	oldHL := cpu.GetHL()
	var op uint16
	switch reg {
	case REG_BC:
		op = cpu.GetBC()
	case REG_DE:
		op = cpu.GetDE()
	case REG_HL:
		op = cpu.GetHL()
	case REG_SP:
		op = cpu.SP
	}

	cpu.SetHL(oldHL + op)

	cpu.SetSubFlag(false)
	cpu.SetCarryFlag(isCarryFlagAddition16(oldHL, op))
	cpu.SetHalfCarryFlag(isHalfCarryFlagAddition16(oldHL, op))

	return 2
}

func (cpu *Cpu) ret() (cycles uint64) {
	readLow, _ := cpu.Memory.ReadByteAt(cpu.SP)
	cpu.SP++

	readHigh, _ := cpu.Memory.ReadByteAt(cpu.SP)
	cpu.SP++
	newPC := (uint16(readLow) | uint16(readHigh)<<8)

	cpu.PC = newPC

	return 4

}

func (cpu *Cpu) retIf(cond bool) (cycles uint64) {
	if cond {
		readLow, _ := cpu.Memory.ReadByteAt(cpu.SP)
		cpu.SP++

		readHigh, _ := cpu.Memory.ReadByteAt(cpu.SP)
		cpu.SP++
		newPC := (uint16(readLow) | uint16(readHigh)<<8)

		cpu.PC = newPC

		return 5
	} else {
		cpu.PC++
		return 2
	}

}

func (cpu *Cpu) pop16(higherRegPtr *uint8, lowerRegPtr *uint8, isAF bool) (cycles uint64) {
	cpu.PC++

	readLow, _ := cpu.Memory.ReadByteAt(cpu.SP)
	*lowerRegPtr = readLow
	if isAF {
		*lowerRegPtr &= 0b11110000
	}
	cpu.SP++

	readHigh, _ := cpu.Memory.ReadByteAt(cpu.SP)
	*higherRegPtr = readHigh
	cpu.SP++

	return 3
}

func (cpu *Cpu) push16(higherRegPtr *uint8, lowerRegPtr *uint8) (cycles uint64) {

	cpu.PC++

	cpu.SP--
	cpu.Memory.SetValue(cpu.SP, *higherRegPtr)
	cpu.SP--
	cpu.Memory.SetValue(cpu.SP, *lowerRegPtr)
	return 4
}

func (cpu *Cpu) storeValInReg(regPtr *uint8, val uint8) (cycles uint64) {
	cpu.PC++
	*regPtr = val
	return 1
}

// In memory, push the program counter PC value corresponding to the address following the CALL instruction to the 2 bytes
// following the byte specified by the current stack pointer SP. Then load the 16-bit immediate operand a16 into Pcpu.
func (cpu *Cpu) call16ImmIf(cond bool) (cycles uint64) {
	if cond {
		cpu.PC++
		newPCAddr, bytesRead := cpu.Memory.Read16At(cpu.PC)
		cpu.PC += bytesRead

		// With the push, the current value of SP is decremented by 1, and the higher-order byte of PC is loaded in the
		// memory address specified by the new SP value. The value of SP is then decremented by 1 again, and the lower-order
		//byte of PC is loaded in the memory address specified by that value of SP.
		cpu.SP--
		cpu.Memory.SetValue(cpu.SP, GetHigher8(cpu.PC))
		cpu.SP--
		cpu.Memory.SetValue(cpu.SP, GetLower8(cpu.PC)) // lower order byte of PC

		//The subroutine is placed after the location specified by the new PC value. When the subroutine finishes, control is
		//returned to the source program using a return instruction and by popping the starting address of the next
		//instruction (which was just pushed) and moving it to the Pcpu.

		cpu.PC = newPCAddr
		// The lower-order byte of a16 is placed in byte 2 of the object code, and the higher-order byte is placed in byte 3.
		newPCAddrHigher := GetHigher8(cpu.PC)
		newPCAddrLower := GetLower8(cpu.PC)
		cpu.Memory.Oam[2] = newPCAddrLower
		cpu.Memory.Oam[3] = newPCAddrHigher

		return 6
	} else {
		cpu.PC += 3
		return 3
	}
}

func (cpu *Cpu) storeMemIntoReg(address uint16, regPtr *uint8) (cycles uint64) {

	val, bytesRead := cpu.Memory.ReadByteAt(address)
	cpu.PC += bytesRead

	*regPtr = val

	return 2
}

func (cpu *Cpu) storeAfterIoImm8MemAddrIntoReg(regPtr *uint8) (cycles uint64) {
	cpu.PC++
	immData, skip := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += skip

	loadedFromMem, _ := cpu.Memory.ReadByteAt(IO_START_ADDR + uint16(immData))

	*regPtr = loadedFromMem

	return 3
}

func (cpu *Cpu) storeRegInImmMemAddr(val uint8) (cycles uint64) {
	cpu.PC++
	a16, bytesRead := cpu.Memory.Read16At(cpu.PC)
	cpu.PC += bytesRead
	cpu.Memory.SetValue(a16, val)
	return 4
}

func (cpu *Cpu) storeRegInAfterIoImmMemAddr(val uint8) (cycles uint64) {
	cpu.PC++
	a8, bytesRead := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += bytesRead
	cpu.Memory.SetValue(IO_START_ADDR+uint16(a8), val)
	return 3
}

func (cpu *Cpu) jumpRelIf(cond bool) (cycles uint64) {
	cpu.PC++
	data, bytesRead := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += bytesRead
	if cond {
		signedData := int8(data)

		if signedData >= 0 {
			cpu.PC += uint16(signedData)
		} else {
			signedAbs := uint16(math.Abs(float64(signedData)))
			cpu.PC -= signedAbs

		}
		return 3
	}
	return 2

}
func (cpu *Cpu) decrementReg8(regPtr *uint8) (cycles uint64) {

	oldRegVal := *regPtr
	*regPtr = oldRegVal - 1

	cpu.SetZeroFlag(*regPtr == 0)
	cpu.SetSubFlag(true)
	cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(oldRegVal, 1))

	cpu.PC++
	return 1
}

func (cpu *Cpu) incrementReg8(regPtr *uint8) (cycles uint64) {
	oldRegVal := *regPtr
	*regPtr = oldRegVal + 1

	cpu.SetZeroFlag(*regPtr == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(isHalfCarryFlagAddition(oldRegVal, 1))

	cpu.PC++
	return 1
}

func (cpu *Cpu) incrementReg16(reg Reg16) (cycles uint64) {
	cpu.PC++

	switch reg {
	case REG_AF:
		cpu.SetAF(cpu.GetAF() + 1)
	case REG_BC:
		cpu.SetBC(cpu.GetBC() + 1)
	case REG_DE:
		cpu.SetDE(cpu.GetDE() + 1)
	case REG_HL:
		cpu.SetHL(cpu.GetHL() + 1)
	case REG_SP:
		cpu.SP += 1
	default:
		fmt.Printf("ERROR: Func %s, reg %s not Implemented!", "incrementReg16", reg.String())
	}
	return 2
}

func (cpu *Cpu) decrementReg16(reg Reg16) (cycles uint64) {
	cpu.PC++

	switch reg {
	case REG_AF:
		cpu.SetAF(cpu.GetAF() - 1)
	case REG_BC:
		cpu.SetBC(cpu.GetBC() - 1)
	case REG_DE:
		cpu.SetDE(cpu.GetDE() - 1)
	case REG_HL:
		cpu.SetHL(cpu.GetHL() - 1)
	case REG_SP:
		cpu.SP -= 1
	default:
		fmt.Printf("ERROR: Func %s, reg %s not Implemented!", "decrementReg16", reg.String())
	}
	return 2
}

func (cpu *Cpu) jumpIf(cond bool) (cycles uint64) {
	cpu.PC++
	newPC, skip := cpu.Memory.Read16At(cpu.PC)
	if cond {
		cpu.PC = newPC
		return 4
	} else {
		cpu.PC += skip
		return 3
	}
}

func (cpu *Cpu) storeRegInMemAddr(address uint16, toStore uint8) (cycles uint64) {

	cpu.Memory.SetValue(address, toStore)

	cpu.PC++
	return 2
}

func (cpu *Cpu) loadImm8IntoReg(regPtr *uint8) (cycles uint64) {
	var skip uint16
	var val uint8
	cpu.PC++
	val, skip = cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += skip

	*regPtr = val

	return 2
}

func (cpu *Cpu) loadImm16Reg(reg *uint16) (cycles uint64) {
	var skip uint16
	var val uint16

	cpu.PC++
	val, skip = cpu.Memory.Read16At(cpu.PC)
	cpu.PC += skip
	*reg = val

	return 3

}

func (cpu *Cpu) loadImm16Reg2Ptr(higherRegPtr *uint8, lowerRegPtr *uint8) (cycles uint64) {
	var skip uint16
	var val uint16

	cpu.PC++
	val, skip = cpu.Memory.Read16At(cpu.PC)
	cpu.PC += skip

	*higherRegPtr = GetHigher8(val)
	*lowerRegPtr = GetLower8(val)

	return 3

}

func (cpu *Cpu) addToRegA(regVal uint8) (cycles uint64) {
	cpu.PC++

	oldVal := cpu.A
	cpu.A += regVal

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetSubFlag(false)
	cpu.SetCarryFlag(isCarryFlagAddition(oldVal, regVal))
	cpu.SetHalfCarryFlag(isHalfCarryFlagAddition(oldVal, regVal))

	return 1

}

func (cpu *Cpu) addWithCarryToRegA(regVal uint8) (cycles uint64) {
	cpu.PC++

	oldVal := cpu.A
	addVal := regVal
	carry := cpu.GetCarryFlag()
	cpu.A += addVal + carry

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetSubFlag(false)

	cpu.SetCarryFlag((uint16(oldVal) + uint16(addVal) + uint16(carry)) > 0xFF)
	cpu.SetHalfCarryFlag(((oldVal & 0xF) + (addVal & 0xF) + carry) > 0xF)

	return 1

}

func (cpu *Cpu) subFromRegA(regVal uint8) (cycles uint64) {

	oldVal := cpu.A
	cpu.A -= regVal

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetSubFlag(true)
	cpu.SetCarryFlag(isCarryFlagSubtraction(oldVal, regVal))
	cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(oldVal, regVal))

	cpu.PC++
	return 1

}

func (cpu *Cpu) subWithCarryFromRegA(regVal uint8) (cycles uint64) {

	oldVal := cpu.A
	subVal := regVal
	carry := cpu.GetCarryFlag()
	cpu.A -= subVal + carry

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetSubFlag(true)
	cpu.SetCarryFlag((uint16(oldVal) - uint16(subVal) - uint16(carry)) > 0xFF)
	cpu.SetHalfCarryFlag(((oldVal & 0xF) - (subVal & 0xF) - carry) > 0xF)

	cpu.PC++
	return 1

}

func (cpu *Cpu) binAndWithRegA(regVal uint8) (cycles uint64) {

	cpu.A &= regVal

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(true)
	cpu.SetCarryFlag(false)

	cpu.PC++
	return 1

}

func (cpu *Cpu) binOrWithRegA(regVal uint8) (cycles uint64) {

	cpu.A |= regVal

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetCarryFlag(false)

	cpu.PC++
	return 1

}

func (cpu *Cpu) xorWithRegA(regVal uint8) (cycles uint64) {

	cpu.A ^= regVal

	cpu.SetZeroFlag(cpu.A == 0)
	cpu.SetCarryFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetSubFlag(false)

	cpu.PC++
	return 1

}

func (cpu *Cpu) compareWithRegA(regVal uint8) (cycles uint64) {

	cpu.SetZeroFlag(cpu.A == regVal)
	cpu.SetSubFlag(true)
	cpu.SetCarryFlag(isCarryFlagSubtraction(cpu.A, regVal))
	cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(cpu.A, regVal))

	cpu.PC++
	return 1
}
