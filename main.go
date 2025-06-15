package main

import (
	g "github.com/AllenDang/giu"
	_ "github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"go-boy/debugger"
	"go-boy/emulator"
	"os"
)

type Emulator = emulator.Emulator
type Debugger = debugger.Debugger

var tests = []string{
	"./testroms/blargg/instr_timing/instr_timing.gb",
	"./testroms/blargg/mem_timing/individual/01-read_timing.gb",
	"./testroms/blargg/mem_timing/individual/02-write_timing.gb",
	"./testroms/blargg/mem_timing/individual/03-modify_timing.gb",

	"./testroms/blargg/cpu_instrs/individual/01-special.gb",
	"./testroms/blargg/cpu_instrs/individual/02-interrupts.gb",
	"./testroms/blargg/cpu_instrs/individual/03-op sp,hl.gb",
	"./testroms/blargg/cpu_instrs/individual/04-op r,imm.gb",
	"./testroms/blargg/cpu_instrs/individual/05-op rp.gb",
	"./testroms/blargg/cpu_instrs/individual/06-ld r,r.gb",
	"./testroms/blargg/cpu_instrs/individual/07-jr,jp,call,ret,rst.gb",
	"./testroms/blargg/cpu_instrs/individual/08-misc instrs.gb",
	"./testroms/blargg/cpu_instrs/individual/09-op r,r.gb",
	"./testroms/blargg/cpu_instrs/individual/10-bit ops.gb",
	"./testroms/blargg/cpu_instrs/individual/11-op a,(hl).gb",
}

func main() {
	//for sdl2
	//runtime.LockOSThread()
	pixelgl.Run(start)
}

func start() {

	var e *Emulator = emulator.NewEmulator()
	isDebugMode := false
	test := false
	argsWithoutProg := os.Args[1:]
	var err error
	var logFile *os.File = nil
	if len(argsWithoutProg) >= 1 {
		if argsWithoutProg[0] == "--debug" {
			isDebugMode = true
		} else if argsWithoutProg[0] == "--test" {
			test = true
		} else if argsWithoutProg[0] == "--log" {
			filename := "gb-log"
			os.Remove(filename)
			logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}

		}

	}
	defer logFile.Close()

	dbg := debugger.NewDebugger()
	dbg.SetEmu(e)

	if isDebugMode {
		go func() {
			wnd := g.NewMasterWindow("GB Debugger", 800, 800, g.MasterWindowFlagsMaximized)
			wnd.Run(debugger.StartLoop(dbg))
		}()
		dbg.RunEmulator()
	} else {
		if test {
			e.RunTests(tests)
		}
		e.Cpu.LogFile = logFile
		e.Run()
	}

}
