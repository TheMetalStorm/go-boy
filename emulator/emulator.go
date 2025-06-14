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
	timerTotalMCycles   uint64
	divMCycleCounter    uint64
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

	e.updateTimers(ranMCyclesThisStep)

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

func (e *Emulator) updateTimers(mCyclesThisStep uint64) {

	e.updateDivReg(mCyclesThisStep)
	e.updateTimaReg(mCyclesThisStep)

}

// NOTE: CGB different
// https://gbdev.gg8.se/wiki/articles/Timer_and_Divider_Registers
func (e *Emulator) updateDivReg(mCyclesThisStep uint64) {
	// DIV increments every 64 M-cycles (256 T-cycles)
	const divRate = 64

	totalMCycles := e.divMCycleCounter + mCyclesThisStep
	incr := totalMCycles / divRate             // Number of DIV increments needed
	remainingMCycles := totalMCycles % divRate // Carry over unused cycles

	if incr > 0 {
		div := e.Cpu.Memory.Io.GetDIV()
		e.Cpu.Memory.SetValue(0xFF04, div+uint8(incr)) // Increment DIV
	}

	e.divMCycleCounter = remainingMCycles // Save for next step
}

func (e *Emulator) updateTimaReg(mCyclesThisStep uint64) {
	tac := e.Cpu.Memory.Io.GetTAC()
	if (tac & 0x04) == 0 { // Timer disabled
		return
	}

	// Get current TIMA and divider rate
	tima := uint16(e.Cpu.Memory.Io.GetTIMA())
	divRate := e.getTimerDivRate(tac) // Helper function (see below)

	// Calculate how many times TIMA should increment this step
	totalMCycles := e.timerTotalMCycles + mCyclesThisStep
	incr := totalMCycles / divRate             // Full increments
	remainingMCycles := totalMCycles % divRate // Carry over

	// Apply increments
	if incr > 0 {
		tima += uint16(incr)
		if tima > 0xFF { // Handle overflow
			tima = uint16(e.Cpu.Memory.Io.GetTMA())

			e.Cpu.Memory.Io.SetInterruptFlagBit(ioregs.TIMER, true) // Set IF.2
		}
		e.Cpu.Memory.Io.SetTIMA(uint8(tima))
	}

	// Save remaining cycles for next step
	e.timerTotalMCycles = remainingMCycles
}

// Helper: Convert TAC to M-cycle divider rate
func (e *Emulator) getTimerDivRate(tac uint8) uint64 {
	switch tac & 0x03 {
	case 0x00:
		return 256 // 1024 T-cycles = 256 M-cycles
	case 0x01:
		return 4 // 16 T-cycles = 4 M-cycles
	case 0x02:
		return 16 // 64 T-cycles = 16 M-cycles
	case 0x03:
		return 64 // 256 T-cycles = 64 M-cycles
	}
	return 0 // Invalid
}

func (e *Emulator) GetCurrentGame() []byte {
	return e.currentGame.GetData()
}

func (e *Emulator) SetScreen(screen *Screen) {
	e.screen = screen
}
