package main

import (
	"go-boy/cpu"
	"go-boy/debugger"
	"os"

	g "github.com/AllenDang/giu"
)

type Cpu = cpu.Cpu
type Debugger = debugger.Debugger

var c *cpu.Cpu = cpu.NewCpu()

func main() {

	isDebugMode := false
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) >= 1 {
		if argsWithoutProg[0] == "--debug" {
			isDebugMode = true
		}

	}

	c.Restart()
	dbg := debugger.NewDebugger()
	dbg.SetCpu(c)

	if isDebugMode {
		go func() {
			wnd := g.NewMasterWindow("GB Debugger", 800, 800, g.MasterWindowFlagsMaximized)
			wnd.Run(debugger.StartLoop(dbg))
		}()
		dbg.RunCpu()
	} else {
		c.Run()
	}

}
