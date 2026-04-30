package main

import (
	"go-boy/debugger"
	"go-boy/emulator"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
)

type Emulator = emulator.Emulator

//type Debugger = debugger.Debugger

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

	isDebugMode := false
	test := false
	profile := false
	argsWithoutProg := os.Args[1:]
	var logFile *os.File
	for _, arg := range argsWithoutProg {
		switch arg {
		case "--debug":
			isDebugMode = true
		case "--test":
			test = true
		case "--log":
			filename := "gb-log"
			os.Remove(filename)
			var err error
			logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}
		case "--profile":
			profile = true
		}
	}

	if profile {
		f, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)
		go func() {
			<-sigCh
			pprof.StopCPUProfile()
			f.Close()
			os.Exit(0)
		}()

		defer func() {
			signal.Stop(sigCh)
			pprof.StopCPUProfile()
		}()
	}

	if logFile != nil {
		defer logFile.Close()
	}

	var e *Emulator = emulator.NewEmulator()

	if isDebugMode {
		dbg := debugger.NewDebugger()
		dbg.SetEmu(e)
		dbg.RunEmulator()

	} else {

		if test {
			e.RunTests(tests)
		}
		e.Cpu.LogFile = logFile
		e.Run()

	}

}
