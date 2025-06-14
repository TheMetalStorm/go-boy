// Aims to be M-Cycle Accurate

package emulator

import (
	"fmt"
	"go-boy/cpu"
	"go-boy/ioregs"
	"go-boy/rom"
	"go-boy/screen"
	"os"
	"time"
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

	//nextCycleCopyTmaToTima bool
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

	e.currentGame = rom.NewRom("./testroms/blargg/cpu_instrs/individual/02-interrupts.gb")

	//e.currentGame = rom.NewRom("./games/tetris.gb")
	e.LoadRom(e.currentGame)
}

func (e *Emulator) LoadRom(r *Rom) {
	// TODO: for now only fills bank 0 + 1, no  Memory Bank Controllers (MBCs)
	for i := 0x0; i <= 0x7fff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		e.Cpu.Memory.SetValue(uint16(i), newVal)
	}
}

func (e *Emulator) RunTests(tests []string) {

	var startNext bool = false
	go changeBool(&startNext)
	for _, test := range tests {
		fmt.Printf("Running test: %s\n", test)
		startNext = false
		e.Restart()
		e.currentGame = rom.NewRom(test)
		e.LoadRom(e.currentGame)
		for {
			if startNext {
				println()
				break
			}
			e.SerialOut()
			e.Step()
		}
	}
	os.Exit(0)

}

func changeBool(startNextTest *bool) {
	for range time.Tick(time.Second * 2) {
		println("")
		*startNextTest = true
	}
}

func (e *Emulator) Run() {
	for {

		e.SerialOut()
		e.Step()
	}
}

func (e *Emulator) Step() {

	ranMCyclesThisStep := uint64(1)
	ranMCyclesThisStep += e.handleInterrupts()

	if !e.Cpu.Halt {
		ranMCyclesThisStep += e.Cpu.Step()
	}

	e.Cpu.UpdateTimers(ranMCyclesThisStep)

	//e.refreshScreen()
	e.ranMCyclesThisFrame += ranMCyclesThisStep

}
func (e *Emulator) handleInterrupts() uint64 {
	requestedInterrupts := e.Cpu.Memory.Io.GetIF()
	enabledInterrupts := e.Cpu.Memory.GetIe()
	activeInterrupts := requestedInterrupts & enabledInterrupts & 0x1f
	if activeInterrupts != 0 {
		e.Cpu.Halt = false
	}

	if e.Cpu.IME {
		if activeInterrupts != 0 {
			e.Cpu.SP--
			e.Cpu.Memory.SetValue(e.Cpu.SP, cpu.GetHigher8(e.Cpu.PC))
			e.Cpu.SP--
			e.Cpu.Memory.SetValue(e.Cpu.SP, cpu.GetLower8(e.Cpu.PC))

			if e.Cpu.Memory.GetInterruptEnabledBit(ioregs.VBLANK) && e.Cpu.Memory.Io.GetInterruptFlagBit(ioregs.VBLANK) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.VBLANK, false)
				e.Cpu.PC = 0x0040
			} else if e.Cpu.Memory.GetInterruptEnabledBit(ioregs.LCD) && e.Cpu.Memory.Io.GetInterruptFlagBit(ioregs.LCD) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.LCD, false)
				e.Cpu.PC = 0x0048
			} else if e.Cpu.Memory.GetInterruptEnabledBit(ioregs.TIMER) && e.Cpu.Memory.Io.GetInterruptFlagBit(ioregs.TIMER) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.TIMER, false)
				e.Cpu.PC = 0x0050
			} else if e.Cpu.Memory.GetInterruptEnabledBit(ioregs.SERIAL) && e.Cpu.Memory.Io.GetInterruptFlagBit(ioregs.SERIAL) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.SERIAL, false)
				e.Cpu.PC = 0x0058
			} else if e.Cpu.Memory.GetInterruptEnabledBit(ioregs.JOYPAD) && e.Cpu.Memory.Io.GetInterruptFlagBit(ioregs.JOYPAD) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.JOYPAD, false)
				e.Cpu.PC = 0x0060
			}
			e.Cpu.IME = false
			return 5
		}
	}
	return 0
}

func (e *Emulator) refreshScreen() {
	e.screen.Update()
}

func (e *Emulator) SerialOut() {
	read, _ := e.Cpu.Memory.ReadByteAt(0xff02)
	if read == 0x81 {
		ch, _ := e.Cpu.Memory.ReadByteAt(0xff01)
		print(string(ch))
		e.Cpu.Memory.SetValue(0xff02, 0x00)

	}
}

func (e *Emulator) GetCurrentGame() []byte {
	return e.currentGame.GetData()
}

func (e *Emulator) SetScreen(screen *Screen) {
	e.screen = screen
}
