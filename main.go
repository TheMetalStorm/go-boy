package main

import (
	"go-boy/cpu"
	"go-boy/rom"
	"go-boy/widgets"

	//"gioui.org/app"
	g "github.com/AllenDang/giu"
)

type Cpu = cpu.Cpu
type Split = widgets.Split

var c *cpu.Cpu = cpu.NewCpu()

func onStartButton() {
	c.Autorun = true
}

func onStopButton() {
	c.Autorun = false
}

func onStepButton() {
	c.DoStep = true
}

func loop() {
	g.SingleWindow().Layout(
		g.Row(
			g.Labelf("0x%04x", c.PC),
			g.Button("Start").OnClick(onStartButton),
			g.Button("Stop").OnClick(onStopButton),
			g.Button("Step").OnClick(onStepButton),
		),
	)
}

func main() {

	go func() {
		wnd := g.NewMasterWindow("Hello world", 500, 500, g.MasterWindowFlagsNotResizable)
		wnd.Run(loop)
	}()
	emulate()
	//app.Main()

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

// func run(window *app.Window) error {
// 	window.Option(app.Size(unit.Dp(400), unit.Dp(600)))

// 	theme := material.NewTheme()
// 	var ops op.Ops
// 	var runButton widget.Clickable
// 	var stopButton widget.Clickable
// 	var stepButton widget.Clickable
// 	var regListState widget.List
// 	regListState.Axis = layout.Vertical

// 	var split Split
// 	split.MaxHeight = unit.Dp(200)
// 	stepClickedNow := false

// 	column := layout.Flex{
// 		Axis:    layout.Horizontal,
// 		Spacing: layout.SpaceEvenly,
// 	}
// 	for {

// 		switch e := window.Event().(type) {
// 		case app.DestroyEvent:
// 			return e.Err
// 		case app.FrameEvent:
// 			gtx := app.NewContext(&ops, e)
// 			// Let's try out the flexbox layout:
// 			layout.Flex{
// 				// Vertical alignment, from top to bottom
// 				Axis: layout.Vertical,
// 				// Empty space is left at the start, i.e. at the top
// 				Spacing: layout.SpaceStart,
// 			}.Layout(gtx,
// 				layout.Rigid(
// 					func(gtx layout.Context) layout.Dimensions {
// 						list := [10]string{}
// 						list[0] = fmt.Sprintf("Reg A: 0x%04x", c.A)
// 						list[1] = fmt.Sprintf("Reg F: 0x%04x (0b%08b)", c.F, c.F)
// 						list[2] = fmt.Sprintf("Reg B: 0x%04x", c.B)
// 						list[3] = fmt.Sprintf("Reg C: 0x%04x", c.C)
// 						list[4] = fmt.Sprintf("Reg D: 0x%04x", c.D)
// 						list[5] = fmt.Sprintf("Reg E: 0x%04x", c.E)
// 						list[6] = fmt.Sprintf("Reg H: 0x%04x", c.H)
// 						list[7] = fmt.Sprintf("Reg L: 0x%04x", c.L)
// 						list[8] = fmt.Sprintf("SP: 0x%04x", c.SP)
// 						list[9] = fmt.Sprintf("PC: 0x%04x", c.PC)

// 						return material.List(theme, &regListState).Layout(gtx, 10, func(gtx layout.Context, index int) layout.Dimensions {
// 							return layout.Stack{}.Layout(gtx,
// 								layout.Stacked(func(gtx layout.Context) layout.Dimensions {

// 									return layout.UniformInset(unit.Dp(8)).Layout(gtx, material.Body1(theme, list[index]).Layout)
// 								}),
// 							)
// 						})
// 					},
// 				),

// 				layout.Rigid(
// 					func(gtx layout.Context) layout.Dimensions {

// 						return split.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
// 							a := material.H1(theme, "Hello")
// 							return a.Layout(gtx)
// 						}, func(gtx layout.Context) layout.Dimensions {
// 							a := material.H1(theme, "World")
// 							return a.Layout(gtx)
// 						})
// 					},
// 				),
// 				layout.Rigid(
// 					func(gtx layout.Context) layout.Dimensions {
// 						return column.Layout(gtx,
// 							layout.Rigid(
// 								func(gtx layout.Context) layout.Dimensions {
// 									btn := material.Button(theme, &runButton, "Start")
// 									return btn.Layout(gtx)
// 								},
// 							),
// 							layout.Rigid(
// 								func(gtx layout.Context) layout.Dimensions {
// 									btn := material.Button(theme, &stepButton, "Step")
// 									return btn.Layout(gtx)
// 								},
// 							),
// 							layout.Rigid(
// 								func(gtx layout.Context) layout.Dimensions {
// 									btn := material.Button(theme, &stopButton, "Stop")
// 									return btn.Layout(gtx)
// 								},
// 							),
// 						)
// 					},
// 				),

// 				// ... then one to hold an empty spacer
// 				layout.Rigid(
// 					// The height of the spacer is 25 Device independent pixels
// 					layout.Spacer{Height: unit.Dp(25)}.Layout,
// 				),
// 			)

// 			if stepButton.Pressed() && !stepClickedNow {
// 				c.DoStep = true
// 				stepClickedNow = true
// 			} else if !stepButton.Pressed() {
// 				stepClickedNow = false
// 			}

// 			if runButton.Pressed() {
// 				c.Autorun = true
// 			}
// 			if stopButton.Pressed() {
// 				c.Autorun = false
// 			}
// 			e.Frame(gtx.Ops)

// 		}

// 	}
// }
