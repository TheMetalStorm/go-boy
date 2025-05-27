package cpu

import (
	"fmt"
	"go-boy/mmap"
	"go-boy/rom"
	"os"
	"slices"
)

type Rom = rom.Rom
type Mmap = mmap.Mmap

// http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html
var MAX_CYCLES_PER_FRAME uint64 = 69905
var IO_START_ADDR uint16 = 0xff00

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
	//loadedRom          *Rom
	ranCyclesThisFrame uint64

	//Refactor into debugger struct
	Autorun     bool
	DoStep      bool
	breakpoints []uint16
	LastBPHit   int

	currentGame    *Rom
	currentBootrom *Rom
}

func NewCpu() *Cpu {
	cpu := Cpu{}
	cpu.Memory = &mmap.Mmap{}
	cpu.Autorun = true
	cpu.DoStep = false
	cpu.LastBPHit = -1
	return &cpu
}
func (cpu *Cpu) Restart() {
	cpu.A = 0
	cpu.F = 0
	cpu.B = 0
	cpu.C = 0
	cpu.D = 0
	cpu.E = 0
	cpu.H = 0
	cpu.L = 0
	cpu.SP = 0
	cpu.PC = 0
	cpu.Memory = &mmap.Mmap{}
	cpu.ranCyclesThisFrame = 0
	cpu.Autorun = false
	cpu.DoStep = false
	cpu.LastBPHit = -1

	//for Testing
	cpu.currentBootrom = rom.NewRom("./bootroms/dmg_boot.bin")
	cpu.LoadBootRom(cpu.currentBootrom)
	cpu.currentGame = rom.NewRom("./games/tetris.gb")
	cpu.LoadRom(cpu.currentGame)

}

func (c *Cpu) LoadBootRom(r *Rom) {
	// TODO: different for gbc boot rom?
	for i := 0; i <= 0xff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		c.Memory.SetValue(uint16(i), newVal)
	}

}

func (c *Cpu) LoadRom(r *Rom) {
	// TODO: for now only fills bank 0
	for i := 0x100; i <= 0x3fff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		c.Memory.SetValue(uint16(i), newVal)
	}
}

func (c *Cpu) UnloadBootrom(r *Rom) {
	for i := 0; i < 0x100; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		c.Memory.SetValue(uint16(i), newVal)
	}

}

func (c *Cpu) ToggleBP(addr uint16) {
	//No point in setting BP on Mem Adress 0
	if addr == 0 {
		return
	}
	for i, b := range c.breakpoints {
		if b == addr {
			c.breakpoints = append(c.breakpoints[:i], c.breakpoints[i+1:]...)
			return
		}
	}
	c.breakpoints = append(c.breakpoints, addr)
}

func (c *Cpu) Run() {
	for {
		//NOTE: HARDCODED FOR DMG BOOT ROM
		if c.PC == 0x100 {
			//unmap boot rom
			c.UnloadBootrom(c.currentGame)
		}
		if c.Autorun {
			if slices.Contains(c.GetBreakpoints(), c.PC) && c.PC != uint16(c.LastBPHit) {
				c.Autorun = false
				c.LastBPHit = int(c.PC)
			} else {
				c.Step()
			}
		} else {
			if c.DoStep {
				c.Step()
				c.DoStep = false
			}
		}
	}
}

func (c *Cpu) Step() {

	//fetch Instruction
	instr, _ := c.Memory.ReadByteAt(c.PC)
	c.ranCyclesThisFrame += 4 //instr fetch (and decode i guess) takes 4 (machine?) cycles
	//decode/Execute
	c.ranCyclesThisFrame += c.decodeExecute(instr)
}

// returns machine cycles it took to execute
func (c *Cpu) decodeExecute(instr byte) (cycles uint64) {

	fmt.Printf("Instr 0x%02x\n", instr)
	switch instr {

	//cb
	case 0xcb:
		return c.handleCB()

	//16 Load 16 Bit Imm to Reg
	case 0x31:
		r := c.loadImm16Reg(&c.SP)
		return r
	case 0x21:
		return c.loadImm16Reg2Ptr(&c.H, &c.L)
	case 0x11:
		return c.loadImm16Reg2Ptr(&c.D, &c.E)

	// Load 8 Bit Imm to Reg
	case 0x3e:
		return c.loadImm8Reg(&c.A)
	case 0x06:
		return c.loadImm8Reg(&c.B)
	case 0x0e:
		return c.loadImm8Reg(&c.C)

	// decrement Reg8
	case 0x05:
		return c.decrementReg8(&c.B)

	// increment Reg8
	case 0x0c:
		return c.incrementReg8(&c.C)

	// increment Reg16
	case 0x23:
		return c.incrementReg16(REG_HL)
	//jump
	case 0xC3:
		return c.jump()
	case 0x20:
		return c.jumpRelIf(c.GetZeroFlag() == 0)
		// call
	case 0xcd:
		return c.call16Imm()
	//ret
	case 0xc9:
		//TODO

		readLow, _ := c.Memory.ReadByteAt(c.SP)
		c.SP++

		readHigh, _ := c.Memory.ReadByteAt(c.SP)
		c.SP++
		newPC := (uint16(readLow) | uint16(readHigh)<<8)

		c.PC = newPC

		return 4

	// xor Reg
	case 0xaf:
		return c.xorReg(&c.A)

		//store reg in mem
	case 0x22:
		hl := c.GetHL()
		mc := c.storeRegInMemAddr(hl, c.A)
		c.SetHL(hl + 1)
		return mc
	case 0x32:
		hl := c.GetHL()
		mc := c.storeRegInMemAddr(hl, c.A)
		c.SetHL(hl - 1)
		return mc
	case 0xe2:
		return c.storeRegInMemAddr(IO_START_ADDR+uint16(c.C), c.A)

	case 0x77:
		return c.storeRegInMemAddr(c.GetHL(), c.A)

	//store reg in imm mem
	case 0xe0:
		return c.storeRegInImmMemAddr(c.A)

	//store mem in reg
	case 0x1a:
		return c.storeMemInReg(c.GetDE(), &c.A)

	// store val in Reg
	case 0x4f:
		return c.storeValInReg(&c.C, c.A)

	// push 16
	case 0xf5:
		return c.push16(&c.A, &c.F)
	case 0xc5:
		return c.push16(&c.B, &c.C)

	// pop 16
	case 0xc1:
		return c.pop16(&c.B, &c.C)

	// set ie
	case 0xfb:
		c.PC++
		c.Memory.Io.SetIE(1)
		return 1

	//RLA
	//similiar to cbRegRotateLeft (other numBytes, numCycles and different flags)
	case 0x17:
		c.PC++
		var newCarry bool
		oldCarry := c.GetCarryFlag()

		oldRegVal := c.A
		c.A = oldRegVal << 1

		if oldCarry == 1 {
			c.A |= 1
		} else {
			c.A &^= (1)
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
	instr, numReadBytes := c.Memory.ReadByteAt(c.PC)
	c.PC += numReadBytes

	fmt.Printf("-> 0xcb%02x\n", instr)

	switch instr {
	case 0x7c:
		return c.cbSetZeroToComplementRegBit(&c.H, 7)
	case 0x11:
		return c.cbRegRotateLeft(&c.C)
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

func (c *Cpu) cbRegRotateLeft(regPtr *uint8) uint64 {
	var oldRegVal uint8
	var oldCarry = c.GetCarryFlag()
	var newCarry bool

	oldRegVal = *regPtr
	*regPtr = oldRegVal << 1
	if oldCarry == 1 {
		*regPtr |= 1
	} else {
		*regPtr &^= (1)
	}

	c.SetZeroFlag(*regPtr == 0)
	c.SetSubFlag(false)
	c.SetHalfCarryFlag(false)
	newCarry = ((oldRegVal >> 7 & 1) == 1)
	c.SetCarryFlag(newCarry)

	return 2
}
func (c *Cpu) cbSetZeroToComplementRegBit(regPtr *uint8, bitPos int) uint64 {
	var bit uint8

	bit = *regPtr >> bitPos & 0x1

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

func (c *Cpu) pop16(higherRegPtr *uint8, lowerRegPtr *uint8) (cycles uint64) {
	c.PC++

	readLow, _ := c.Memory.ReadByteAt(c.SP)
	*lowerRegPtr = readLow
	c.SP++

	readHigh, _ := c.Memory.ReadByteAt(c.SP)
	*higherRegPtr = readHigh
	c.SP++

	return 3
}

func (c *Cpu) push16(higherRegPtr *uint8, lowerRegPtr *uint8) (cycles uint64) {

	c.PC++

	c.SP--
	c.Memory.SetValue(c.SP, *higherRegPtr)
	c.SP--
	c.Memory.SetValue(c.SP, *lowerRegPtr)
	return 4
}

func (c *Cpu) storeValInReg(regPtr *uint8, val uint8) (cycles uint64) {
	c.PC++
	*regPtr = val
	return 1
}

// In memory, push the program counter PC value corresponding to the address following the CALL instruction to the 2 bytes
// following the byte specified by the current stack pointer SP. Then load the 16-bit immediate operand a16 into PC.
func (c *Cpu) call16Imm() (cycles uint64) {

	c.PC++
	newPCAddr, bytesRead := c.Memory.Read16At(c.PC)
	c.PC += bytesRead

	// With the push, the current value of SP is decremented by 1, and the higher-order byte of PC is loaded in the
	// memory address specified by the new SP value. The value of SP is then decremented by 1 again, and the lower-order
	//byte of PC is loaded in the memory address specified by that value of SP.
	c.SP--
	c.Memory.SetValue(c.SP, getHigher(c.PC))
	c.SP--
	c.Memory.SetValue(c.SP, getLower(c.PC)) // lower order byte of PC

	//The subroutine is placed after the location specified by the new PC value. When the subroutine finishes, control is
	//returned to the source program using a return instruction and by popping the starting address of the next
	//instruction (which was just pushed) and moving it to the PC.

	c.PC = newPCAddr
	// The lower-order byte of a16 is placed in byte 2 of the object code, and the higher-order byte is placed in byte 3.
	newPCAddrHigher := getHigher(c.PC)
	newPCAddrLower := getLower(c.PC)
	c.Memory.Oam[2] = newPCAddrLower
	c.Memory.Oam[3] = newPCAddrHigher

	return 6
}

func (c *Cpu) storeMemInReg(address uint16, regPtr *uint8) (cycles uint64) {

	val, bytesRead := c.Memory.ReadByteAt(address)
	c.PC += bytesRead

	*regPtr = val

	return 2
}

func (c *Cpu) storeRegInImmMemAddr(val uint8) (cycles uint64) {
	c.PC++
	a8, bytesRead := c.Memory.ReadByteAt(c.PC)
	c.PC += bytesRead
	c.Memory.SetValue(IO_START_ADDR+uint16(a8), val)
	return 3
}

func (c *Cpu) jumpRelIf(cond bool) (cycles uint64) {
	c.PC++

	if cond {
		var data byte
		var bytesRead uint16

		data, bytesRead = c.Memory.ReadByteAt(c.PC)

		signedData := int8(data)

		c.PC += bytesRead + uint16(signedData)
		return 3
	}
	return 2

}
func (c *Cpu) decrementReg8(regPtr *uint8) (cycles uint64) {

	oldRegVal := *regPtr
	*regPtr = oldRegVal - 1

	c.SetZeroFlag(*regPtr == 0)
	c.SetSubFlag(true)
	c.SetHalfCarryFlag(isHalfCarryFlagSubtraction(oldRegVal, 1, *regPtr))

	c.PC++
	return 1
}

func (c *Cpu) incrementReg8(regPtr *uint8) (cycles uint64) {
	oldRegVal := *regPtr
	*regPtr = oldRegVal + 1

	c.SetZeroFlag(*regPtr == 0)
	c.SetSubFlag(false)
	c.SetHalfCarryFlag(isHalfCarryFlagAddition(int(oldRegVal), 1, int(*regPtr)))

	c.PC++
	return 1
}

func (c *Cpu) incrementReg16(reg Reg16) (cycles uint64) {
	c.PC++

	switch reg {
	case REG_AF:
		c.SetAF(c.GetAF() + 1)
	case REG_BC:
		c.SetBC(c.GetBC() + 1)
	case REG_DE:
		c.SetDE(c.GetDE() + 1)
	case REG_HL:
		c.SetHL(c.GetHL() + 1)
	default:
		fmt.Printf("ERROR: Func %s, reg %s not Implemented!", "incrementReg16", reg.String())
	}
	return 2
}

func (c *Cpu) jump() (cycles uint64) {
	c.PC, _ = c.Memory.Read16At(c.PC + 1)
	return 4
}

func (c *Cpu) storeRegInMemAddr(address uint16, toStore uint8) (cycles uint64) {

	c.Memory.SetValue(address, toStore)

	c.PC++
	return 2
}

func (c *Cpu) loadImm8Reg(regPtr *uint8) (cycles uint64) {
	var skip uint16
	var val uint8
	c.PC++
	val, skip = c.Memory.ReadByteAt(c.PC)
	c.PC += skip

	*regPtr = val

	return 2
}

func (c *Cpu) loadImm16Reg(reg *uint16) (cycles uint64) {
	var skip uint16
	var val uint16

	c.PC++
	val, skip = c.Memory.Read16At(c.PC)
	c.PC += skip
	*reg = val

	return 3

}

func (c *Cpu) loadImm16Reg2Ptr(higherRegPtr *uint8, lowerRegPtr *uint8) (cycles uint64) {
	var skip uint16
	var val uint16

	c.PC++
	val, skip = c.Memory.Read16At(c.PC)
	c.PC += skip

	*higherRegPtr = getHigher(val)
	*lowerRegPtr = getLower(val)

	return 3

}

func (c *Cpu) xorReg(regPtr *uint8) (cycles uint64) {

	*regPtr = 0

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
func (c *Cpu) br() {
	c.DumpRegs()
	c.Memory.DumpHram()
	c.Autorun = false
}

func (c *Cpu) DumpRegs() {
	fmt.Printf("Registers:\n\n")
	fmt.Printf("A: 0x%02X\n", c.A)
	fmt.Printf("F: 0x%02X\n", c.F)
	fmt.Printf("B: 0x%02X\n", c.B)
	fmt.Printf("C: 0x%02X\n", c.C)
	fmt.Printf("D: 0x%02X\n", c.D)
	fmt.Printf("E: 0x%02X\n", c.E)
	fmt.Printf("H: 0x%02X\n", c.H)
	fmt.Printf("L: 0x%02X\n", c.L)
	fmt.Printf("SP: 0x%04X\n", c.SP)
	fmt.Printf("PC: 0x%04X\n", c.PC)
}

func (c *Cpu) GetAF() uint16 {
	return uint16(c.A)<<8 | uint16(c.F)
}

func (c *Cpu) GetBC() uint16 {
	return uint16(c.B)<<8 | uint16(c.C)
}

func (c *Cpu) GetDE() uint16 {
	return uint16(c.D)<<8 | uint16(c.E)

}

func (c *Cpu) GetHL() uint16 {
	return uint16(c.H)<<8 | uint16(c.L)

}

func (c *Cpu) GetZeroFlag() uint8 { //z
	return (c.F >> 0x7) & 0x1
}

func (c *Cpu) GetSubFlag() uint8 { //n
	return (c.F >> 0x6) & 0x1

}

func (c *Cpu) GetHalfCarryFlag() uint8 { //h
	return (c.F >> 0x5) & 0x1

}

func (c *Cpu) GetCarryFlag() uint8 { // c
	return (c.F >> 0x4) & 0x1
}

//Setter

func (c *Cpu) SetAF(value uint16) {
	c.A = uint8(value >> 8)
	c.F = uint8(value)
}

func (c *Cpu) SetBC(value uint16) {
	c.B = uint8(value >> 8)
	c.C = uint8(value)
}

func (c *Cpu) SetDE(value uint16) {
	c.D = uint8(value >> 8)
	c.E = uint8(value)
}

func (c *Cpu) SetHL(value uint16) {
	c.H = uint8(value >> 8)
	c.L = uint8(value)
}
func (c *Cpu) SetZeroFlag(cond bool) { //z
	if cond {
		c.SetAF(c.GetAF() | 1<<7)
	} else {
		c.SetAF(c.GetAF() &^ (1 << 7))
	}
}

func (c *Cpu) SetSubFlag(cond bool) { //n
	if cond {
		c.SetAF(c.GetAF() | 1<<6)
	} else {
		c.SetAF(c.GetAF() &^ (1 << 6))
	}
}

func (c *Cpu) SetHalfCarryFlag(cond bool) { //h

	if cond {
		c.SetAF(c.GetAF() | 1<<5)
	} else {
		c.SetAF(c.GetAF() &^ (1 << 5))
	}
}

func (c *Cpu) SetCarryFlag(cond bool) { // c
	if cond {
		c.SetAF(c.GetAF() | 1<<4)
	} else {
		c.SetAF(c.GetAF() &^ (1 << 4))
	}
}

func (c *Cpu) GetBreakpoints() []uint16 {
	return c.breakpoints
}
