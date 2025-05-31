package emulator

import (
	"go-boy/cpu"
	"go-boy/rom"
	"go-boy/screen"
)

type Rom = rom.Rom
type Screen = screen.Screen
type Cpu = cpu.Cpu

// http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html
var MAX_CYCLES_PER_FRAME uint64 = 69905
var GB_CLOCK_SPEED_HZ uint64 = 4194304
var DIV_REG_INCREMENT_HZ = 16384

type Emulator struct {
	Cpu         *Cpu
	screen      *Screen
	currentGame *Rom

	ranMCyclesThisFrame uint64
}

func NewEmulator() *Emulator {
	emu := &Emulator{}
	emu.Cpu = cpu.NewCpu()
	emu.Restart()
	return emu
}

func (e *Emulator) Restart() {
	e.Cpu.Restart()
	e.ranMCyclesThisFrame = 0

	e.currentGame = rom.NewRom("./games/tetris.gb")
	e.LoadRom(e.currentGame)
}

func (e *Emulator) LoadRom(r *Rom) {
	// TODO: for now only fills bank 0
	for i := 0x0; i <= 0x3fff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		e.Cpu.Memory.SetValue(uint16(i), newVal)
	}
}

func (e *Emulator) Run() {

	for {
		e.Step()
	}
}

func (e *Emulator) Step() {

	ranMCyclesThisStep := e.Cpu.Step()

	e.ranMCyclesThisFrame += ranMCyclesThisStep
	println(e.ranMCyclesThisFrame)
	// TODO
	//e.handleInterrupts()

	e.updateTimers(ranMCyclesThisStep)

	e.refreshScreen()
}

func (e *Emulator) updateTimers(cyclesThisStep uint64) {
	//TODO other timers
	//https://gbdev.io/pandocs/Timer_and_Divider_Registers.html#ff05--tima-timer-counter
	e.updateDivReg(cyclesThisStep)

}

func (e *Emulator) refreshScreen() {
	e.screen.Update()
}

func (e *Emulator) updateDivReg(cyclesThisStep uint64) {
	// if we crossed the 64 M Cycles boundary this Step
	if (e.ranMCyclesThisFrame-cyclesThisStep)%64 != e.ranMCyclesThisFrame%64 {
		read, _ := e.Cpu.Memory.ReadByteAt(0xFF04)
		e.Cpu.Memory.SetValue(0xFF04, read+1)
	}

}

func (e *Emulator) GetCurrentGame() []byte {
	return e.currentGame.GetData()
}

func (e *Emulator) SetScreen(screen *Screen) {
	e.screen = screen
}
