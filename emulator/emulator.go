// Aims to be M-Cycle Accurate

package emulator

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"go-boy/cpu"
	"go-boy/ioregs"
	"go-boy/ppu"
	"go-boy/rom"
	"go-boy/screen"
	"os"
	"time"
)

type Rom = rom.Rom
type Screen = screen.Screen
type Cpu = cpu.Cpu

type Ppu = ppu.Ppu

// http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html
var MAX_CYCLES_PER_FRAME uint64 = 69905
var GB_CLOCK_SPEED_HZ uint64 = 4194304
var DIV_REG_INCREMENT_HZ = 16384

var mult int = 5

type Emulator struct {
	Cpu         *Cpu
	Ppu         *Ppu
	currentGame *Rom

	doRender            bool
	ranMCyclesThisFrame uint64

	//nextCycleCopyTmaToTima bool
}

func NewEmulator() *Emulator {
	emu := &Emulator{}
	emu.doRender = true
	emu.Cpu = cpu.NewCpu()
	emu.Ppu = ppu.NewPpu(mult)
	emu.Restart()
	return emu
}

func (e *Emulator) Restart() {
	e.Cpu.Restart()
	e.Ppu.Restart(mult)
	e.ranMCyclesThisFrame = 0

	e.currentGame = rom.NewRom("./games/tetris.gb")

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
	//for {
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		e.SerialOut()
		e.Step()
		rl.EndDrawing()
	}
}

func (e *Emulator) Step() {

	ranMCyclesThisStep := uint64(1)
	ranMCyclesThisStep += e.handleInterrupts()

	if !e.Cpu.Halt {
		ranMCyclesThisStep += e.Cpu.Step()
	}
	e.Cpu.UpdateTimers(ranMCyclesThisStep)
	if e.doRender {
		e.Ppu.Step(e.Cpu)
		e.doRender = false
	}
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
