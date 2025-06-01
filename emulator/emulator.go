// Aims to be M-Cycle Accurate

package emulator

import (
	"go-boy/cpu"
	"go-boy/ioregs"
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

	RanMCyclesThisFrame uint64
}

func NewEmulator() *Emulator {
	emu := &Emulator{}
	emu.Cpu = cpu.NewCpu()
	emu.Restart()
	return emu
}

func (e *Emulator) Restart() {
	e.Cpu.Restart()
	e.RanMCyclesThisFrame = 0

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
	ranMCyclesThisStep += e.handleInterrupts()

	e.updateTimers(ranMCyclesThisStep)

	//e.refreshScreen()
	e.RanMCyclesThisFrame += ranMCyclesThisStep

}

// TODO: The effect of ei is delayed by one instruction. This means that ei followed immediately by di does not allow any interrupts between them.
// This interacts with the halt bug in an interesting way.
// https://gbdev.io/pandocs/Interrupts.html#ffff--ie-interrupt-enable
func (e *Emulator) handleInterrupts() uint64 {
	if e.Cpu.IME {
		requestedInterrupts := e.Cpu.Memory.Io.GetIF()
		enabledInterrupts := e.Cpu.Memory.Ie

		if requestedInterrupts&enabledInterrupts > 0 {
			e.Cpu.IME = false

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
			return 5
		}
	}
	return 0
}

func (e *Emulator) refreshScreen() {
	e.screen.Update()
}

func (e *Emulator) updateTimers(mCyclesThisStep uint64) {
	//TODO other timers
	//https://gbdev.io/pandocs/Timer_and_Divider_Registers.html#ff05--tima-timer-counter
	e.updateDivReg(mCyclesThisStep)
	e.updateTimaReg(mCyclesThisStep)

}

// NOTE: CGB different
// https://gbdev.gg8.se/wiki/articles/Timer_and_Divider_Registers
func (e *Emulator) updateDivReg(mCyclesThisStep uint64) {
	// if we crossed the 64 M Cycles boundary this Step
	if (int(e.RanMCyclesThisFrame+mCyclesThisStep) / 64) != int(e.RanMCyclesThisFrame/64) {

		div := e.Cpu.Memory.Io.GetDIV()
		e.Cpu.Memory.SetValue(0xFF04, div+1)
	}
}

func (e *Emulator) updateTimaReg(mCyclesThisStep uint64) {

	tima := uint16(e.Cpu.Memory.Io.GetTIMA())
	updatFreq := uint64(0)
	timerControl := e.Cpu.Memory.Io.GetTAC()
	timerEnabled := getBit(timerControl, 2)
	if timerEnabled {
		switch timerControl & 0x03 {
		case 0x00: // 00
			updatFreq = 256
		case 0x01: // 01
			updatFreq = 4
		case 0x02: // 10
			updatFreq = 16
		case 0x03: // 11
			updatFreq = 64
		}

		if int(e.RanMCyclesThisFrame+mCyclesThisStep)/int(updatFreq) != int(int(e.RanMCyclesThisFrame)/int(updatFreq)) {
			incr := int(e.RanMCyclesThisFrame+mCyclesThisStep)/int(updatFreq) - int(int(e.RanMCyclesThisFrame)/int(updatFreq))
			tima += uint16(incr)
			if tima > 255 { //tima overflow
				tima = uint16(e.Cpu.Memory.Io.GetTMA())
				e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.TIMER, true)

			}
			e.Cpu.Memory.Io.SetTIMA(uint8(tima))
		}
	}
}

func (e *Emulator) GetCurrentGame() []byte {
	return e.currentGame.GetData()
}

func (e *Emulator) SetScreen(screen *Screen) {
	e.screen = screen
}

func getBit(num uint8, bit uint8) bool {
	res := (num >> bit) & 1
	return res != 0
}
