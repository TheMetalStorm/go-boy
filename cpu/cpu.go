package cpu

import (
	"fmt"
	"go-boy/mmap"

	"os"
)

type Mmap = mmap.Mmap

type Cpu struct {
	A uint8
	F uint8
	B uint8
	C uint8
	D uint8
	E uint8
	H uint8
	L uint8

	SP uint16
	PC uint16

	Memory *Mmap
}

var IO_START_ADDR uint16 = 0xff00

func NewCpu() *Cpu {
	cpu := &Cpu{}
	cpu.Restart()
	return cpu
}

func (cpu *Cpu) Restart() {

	cpu.Memory = &mmap.Mmap{}

	cpu.A = 0x01
	cpu.F = 0xB0
	cpu.B = 0x00
	cpu.C = 0x13
	cpu.D = 0x00
	cpu.E = 0xD8
	cpu.H = 0x01
	cpu.L = 0x4D
	cpu.SP = 0xFFFE
	cpu.PC = 0x100

	cpu.Memory.SetValue(0xFF05, 0x00) // TIMA
	cpu.Memory.SetValue(0xFF06, 0x00) // TMA
	cpu.Memory.SetValue(0xFF07, 0x00) // TAC
	cpu.Memory.SetValue(0xFF10, 0x80) // NR10
	cpu.Memory.SetValue(0xFF11, 0xBF) // NR11
	cpu.Memory.SetValue(0xFF12, 0xF3) // NR12
	cpu.Memory.SetValue(0xFF14, 0xBF) // NR14
	cpu.Memory.SetValue(0xFF16, 0x3F) // NR21
	cpu.Memory.SetValue(0xFF17, 0x00) // NR22
	cpu.Memory.SetValue(0xFF19, 0xBF) // NR24
	cpu.Memory.SetValue(0xFF1A, 0x7F) // NR30
	cpu.Memory.SetValue(0xFF1B, 0xFF) // NR31
	cpu.Memory.SetValue(0xFF1C, 0x9F) // NR32
	cpu.Memory.SetValue(0xFF1E, 0xBF) // NR33
	cpu.Memory.SetValue(0xFF20, 0xFF) // NR41
	cpu.Memory.SetValue(0xFF21, 0x00) // NR42
	cpu.Memory.SetValue(0xFF22, 0x00) // NR43
	cpu.Memory.SetValue(0xFF23, 0xBF) // NR30
	cpu.Memory.SetValue(0xFF24, 0x77) // NR50
	cpu.Memory.SetValue(0xFF25, 0xF3) // NR51
	cpu.Memory.SetValue(0xFF26, 0xF1) // NR52 (GB)
	cpu.Memory.SetValue(0xFF40, 0x91) // LCDC
	cpu.Memory.SetValue(0xFF42, 0x00) // SCY
	cpu.Memory.SetValue(0xFF43, 0x00) // SCX
	cpu.Memory.SetValue(0xFF45, 0x00) // LYC
	cpu.Memory.SetValue(0xFF47, 0xFC) // BGP
	cpu.Memory.SetValue(0xFF48, 0xFF) // OBP0
	cpu.Memory.SetValue(0xFF49, 0xFF) // OBP1
	cpu.Memory.SetValue(0xFF4A, 0x00) // WY
	cpu.Memory.SetValue(0xFF4B, 0x00) // WX
	cpu.Memory.SetValue(0xFFFF, 0x00) // IE

}

func (cpu *Cpu) Step() uint64 {
	instr, _ := cpu.Memory.ReadByteAt(cpu.PC)
	var ranMCyclesThisStep uint64 = 4 //instr fetch (and decode i guess) takes 4 (machine?) cycles
	//decode/Execute
	ranMCyclesThisStep += cpu.decodeExecute(instr)
	return ranMCyclesThisStep
}

// returns machine cycles it took to execute
func (cpu *Cpu) decodeExecute(instr byte) (cycles uint64) {

	switch instr {

	//nop
	case 0x00:
		cpu.PC++
		return 1
	//cb
	case 0xcb:
		return cpu.handleCB()

	//16 Load 16 Bit Imm to Reg
	case 0x31:
		r := cpu.loadImm16Reg(&cpu.SP)
		return r
	case 0x21:
		return cpu.loadImm16Reg2Ptr(&cpu.H, &cpu.L)
	case 0x11:
		return cpu.loadImm16Reg2Ptr(&cpu.D, &cpu.E)

	// Load 8 Bit Imm to Reg
	case 0x3e:
		return cpu.loadImm8IntoReg(&cpu.A)
	case 0x06:
		return cpu.loadImm8IntoReg(&cpu.B)
	case 0x0e:
		return cpu.loadImm8IntoReg(&cpu.C)
	case 0x1e:
		return cpu.loadImm8IntoReg(&cpu.E)
	case 0x2e:
		return cpu.loadImm8IntoReg(&cpu.L)

		// decrement Reg8
	case 0x3d:
		return cpu.decrementReg8(&cpu.A)
	case 0x05:
		return cpu.decrementReg8(&cpu.B)
	case 0x0d:
		return cpu.decrementReg8(&cpu.C)

		// increment Reg8
	case 0x04:
		return cpu.incrementReg8(&cpu.B)
	case 0x0c:
		return cpu.incrementReg8(&cpu.C)

		// increment Reg16
	case 0x13:
		return cpu.incrementReg16(REG_DE)
	case 0x23:
		return cpu.incrementReg16(REG_HL)
	//jump
	case 0xC3:
		return cpu.jump()
	case 0x20:
		return cpu.jumpRelIf(cpu.GetZeroFlag() == 0)
	case 0x28:
		return cpu.jumpRelIf(cpu.GetZeroFlag() != 0)
	case 0x18:
		return cpu.jumpRelIf(true)
		// call
	case 0xcd:
		return cpu.call16Imm()
	//ret
	case 0xc9:

		readLow, _ := cpu.Memory.ReadByteAt(cpu.SP)
		cpu.SP++

		readHigh, _ := cpu.Memory.ReadByteAt(cpu.SP)
		cpu.SP++
		newPC := (uint16(readLow) | uint16(readHigh)<<8)

		cpu.PC = newPC

		return 4

	// xor Reg
	case 0xaf:
		return cpu.xorReg(&cpu.A)

		//store reg in mem
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
	case 0xe2:
		return cpu.storeRegInMemAddr(IO_START_ADDR+uint16(cpu.C), cpu.A)

	case 0x77:
		return cpu.storeRegInMemAddr(cpu.GetHL(), cpu.A)

	case 0xea:
		return cpu.storeRegInImmMemAddr(cpu.A)

	//store reg in imm mem
	case 0xe0:
		return cpu.storeRegInAfterIoImmMemAddr(cpu.A)

	//store mem in reg
	case 0x1a:
		return cpu.storeMemIntoReg(cpu.GetDE(), &cpu.A)

		//store imm mem in reg
	case 0xf0:
		return cpu.storeAfterIoImm8MemAddrIntoReg(&cpu.A)

	// store Reg in Reg
	case 0x7b:
		return cpu.storeValInReg(&cpu.A, cpu.E)
	case 0x4f:
		return cpu.storeValInReg(&cpu.C, cpu.A)
	case 0x57:
		return cpu.storeValInReg(&cpu.D, cpu.A)
	case 0x67:
		return cpu.storeValInReg(&cpu.H, cpu.A)
	// push 16
	case 0xf5:
		return cpu.push16(&cpu.A, &cpu.F)
	case 0xc5:
		return cpu.push16(&cpu.B, &cpu.C)

	// pop 16
	case 0xc1:
		return cpu.pop16(&cpu.B, &cpu.C)

	// set ie
	case 0xfb:
		cpu.PC++
		cpu.Memory.Io.SetIE(1)
		return 1
	case 0xf3:
		cpu.PC++
		cpu.Memory.Io.SetIE(0)
		return 1

	//compare imm to A
	case 0xFE:
		cpu.PC++
		data, skip := cpu.Memory.ReadByteAt(cpu.PC)
		cpu.PC += skip
		cpu.SetZeroFlag((cpu.A - data) == 0)
		cpu.SetSubFlag(true)
		cpu.SetHalfCarryFlag(isHalfCarryFlagSubtraction(cpu.A, data))
		cpu.SetCarryFlag(isCarryFlagSubtraction(cpu.A, data))
		return 2
	//RLA
	//similiar to cbRegRotateLeft (other numBytes, numCycles and different flags)
	case 0x17:
		cpu.PC++
		var newCarry bool
		oldCarry := cpu.GetCarryFlag()

		oldRegVal := cpu.A
		cpu.A = oldRegVal << 1

		if oldCarry == 1 {
			cpu.A |= 1
		} else {
			cpu.A &^= (1)
		}

		cpu.SetZeroFlag(false)
		cpu.SetSubFlag(false)
		cpu.SetHalfCarryFlag(false)
		newCarry = ((oldRegVal >> 7 & 1) == 1)
		cpu.SetCarryFlag(newCarry)

		return 1
	default:
		fmt.Printf("ERROR at PC 0x%04x: 0x%02x is not a recognized instruction!\n", cpu.PC, instr)
		os.Exit(0)
		return 0

	}
}

func (cpu *Cpu) handleCB() (cycles uint64) {

	cpu.PC++
	instr, numReadBytes := cpu.Memory.ReadByteAt(cpu.PC)
	cpu.PC += numReadBytes

	switch instr {
	case 0x7c:
		return cpu.cbSetZeroToComplementRegBit(&cpu.H, 7)
	case 0x11:
		return cpu.cbRegRotateLeft(&cpu.C)
	default:
		cpu.PC -= 2
		fmt.Printf("ERROR at PC 0x%04x: 0xcb%02x is not a recognized instruction!\n", cpu.PC, instr)
		os.Exit(0)

		return 0
	}
}

func (cpu *Cpu) cbRegRotateLeft(regPtr *uint8) uint64 {
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
	var bit uint8

	bit = *regPtr >> bitPos & 0x1

	if bit == 0 {
		cpu.SetZeroFlag(true)
	} else {
		cpu.SetZeroFlag(false)
	}
	cpu.SetSubFlag(false)
	cpu.SetHalfCarryFlag(true)
	return 2
}

func isCarryFlagSubtraction(valA uint8, valB uint8) bool {

	return valB > valA
}

func isCarryFlagAddition(valA uint8, valB uint8) bool {

	var add uint16 = uint16(valA) + uint16(valB)

	return (add) > 0xFF
}

func isHalfCarryFlagSubtraction(valA uint8, valB uint8) bool {

	lowerA := getLower4(valA)
	lowerB := getLower4(valB)

	return lowerB > lowerA
}

func isHalfCarryFlagAddition(valA uint8, valB uint8) bool {

	lowerA := getLower4(valA)
	lowerB := getLower4(valB)

	return (lowerA + lowerB) > 0xF
}

func (cpu *Cpu) pop16(higherRegPtr *uint8, lowerRegPtr *uint8) (cycles uint64) {
	cpu.PC++

	readLow, _ := cpu.Memory.ReadByteAt(cpu.SP)
	*lowerRegPtr = readLow
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
func (cpu *Cpu) call16Imm() (cycles uint64) {

	cpu.PC++
	newPCAddr, bytesRead := cpu.Memory.Read16At(cpu.PC)
	cpu.PC += bytesRead

	// With the push, the current value of SP is decremented by 1, and the higher-order byte of PC is loaded in the
	// memory address specified by the new SP value. The value of SP is then decremented by 1 again, and the lower-order
	//byte of PC is loaded in the memory address specified by that value of SP.
	cpu.SP--
	cpu.Memory.SetValue(cpu.SP, getHigher8(cpu.PC))
	cpu.SP--
	cpu.Memory.SetValue(cpu.SP, getLower8(cpu.PC)) // lower order byte of PC

	//The subroutine is placed after the location specified by the new PC value. When the subroutine finishes, control is
	//returned to the source program using a return instruction and by popping the starting address of the next
	//instruction (which was just pushed) and moving it to the Pcpu.

	cpu.PC = newPCAddr
	// The lower-order byte of a16 is placed in byte 2 of the object code, and the higher-order byte is placed in byte 3.
	newPCAddrHigher := getHigher8(cpu.PC)
	newPCAddrLower := getLower8(cpu.PC)
	cpu.Memory.Oam[2] = newPCAddrLower
	cpu.Memory.Oam[3] = newPCAddrHigher

	return 6
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

		cpu.PC += uint16(signedData)
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
	default:
		fmt.Printf("ERROR: Func %s, reg %s not Implemented!", "incrementReg16", reg.String())
	}
	return 2
}

func (cpu *Cpu) jump() (cycles uint64) {
	cpu.PC, _ = cpu.Memory.Read16At(cpu.PC + 1)
	return 4
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

	*higherRegPtr = getHigher8(val)
	*lowerRegPtr = getLower8(val)

	return 3

}

func (cpu *Cpu) xorReg(regPtr *uint8) (cycles uint64) {

	*regPtr = 0

	cpu.SetZeroFlag(true)
	cpu.SetCarryFlag(false)
	cpu.SetHalfCarryFlag(false)
	cpu.SetSubFlag(false)

	cpu.PC++
	return 1

}

func getHigher8(orig uint16) uint8 {
	return uint8(orig >> 8 & 0xFF)
}

func getLower8(orig uint16) uint8 {
	return uint8(orig)
}

func getHigher4(orig uint8) uint8 {
	return uint8((orig >> 4) & 0x0F)
}

func getLower4(orig uint8) uint8 {
	return uint8(orig & 0x0f)
}

func (cpu *Cpu) DumpRegs() {
	fmt.Printf("Registers:\n\n")
	fmt.Printf("A: 0x%02X\n", cpu.A)
	fmt.Printf("F: 0x%02X\n", cpu.F)
	fmt.Printf("B: 0x%02X\n", cpu.B)
	fmt.Printf("C: 0x%02X\n", cpu.C)
	fmt.Printf("D: 0x%02X\n", cpu.D)
	fmt.Printf("E: 0x%02X\n", cpu.E)
	fmt.Printf("H: 0x%02X\n", cpu.H)
	fmt.Printf("L: 0x%02X\n", cpu.L)
	fmt.Printf("SP: 0x%04X\n", cpu.SP)
	fmt.Printf("PC: 0x%04X\n", cpu.PC)
}

func (cpu *Cpu) GetAF() uint16 {
	return uint16(cpu.A)<<8 | uint16(cpu.F)
}

func (cpu *Cpu) GetBC() uint16 {
	return uint16(cpu.B)<<8 | uint16(cpu.C)
}

func (cpu *Cpu) GetDE() uint16 {
	return uint16(cpu.D)<<8 | uint16(cpu.E)

}

func (cpu *Cpu) GetHL() uint16 {
	return uint16(cpu.H)<<8 | uint16(cpu.L)

}

func (cpu *Cpu) GetZeroFlag() uint8 { //z
	return (cpu.F >> 0x7) & 0x1
}

func (cpu *Cpu) GetSubFlag() uint8 { //n
	return (cpu.F >> 0x6) & 0x1

}

func (cpu *Cpu) GetHalfCarryFlag() uint8 { //h
	return (cpu.F >> 0x5) & 0x1

}

func (cpu *Cpu) GetCarryFlag() uint8 { // c
	return (cpu.F >> 0x4) & 0x1
}

//Setter

func (cpu *Cpu) SetAF(value uint16) {
	cpu.A = uint8(value >> 8)
	cpu.F = uint8(value)
}

func (cpu *Cpu) SetBC(value uint16) {
	cpu.B = uint8(value >> 8)
	cpu.C = uint8(value)
}

func (cpu *Cpu) SetDE(value uint16) {
	cpu.D = uint8(value >> 8)
	cpu.E = uint8(value)
}

func (cpu *Cpu) SetHL(value uint16) {
	cpu.H = uint8(value >> 8)
	cpu.L = uint8(value)
}
func (cpu *Cpu) SetZeroFlag(cond bool) { //z
	if cond {
		cpu.SetAF(cpu.GetAF() | 1<<7)
	} else {
		cpu.SetAF(cpu.GetAF() &^ (1 << 7))
	}
}

func (cpu *Cpu) SetSubFlag(cond bool) { //n
	if cond {
		cpu.SetAF(cpu.GetAF() | 1<<6)
	} else {
		cpu.SetAF(cpu.GetAF() &^ (1 << 6))
	}
}

func (cpu *Cpu) SetHalfCarryFlag(cond bool) { //h

	if cond {
		cpu.SetAF(cpu.GetAF() | 1<<5)
	} else {
		cpu.SetAF(cpu.GetAF() &^ (1 << 5))
	}
}

func (cpu *Cpu) SetCarryFlag(cond bool) { // c
	if cond {
		cpu.SetAF(cpu.GetAF() | 1<<4)
	} else {
		cpu.SetAF(cpu.GetAF() &^ (1 << 4))
	}
}
