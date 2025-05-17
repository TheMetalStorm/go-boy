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

func (c *Cpu) LoadRom(r *Rom) {

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

	case 0x31:
		return c.loadImm16Reg(REG_SP)
	case 0xC3:
		return c.jump()
	case 0xaf:
		return c.xorReg(REG_A)
	case 0x21:
		return c.loadImm16Reg(REG_HL)
	case 0x0e:
		return c.loadImm8Reg(REG_A)
	case 0x06:
		return c.loadImm8Reg(REG_B)
	case 0x32:
		mc := c.storeInMemAddr(c.HL, c.GetA())
		c.HL--
		return mc
	case 0x05:
		return c.decrementReg8(REG_B)
	case 0x20:
		return c.jumpRelIf(c.GetZeroFlag() == 0)
	case 0xcb:
		return c.handleCB()
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
		return c.bcSetZeroToComplementRegBit(REG_H, 7)
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

func (c *Cpu) bcSetZeroToComplementRegBit(reg Reg8, bitPos int) uint64 {
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
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "bcSetZeroToRegBit", reg.String())
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

func (c *Cpu) jump() (cycles uint64) {
	c.PC, _ = c.memory.Read16At(c.PC + 1)
	return 4
}

func (c *Cpu) storeInMemAddr(storeAddr uint16, toStore uint8) (cycles uint64) {
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

func (c *Cpu) loadImm16Reg(reg Reg16) (cycles uint64) {
	var skip uint16
	var val uint16

	c.PC++
	val, skip = c.memory.Read16At(c.PC)
	c.PC += skip

	switch reg {
	case REG_HL:
		c.HL = val
	case REG_BC:
		c.BC = val
	case REG_DE:
		c.DE = val
	case REG_SP:
		c.SP = val
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "loadImm16Reg", reg.String())
		os.Exit(1)
	}

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
