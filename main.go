package main

import (
	"go-boy/debugger"
	"go-boy/emulator"
	"os"
	"runtime"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
	"github.com/AllenDang/cimgui-go/imgui"
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
	runtime.LockOSThread()
	// f, _ := os.Create("cpu.prof")

	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

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

	b, _ := backend.CreateBackend(sdlbackend.NewSDLBackend())

	if logFile != nil {
		defer logFile.Close()
	}
	b.SetBgColor(imgui.NewVec4(0.1, 0.1, 0.1, 1.0))
	b.SetWindowFlags(sdlbackend.SDLWindowFlagsTransparent, 1)

	//Since Imgui created our SDL Context, we always have to create at least one window through ImGUI for sdl (events) to work
	b.CreateWindow("GB Debugger", 1200, 900)

	if isDebugMode {
		dbg := debugger.NewDebugger()
		dbg.SetEmu(e)

		go dbg.RunEmulator()
		// dbg.CreateTileViewer()
		b.Run(func() {
			e.Ppu.Step(e.Cpu)
			dbg.Render()
		})

	} else {
		if test {
			e.RunTests(tests)
		}
		e.Cpu.LogFile = logFile
		go e.Run()
		e.Ppu.Render(e.Cpu)

	}

}
