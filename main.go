package main

import (
	"bufio"
	"go-boy/cpu"
	"go-boy/rom"
	"image/color"
	"log"
	"os"
	"strconv"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

type Cpu = cpu.Cpu

var c *cpu.Cpu = cpu.NewCpu()

func main() {

	go func() {
		window := new(app.Window)
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	emulate()
	app.Main()

}

func emulate() {
	for {
		reader := bufio.NewReader(os.Stdin)
		//toLoad := rom.NewRom("./games/Tetris.gb")
		bootrom := rom.NewRom("./bootroms/dmg_boot.bin")

		c.LoadBootRom(bootrom)
		// cpu.PatchBootRom(bootrom)
		c.Autorun = false
		for {
			if c.Autorun {
				c.Step()
			} else {
				text, _, _ := reader.ReadRune()
				if text == 'n' {
					c.Step()
				} else if text == 'a' {
					c.Autorun = true
				}
			}
		}
	}
}

func run(window *app.Window) error {

	theme := material.NewTheme()
	var ops op.Ops

	for {

		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			// Define an large label with an appropriate text:
			var s string = strconv.FormatUint(uint64(c.PC), 10)
			title := material.H1(theme, s)

			// Change the color of the label.
			maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
			title.Color = maroon

			// Change the position of the label.
			title.Alignment = text.Middle

			// Draw the label to the graphics context.
			title.Layout(gtx)

			// title := material.Button(theme, )
			// Pass the drawing operations to the GPU.
			e.Frame(gtx.Ops)
		}

	}
}
