package main

import (
	"fmt"
	"go-boy/cpu"
	"go-boy/rom"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
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

	//toLoad := rom.NewRom("./games/Tetris.gb")
	bootrom := rom.NewRom("./bootroms/dmg_boot.bin")

	c.LoadBootRom(bootrom)
	// cpu.PatchBootRom(bootrom)
	c.Autorun = false

	for {
		if c.Autorun {
			c.Step()
		} else {
			if c.DoStep {
				c.Step()
				c.DoStep = false
			}
		}
	}

}

func run(window *app.Window) error {
	window.Option(app.Size(unit.Dp(400), unit.Dp(600)))

	theme := material.NewTheme()
	var ops op.Ops
	var runButton widget.Clickable
	var stepButton widget.Clickable

	stepClickedNow := false

	column := layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}
	for {

		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
			// case app.FrameEvent:
			// 	// This graphics context is used for managing the rendering state.
			// 	gtx := app.NewContext(&ops, e)
			// 	stepButton := material.Button(theme, &stepButton, "Step")

			// 	autoRunButton := material.Button(theme, &runButton, "Run")

			// 	// Define an large label with an appropriate text:
			// 	var s string = strconv.FormatUint(uint64(c.PC), 10)
			// 	title := material.H1(theme, "PC: "+s)

			// 	// Change the color of the label.
			// 	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
			// 	title.Color = maroon

			// 	// Change the position of the label.
			// 	title.Alignment = text.Middle

			// 	// Draw the label to the graphics context.
			// 	title.Layout(gtx)
			// 	stepButton.Layout(gtx)
			// 	autoRunButton.Layout(gtx)
			// 	// title := material.Button(theme, )
			// 	// Pass the drawing operations to the GPU.
			// 	e.Frame(gtx.Ops)

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			// Let's try out the flexbox layout:
			layout.Flex{
				// Vertical alignment, from top to bottom
				Axis: layout.Vertical,
				// Empty space is left at the start, i.e. at the top
				Spacing: layout.SpaceStart,
			}.Layout(gtx,

				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						format := fmt.Sprintf("%04x", c.PC)
						btn := material.H4(theme, "PC :0x"+format)
						return btn.Layout(gtx)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						return column.Layout(gtx,
							layout.Rigid(
								func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(theme, &runButton, "Start")
									return btn.Layout(gtx)
								},
							),
							layout.Rigid(
								func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(theme, &stepButton, "Step")
									return btn.Layout(gtx)
								},
							),
						)
					},
				),

				// ... then one to hold an empty spacer
				layout.Rigid(
					// The height of the spacer is 25 Device independent pixels
					layout.Spacer{Height: unit.Dp(25)}.Layout,
				),
			)

			if stepButton.Pressed() && !stepClickedNow {
				c.DoStep = true
				stepClickedNow = true
			} else if !stepButton.Pressed() {
				stepClickedNow = false
			}

			if runButton.Pressed() {
				c.Autorun = true
			}
			e.Frame(gtx.Ops)

		}

	}
}
