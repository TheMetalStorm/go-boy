package main

import (
	"go-boy/cpu"
	"go-boy/rom"
)

type Cpu = cpu.Cpu

func main() {

	cpu := cpu.NewCpu()
	//toLoad := rom.NewRom("./games/Tetris.gb")
	bootrom := rom.NewRom("./bootroms/dmg_boot.bin")

	cpu.LoadBootRom(bootrom)
	// cpu.PatchBootRom(bootrom)

	for {
		cpu.Step()

	}
}
