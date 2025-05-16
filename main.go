package main

import (
	"go-boy/cpu"
	"go-boy/rom"
)

type Cpu = cpu.Cpu

func main() {

	cpu := cpu.NewCpu()
	rom := rom.NewRom("./games/Tetris.gb")

	cpu.LoadRom(rom)

	for {
		cpu.Step()
	}
}
