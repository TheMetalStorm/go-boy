// Aims to be M-Cycle Accurate

package emulator

import (
	"go-boy/internal"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Rom = internal.Rom
type Cpu = internal.Cpu

type Ppu = internal.Ppu

// http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html
var MAX_CYCLES_PER_FRAME uint64 = 69905
var GB_CLOCK_SPEED_HZ uint64 = 4194304
var DIV_REG_INCREMENT_HZ = 16384

var screenSizeMultiplier int = 5

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
	emu.Cpu = internal.NewCpu()
	emu.Ppu = internal.NewPpu(screenSizeMultiplier)

	emu.Ppu.Cpu = emu.Cpu
	emu.Cpu.Ppu = emu.Ppu

	emu.Restart()
	return emu
}

func (e *Emulator) Restart() {
	e.Cpu.Restart()
	e.Ppu.Restart(screenSizeMultiplier)
	e.ranMCyclesThisFrame = 0

	e.currentGame = internal.NewRom("./games/Tetris.gb")

	e.LoadRom(e.currentGame)
}

func (e *Emulator) LoadRom(r *Rom) {
	// TODO: for now only fills bank 0 + 1, no  Memory Bank Controllers (MBCs)
	for i := 0x0; i <= 0x7fff; i++ {
		newVal, _ := r.ReadByteAt(uint16(i))
		e.Cpu.Memory.SetValueForRom(uint16(i), newVal)
	}
}

func (e *Emulator) RunTests(tests []string) {

	var startNext bool = false
	go changeBool(&startNext)
	for _, test := range tests {
		startNext = false
		e.Restart()
		e.currentGame = internal.NewRom(test)
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

func (e *Emulator) handleInput() {
	if e.Ppu.Window.GetFlags()&sdl.WINDOW_INPUT_FOCUS != 0 {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent: // NOTE: Please use `*sdl.QuitEvent` for `v0.4.x` (current version).
				println("Quit")
				// running = false
				break
			case *sdl.KeyboardEvent:
				println("Keyboard event", t.Keysym.Sym)

				break
			}
		}
	}
}

func (e *Emulator) Step() {

	e.handleInput()
	ranMCyclesThisStep := uint64(1)
	ranMCyclesThisStep += e.handleInterrupts()

	if !e.Cpu.Halt {
		ranMCyclesThisStep += e.Cpu.Step()
	}
	e.Cpu.UpdateTimers(ranMCyclesThisStep)

	e.Ppu.Step(ranMCyclesThisStep)

	e.ranMCyclesThisFrame += ranMCyclesThisStep

	if e.ranMCyclesThisFrame >= MAX_CYCLES_PER_FRAME {
		e.ranMCyclesThisFrame = 0
		time.Sleep(time.Second / 60)
		e.Ppu.Render()
	}
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
			e.Cpu.Memory.SetValue(e.Cpu.SP, internal.GetHigher8(e.Cpu.PC))
			e.Cpu.SP--
			e.Cpu.Memory.SetValue(e.Cpu.SP, internal.GetLower8(e.Cpu.PC))

			if e.Cpu.Memory.GetInterruptEnabledBit(internal.VBLANK) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.VBLANK) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.VBLANK, false)
				e.Cpu.PC = 0x0040
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.LCD) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.LCD) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.LCD, false)
				e.Cpu.PC = 0x0048
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.TIMER) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.TIMER) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.TIMER, false)
				e.Cpu.PC = 0x0050
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.SERIAL) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.SERIAL) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.SERIAL, false)
				e.Cpu.PC = 0x0058
			} else if e.Cpu.Memory.GetInterruptEnabledBit(internal.JOYPAD) && e.Cpu.Memory.Io.GetInterruptFlagBit(internal.JOYPAD) {
				e.Cpu.Memory.Io.SetInterruptFlagBit(internal.JOYPAD, false)
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
