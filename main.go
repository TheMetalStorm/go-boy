package main

import (
	"go-boy/debugger"
	"go-boy/emulator"
	"os"

	g "github.com/AllenDang/giu"
)

type Emulator = emulator.Emulator
type Debugger = debugger.Debugger

var e *Emulator = emulator.NewEmulator()

func main() {

	isDebugMode := false
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) >= 1 {
		if argsWithoutProg[0] == "--debug" {
			isDebugMode = true
		}

	}

	e.Restart()

	dbg := debugger.NewDebugger()
	dbg.SetEmu(e)

	if isDebugMode {
		go func() {
			wnd := g.NewMasterWindow("GB Debugger", 800, 800, g.MasterWindowFlagsMaximized)
			wnd.Run(debugger.StartLoop(dbg))
		}()
		dbg.RunEmulator()
	} else {
		e.Run()
	}

}
