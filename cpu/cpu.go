package cpu

import (
	"fmt"
	"go-boy/mmap"
	"go-boy/rom"
	"os"
)

type Rom = rom.Rom
type Mmap = mmap.Mmap

// http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html
var MAX_CYCLES_PER_FRAME uint64 = 69905
var IO_START_ADDR uint16 = 0xff00

type Cpu struct {
	AF uint16
	BC uint16
	DE uint16
	HL uint16
	SP uint16
	PC uint16

	memory *Mmap
	//loadedRom          *Rom
	ranCyclesThisFrame uint64
}

func NewCpu() *Cpu {
	cpu := Cpu{}
	cpu.memory = &mmap.Mmap{}
	return &cpu
}

func (c *Cpu) LoadBootRom(r *Rom) {
	// TODO: different for gbc boot rom?
	for i := 0; i <= 0xff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		c.memory.SetValue(uint16(i), newVal)
	}

}

func (c *Cpu) Step() {

	//fetch Instruction
	instr, _ := c.memory.ReadByteAt(c.PC)
	c.ranCyclesThisFrame += 4 //instr fetch (and decode i guess) takes 4 (machine?) cycles
	//decode/Execute
	c.ranCyclesThisFrame += c.decodeExecute(instr)
}

// returns machine cycles it took to execute
func (c *Cpu) decodeExecute(instr byte) (cycles uint64) {
	fmt.Printf("Instr %02x\n", instr)

	switch instr {

	//cb
	case 0xcb:
		return c.handleCB()

	//16 Load 16 Bit Imm to Reg
	case 0x31:
		return c.loadImm16Reg(&c.SP)
	case 0x21:
		return c.loadImm16Reg(&c.HL)
	case 0x11:
		return c.loadImm16Reg(&c.DE)

	// Load 8 Bit Imm to Reg
	case 0x3e:
		return c.loadImm8Reg(REG_A)
	case 0x06:
		return c.loadImm8Reg(REG_B)
	case 0x0e:
		return c.loadImm8Reg(REG_C)

	// decrement Reg8
	case 0x05:
		return c.decrementReg8(REG_B)

	// increment Reg8
	case 0x0c:
		return c.incrementReg8(REG_C)

	//jump
	case 0xC3:
		return c.jump()
	case 0x20:
		return c.jumpRelIf(c.GetZeroFlag() == 0)

	// xor Reg
	case 0xaf:
		return c.xorReg(REG_A)

	//store reg in mem
	case 0x32:
		mc := c.storeRegInMemAddr(c.HL, c.GetA())
		c.HL--
		return mc
	case 0xe2:
		return c.storeRegInMemAddr(IO_START_ADDR+uint16(c.GetC()), c.GetA())

	case 0x77:
		return c.storeRegInMemAddr(c.HL, c.GetA())

	//store reg in imm mem
	case 0xe0:
		return c.storeRegInImmMemAddr(c.GetA())

	//store mem in reg
	case 0x1a:
		return c.storeMemInReg(c.DE, REG_A)

	// store val in Reg
	case 0x4f:
		return c.storeValInReg(REG_C, c.GetA())
	// call
	case 0xcd:
		return c.call16Imm()

	// push 16
	case 0xc5:
		return c.push16(&c.BC)

	// set ie
	case 0xfb:
		c.PC++
		c.memory.Io.SetIE(1)
		return 1

	//RLA
	//similiar to cbRegRotateLeft (other numBytes, numCycles and different flags)
	case 0x17:
		c.PC++
		var newCarry bool
		oldCarry := c.GetCarryFlag()

		oldRegVal := c.GetA()
		newRegVal := oldRegVal << 1

		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}

		c.SetZeroFlag(false)
		c.SetSubFlag(false)
		c.SetHalfCarryFlag(false)
		newCarry = ((oldRegVal >> 7 & 1) == 1)
		c.SetCarryFlag(newCarry)

		return 1
	default:
		fmt.Printf("ERROR: 0x%02x is not a recognized instruction!\n", instr)
		fmt.Println("-----------------------------------------------------------------")
		c.DumpRegs()
		fmt.Println("-----------------------------------------------------------------")
		os.Exit(1)
		return 0
	}
}

func (c *Cpu) handleCB() (cycles uint64) {
	c.PC++
	instr, numReadBytes := c.memory.ReadByteAt(c.PC)
	c.PC += numReadBytes
	fmt.Printf("-> cb%02x\n", instr)

	switch instr {
	case 0x7c:
		return c.cbSetZeroToComplementRegBit(REG_H, 7)
	case 0x11:
		return c.cbRegRotateLeft(REG_C)
	default:
		c.PC -= 2
		fmt.Printf("ERROR: 0xcb%02x is not a recognized instruction!\n", instr)
		fmt.Println("-----------------------------------------------------------------")
		c.DumpRegs()
		fmt.Println("-----------------------------------------------------------------")
		os.Exit(1)
		return 0

	}
}

func (c *Cpu) cbRegRotateLeft(reg Reg8) uint64 {
	var oldRegVal, newRegVal uint8
	var oldCarry = c.GetCarryFlag()

	var newCarry bool

	switch reg {
	case REG_A:
		oldRegVal = c.GetA()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetA(newRegVal)
	case REG_B:
		oldRegVal = c.GetB()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetB(newRegVal)
	case REG_C:
		oldRegVal = c.GetC()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetC(newRegVal)
	case REG_D:
		oldRegVal = c.GetD()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetD(newRegVal)
	case REG_E:
		oldRegVal = c.GetE()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetE(newRegVal)
	case REG_H:
		oldRegVal = c.GetH()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetH(newRegVal)
	case REG_L:
		oldRegVal = c.GetL()
		newRegVal = oldRegVal << 1
		if oldCarry == 1 {
			newRegVal |= 1
		} else {
			newRegVal &^= (1)
		}
		c.SetL(newRegVal)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "bcRegRotateLeft", reg.String())
		os.Exit(1)
	}

	c.SetZeroFlag(newRegVal == 0)
	c.SetSubFlag(false)
	c.SetHalfCarryFlag(false)
	newCarry = ((oldRegVal >> 7 & 1) == 1)
	c.SetCarryFlag(newCarry)

	return 2
}
func (c *Cpu) cbSetZeroToComplementRegBit(reg Reg8, bitPos int) uint64 {
	var bit uint8

	switch reg {
	case REG_A:
		bit = c.GetA() >> bitPos & 0x1
	case REG_B:
		bit = c.GetB() >> bitPos & 0x1
	case REG_C:
		bit = c.GetC() >> bitPos & 0x1
	case REG_D:
		bit = c.GetD() >> bitPos & 0x1
	case REG_E:
		bit = c.GetE() >> bitPos & 0x1
	case REG_H:
		bit = c.GetH() >> bitPos & 0x1
	case REG_L:
		bit = c.GetL() >> bitPos & 0x1
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "bcSetZeroToComplementRegBit", reg.String())
		os.Exit(1)
	}

	if bit == 0 {
		c.SetZeroFlag(true)
	} else {
		c.SetZeroFlag(false)
	}
	c.SetSubFlag(false)
	c.SetHalfCarryFlag(true)
	return 2
}

func isHalfCarryFlagSubtraction(valA uint8, valB uint8, result uint8) bool {

	return (valA^(-valB)^result)&0x10 != 0
}

func isHalfCarryFlagAddition(valA int, valB int, result int) bool {
	return (valA^valB^result)&0x10 != 0
}

func (c *Cpu) push16(regPtr *uint16) (cycles uint64) {

	c.PC++

	val := *regPtr

	c.SP--
	c.memory.SetValue(c.SP, getHigher(val))
	c.SP--
	c.memory.SetValue(c.SP, getLower(val))

	return 4
}

func (c *Cpu) storeValInReg(reg Reg8, val uint8) (cycles uint64) {
	c.PC++
	switch reg {
	case REG_A:
		c.SetA(val)
	case REG_B:
		c.SetB(val)
	case REG_C:
		c.SetC(val)
	case REG_D:
		c.SetD(val)
	case REG_E:
		c.SetE(val)
	case REG_H:
		c.SetH(val)
	case REG_L:
		c.SetL(val)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "storeMemInRegs", reg.String())
		os.Exit(1)
	}

	return 1
}

// In memory, push the program counter PC value corresponding to the address following the CALL instruction to the 2 bytes
// following the byte specified by the current stack pointer SP. Then load the 16-bit immediate operand a16 into PC.
func (c *Cpu) call16Imm() (cycles uint64) {

	c.PC++
	newPCAddr, bytesRead := c.memory.Read16At(c.PC)
	c.PC += bytesRead

	// With the push, the current value of SP is decremented by 1, and the higher-order byte of PC is loaded in the
	// memory address specified by the new SP value. The value of SP is then decremented by 1 again, and the lower-order
	//byte of PC is loaded in the memory address specified by that value of SP.
	c.SP--
	c.memory.SetValue(c.SP, getHigher(c.PC))
	c.SP--
	c.memory.SetValue(c.SP, getLower(c.PC)) // lower order byte of PC

	//The subroutine is placed after the location specified by the new PC value. When the subroutine finishes, control is
	//returned to the source program using a return instruction and by popping the starting address of the next
	//instruction (which was just pushed) and moving it to the PC.
	c.PC = newPCAddr
	// The lower-order byte of a16 is placed in byte 2 of the object code, and the higher-order byte is placed in byte 3.

	newPCAddrHigher := getHigher(c.PC)
	newPCAddrLower := getLower(c.PC)
	c.memory.Oam[2] = newPCAddrLower
	c.memory.Oam[3] = newPCAddrHigher

	return 6
}

func (c *Cpu) storeMemInReg(address uint16, reg Reg8) (cycles uint64) {

	val, bytesRead := c.memory.ReadByteAt(address)
	c.PC += bytesRead

	switch reg {
	case REG_A:
		c.SetA(val)
	case REG_B:
		c.SetB(val)
	case REG_C:
		c.SetC(val)
	case REG_D:
		c.SetD(val)
	case REG_E:
		c.SetE(val)
	case REG_H:
		c.SetH(val)
	case REG_L:
		c.SetL(val)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "storeMemInRegs", reg.String())
		os.Exit(1)
	}

	return 2
}

func (c *Cpu) storeRegInImmMemAddr(val uint8) (cycles uint64) {
	c.PC++
	a8, bytesRead := c.memory.ReadByteAt(c.PC)
	c.PC += bytesRead
	c.memory.SetValue(IO_START_ADDR+uint16(a8), val)
	return 3
}

func (c *Cpu) jumpRelIf(cond bool) (cycles uint64) {
	c.PC++

	if cond {
		var data byte
		var bytesRead uint16

		data, bytesRead = c.memory.ReadByteAt(c.PC)

		signedData := int8(data)

		fmt.Printf("Jump From %04x to %04x (2 + read data %04x )\n", c.PC-1, c.PC+bytesRead+uint16(signedData), int16(signedData))
		c.PC += bytesRead + uint16(signedData)
		return 3
	}
	return 2

}
func (c *Cpu) decrementReg8(reg Reg8) (cycles uint64) {
	var oldRegVal uint8
	var newRegVal uint8

	switch reg {
	case REG_A:
		oldRegVal = c.GetA()
		newRegVal = oldRegVal - 1
		c.SetA(newRegVal)

	case REG_B:
		oldRegVal = c.GetB()
		newRegVal = oldRegVal - 1
		c.SetB(newRegVal)

	case REG_C:
		oldRegVal = c.GetC()
		newRegVal = oldRegVal - 1
		c.SetC(newRegVal)

	case REG_D:
		oldRegVal = c.GetD()
		newRegVal = oldRegVal - 1
		c.SetD(newRegVal)

	case REG_E:
		oldRegVal = c.GetE()
		newRegVal = oldRegVal - 1
		c.SetE(newRegVal)

	case REG_H:
		oldRegVal = c.GetH()
		newRegVal = oldRegVal - 1
		c.SetH(newRegVal)

	case REG_L:
		oldRegVal = c.GetL()
		newRegVal = oldRegVal - 1
		c.SetL(newRegVal)

	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "decrementReg8", reg.String())
		os.Exit(1)
	}

	c.SetZeroFlag(newRegVal == 0)
	c.SetSubFlag(true)
	c.SetHalfCarryFlag(isHalfCarryFlagSubtraction(oldRegVal, 1, newRegVal))

	c.PC++
	return 1
}

func (c *Cpu) incrementReg8(reg Reg8) (cycles uint64) {
	var oldRegVal uint8
	var newRegVal uint8

	switch reg {
	case REG_A:
		oldRegVal = c.GetA()
		newRegVal = oldRegVal + 1
		c.SetA(newRegVal)

	case REG_B:
		oldRegVal = c.GetB()
		newRegVal = oldRegVal + 1
		c.SetB(newRegVal)

	case REG_C:
		oldRegVal = c.GetC()
		newRegVal = oldRegVal + 1
		c.SetC(newRegVal)

	case REG_D:
		oldRegVal = c.GetD()
		newRegVal = oldRegVal + 1
		c.SetD(newRegVal)

	case REG_E:
		oldRegVal = c.GetE()
		newRegVal = oldRegVal + 1
		c.SetE(newRegVal)

	case REG_H:
		oldRegVal = c.GetH()
		newRegVal = oldRegVal + 1
		c.SetH(newRegVal)

	case REG_L:
		oldRegVal = c.GetL()
		newRegVal = oldRegVal + 1
		c.SetL(newRegVal)

	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "incrementReg8", reg.String())
		os.Exit(1)
	}

	c.SetZeroFlag(newRegVal == 0)
	c.SetSubFlag(false)
	c.SetHalfCarryFlag(isHalfCarryFlagAddition(int(oldRegVal), 1, int(newRegVal)))

	c.PC++
	return 1
}

func (c *Cpu) jump() (cycles uint64) {
	c.PC, _ = c.memory.Read16At(c.PC + 1)
	return 4
}

func (c *Cpu) storeRegInMemAddr(storeAddr uint16, toStore uint8) (cycles uint64) {
	c.memory.SetValue(storeAddr, toStore)
	c.PC++
	return 2
}

func (c *Cpu) loadImm8Reg(reg Reg8) (cycles uint64) {
	var skip uint16
	var val uint8
	c.PC++
	val, skip = c.memory.ReadByteAt(c.PC)
	c.PC += skip
	switch reg {
	case REG_A:
		c.SetA(val)
	case REG_B:
		c.SetB(val)
	case REG_C:
		c.SetC(val)
	case REG_D:
		c.SetD(val)
	case REG_E:
		c.SetE(val)
	case REG_H:
		c.SetH(val)
	case REG_L:
		c.SetL(val)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "loadImm8Reg", reg.String())
		os.Exit(1)
	}
	return 2
}

func (c *Cpu) loadImm16Reg(reg *uint16) (cycles uint64) {
	var skip uint16
	var val uint16

	c.PC++
	val, skip = c.memory.Read16At(c.PC)
	c.PC += skip
	*reg = val

	return 3

}

func (c *Cpu) xorReg(reg Reg8) (cycles uint64) {

	switch reg {
	case REG_A:
		c.SetA(0)
	case REG_B:
		c.SetB(0)
	case REG_C:
		c.SetC(0)
	case REG_D:
		c.SetD(0)
	case REG_E:
		c.SetE(0)
	case REG_H:
		c.SetH(0)
	case REG_L:
		c.SetL(0)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "xorReg", reg.String())
		os.Exit(1)
	}

	c.SetZeroFlag(true)
	c.SetCarryFlag(false)
	c.SetHalfCarryFlag(false)
	c.SetSubFlag(false)

	c.PC++
	return 1

}

func getHigher(orig uint16) uint8 {
	return uint8(orig >> 8 & 0xFF)
}

func getLower(orig uint16) uint8 {
	return uint8(orig)
}

func (c *Cpu) DumpRegs() {
	fmt.Printf("Registers:\n\n")
	fmt.Printf("A: 0x%02X\n", c.GetA())
	fmt.Printf("F: 0x%02X\n", c.GetF())
	fmt.Printf("B: 0x%02X\n", c.GetB())
	fmt.Printf("C: 0x%02X\n", c.GetC())
	fmt.Printf("D: 0x%02X\n", c.GetD())
	fmt.Printf("E: 0x%02X\n", c.GetE())
	fmt.Printf("H: 0x%02X\n", c.GetH())
	fmt.Printf("L: 0x%02X\n", c.GetL())
	fmt.Printf("SP: 0x%04X\n", c.SP)
	fmt.Printf("PC: 0x%04X\n", c.PC)
}

// Getters for high and low bytes
func (c *Cpu) GetA() uint8 {
	return uint8(c.AF >> 8)
}

func (c *Cpu) GetF() uint8 {
	return uint8(c.AF & 0xFF)
}

func (c *Cpu) GetZeroFlag() uint8 { //z
	return (c.GetF() >> 0x7) & 0x1
}

func (c *Cpu) GetSubFlag() uint8 { //n
	return (c.GetF() >> 0x6) & 0x1

}

func (c *Cpu) GetHalfCarryFlag() uint8 { //h
	return (c.GetF() >> 0x5) & 0x1

}

func (c *Cpu) GetCarryFlag() uint8 { // c
	return (c.GetF() >> 0x4) & 0x1

}

func (c *Cpu) GetB() uint8 {
	return uint8(c.BC >> 8)
}

func (c *Cpu) GetC() uint8 {
	return uint8(c.BC & 0xFF)
}

func (c *Cpu) GetD() uint8 {
	return uint8(c.DE >> 8)
}

func (c *Cpu) GetE() uint8 {
	return uint8(c.DE & 0xFF)
}

func (c *Cpu) GetH() uint8 {
	return uint8(c.HL >> 8)
}

func (c *Cpu) GetL() uint8 {
	return uint8(c.HL & 0xFF)
}

//Setter

func (c *Cpu) SetA(setTo uint8) {
	newF := uint16(c.GetF())
	newA := uint16(setTo)
	c.AF = uint16(newF | newA<<8)
}

func (c *Cpu) SetZeroFlag(cond bool) { //z
	if cond {
		c.AF |= 1 << 7
	} else {
		c.AF &^= (1 << 7)
	}
}

func (c *Cpu) SetSubFlag(cond bool) { //n
	if cond {
		c.AF |= 1 << 6
	} else {
		c.AF &^= (1 << 6)
	}
}

func (c *Cpu) SetHalfCarryFlag(cond bool) { //h

	if cond {
		c.AF |= 1 << 5
	} else {
		c.AF &^= (1 << 5)
	}
}

func (c *Cpu) SetCarryFlag(cond bool) { // c
	if cond {
		c.AF |= 1 << 4
	} else {
		c.AF &^= (1 << 4)
	}
}

func (c *Cpu) SetB(setTo uint8) {
	newC := uint16(c.GetC())
	newB := uint16(setTo)
	c.BC = uint16(newC | newB<<8)
}

func (c *Cpu) SetC(setTo uint8) {
	newC := uint16(setTo)
	newB := uint16(c.GetB())
	c.BC = uint16(newC | newB<<8)
}

func (c *Cpu) SetD(setTo uint8) {
	newE := uint16(c.GetE())
	newD := uint16(setTo)
	c.DE = uint16(newE | newD<<8)
}

func (c *Cpu) SetE(setTo uint8) {
	newE := uint16(setTo)
	newD := uint16(c.GetD())
	c.DE = uint16(newE | newD<<8)
}

func (c *Cpu) SetH(setTo uint8) {
	newL := uint16(c.GetL())
	newH := uint16(setTo)
	c.HL = uint16(newL | newH<<8)
}

func (c *Cpu) SetL(setTo uint8) {
	newL := uint16(setTo)
	newH := uint16(c.GetH())
	c.HL = uint16(newL | newH<<8)
}
