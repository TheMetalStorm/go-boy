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

	IME          bool
	setIMETrueIn int
	pendingIME   bool

	Halt    bool
	HaltBug bool
	Stop    bool
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
	cpu.Halt = false
	cpu.Stop = false
	cpu.IME = false
	cpu.pendingIME = false
	cpu.setIMETrueIn = 0

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

	if !cpu.Halt {
		cpu.log()
	}

	var ranMCyclesThisStep uint64 = 1 //instr fetch  takes 1 m cycles
	//decode/Execute
	ranMCyclesThisStep += cpu.decodeExecute(instr)

	if cpu.pendingIME {
		if cpu.setIMETrueIn > 0 {
			cpu.setIMETrueIn--
		} else {
			cpu.IME = true
			cpu.pendingIME = false
		}
	}
	return ranMCyclesThisStep
}

func (cpu *Cpu) log() {
	f, err := os.OpenFile("gb-log2", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	pc1, _ := cpu.Memory.ReadByteAt(cpu.PC)
	pc2, _ := cpu.Memory.ReadByteAt(cpu.PC + 1)
	pc3, _ := cpu.Memory.ReadByteAt(cpu.PC + 2)
	pc4, _ := cpu.Memory.ReadByteAt(cpu.PC + 3)

	toWrite := fmt.Sprintf("A:%02x F:%02x B:%02x C:%02x D:%02x E:%02x H:%02x L:%02x SP:%04x PC:%04x PCMEM:%02x,%02x,%02x,%02x\n",
		cpu.A, cpu.F, cpu.B, cpu.C, cpu.D, cpu.E, cpu.H, cpu.L, cpu.SP, cpu.PC, pc1, pc2, pc3, pc4)
	if _, err = f.WriteString(toWrite); err != nil {
		panic(err)
	}
}

func isCarryFlagSubtraction(valA uint8, valB uint8) bool {

	return valB > valA
}

func isCarryFlagSubtraction16(valA uint16, valB uint16) bool {

	return valB > valA
}

func isCarryFlagAddition(valA uint8, valB uint8) bool {

	var add uint16 = uint16(valA) + uint16(valB)

	return (add) > 0xFF
}

func isCarryFlagAddition16(valA uint16, valB uint16) bool {

	var add uint32 = uint32(valA) + uint32(valB)

	return (add) > 0xFFFF
}

func isHalfCarryFlagSubtraction(valA uint8, valB uint8) bool {

	lowerA := getLower4(valA)
	lowerB := getLower4(valB)

	return lowerB > lowerA
}

func isHalfCarryFlagSubtraction16(valA uint16, valB uint16) bool {

	tempA := valA & 0xFFF
	tempB := valB & 0xFFF

	return tempB > tempA

}

func isHalfCarryFlagAddition(valA uint8, valB uint8) bool {

	lowerA := getLower4(valA)
	lowerB := getLower4(valB)

	return (lowerA + lowerB) > 0xF
}

func isHalfCarryFlagAddition16(valA uint16, valB uint16) bool {

	tempA := valA & 0xFFF
	tempB := valB & 0xFFF

	return (uint32(tempA) + uint32(tempB)) > 0xFFF
}

func GetHigher8(orig uint16) uint8 {
	return uint8(orig >> 8 & 0xFF)
}

func GetLower8(orig uint16) uint8 {
	return uint8(orig)
}

func getHigher4(orig uint8) uint8 {
	return uint8((orig >> 4) & 0x0F)
}

func getLower4(orig uint8) uint8 {
	return uint8(orig & 0x0f)
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
