package cpu

import (
	"fmt"
	"go-boy/rom"
	"os"
)

type Rom = rom.Rom

type Cpu struct {
	AF uint16
	BC uint16
	DE uint16
	HL uint16
	SP uint16
	PC uint16

	loadedRom *Rom
}

func (c *Cpu) LoadRom(r *Rom) {
	// TODO: reset cpu state
	c.loadedRom = r
}

func (c *Cpu) Step() {
	//fetch Instruction
	instr, _ := c.loadedRom.ReadByte(c.PC)
	//decode/Execute
	c.decodeExecute(instr)
}

func (c *Cpu) decodeExecute(instr byte) {
	c.PC += 1

	switch instr {

	case 0xC3:
		c.jumpWhen(c.GetZeroFlag() == 0)
	case 0xaf:
		c.xorReg(REG_A)
	case 0x21:
		c.loadImm16Reg(REG_HL)
	case 0x0e:
		c.loadImm8Reg(REG_A)
	default:
		c.PC -= 1
		fmt.Printf("ERROR: 0x%02x is not a recognized instruction!\n", instr)
		fmt.Println("-----------------------------------------------------------------")
		c.Dump()
		fmt.Println("-----------------------------------------------------------------")
		os.Exit(1)
	}

}

func (c *Cpu) jumpWhen(cond bool) {
	if cond {
		c.PC, _ = c.loadedRom.Read16(c.PC)
	}
}

func (c *Cpu) loadImm8Reg(reg Reg) {
	var skip uint16
	var val uint8
	val, skip = c.loadedRom.ReadByte(c.PC)
	c.PC += skip
	switch reg {
	case REG_A:
		c.SetA(val)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "loadImm8Reg", reg.String())
		os.Exit(1)
	}
}

func (c *Cpu) loadImm16Reg(reg Reg) {
	var skip uint16
	var val uint16

	val, skip = c.loadedRom.Read16(c.PC)
	c.PC += skip

	switch reg {
	case REG_HL:
		c.HL = val
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "loadImm16Reg", reg.String())
		os.Exit(1)
	}

}

func (c *Cpu) xorReg(reg Reg) {

	switch reg {
	case REG_A:
		c.SetA(0)
	default:
		fmt.Printf("ERROR: func %s, %s is not a recognized implemented!\n", "xorReg", reg.String())
		os.Exit(1)
	}

	c.SetZeroFlag(1)
	c.SetCarryFlag(0)
	c.SetHalfCarryFlag(0)
	c.SetSubFlag(0)

}

func (c *Cpu) Dump() {
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

func (c *Cpu) SetZeroFlag(setTo int) { //z
	if setTo == 1 {
		c.AF |= 1 << 7
	} else if setTo == 0 {
		c.AF &^= (1 << 7)
	}
}

func (c *Cpu) SetSubFlag(setTo int) { //n
	if setTo == 1 {
		c.AF |= 1 << 6
	} else if setTo == 0 {
		c.AF &^= (1 << 6)
	}
}

func (c *Cpu) SetHalfCarryFlag(setTo int) { //h
	if setTo == 1 {
		c.AF |= 1 << 5
	} else if setTo == 0 {
		c.AF &^= (1 << 5)
	}
}

func (c *Cpu) SetCarryFlag(setTo int) { // c
	if setTo == 1 {
		c.AF |= 1 << 4
	} else if setTo == 0 {
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
