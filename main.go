package main

import (
	"fmt"
	"go-boy/cpu"
	"go-boy/rom"
	"slices"

	g "github.com/AllenDang/giu"
)

type Cpu = cpu.Cpu

var splitPos float32 = 200

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

	curSizeX, _ := g.SingleWindow().CurrentSize()
	//TODO add Stack
	hramRows := makeHramTableFromSlice(c.Memory.Hram[:])
	slices.Reverse(hramRows)
	bank0Rows := makeHexTable(c.Memory.Bank0[:], 0)
	regColumns := makeRegColumns()

	g.SingleWindow().Layout(
		g.Row(
			g.Label("Control: "),
			g.Button("Start").OnClick(onStartButton),
			g.Button("Stop").OnClick(onStopButton),
			g.Button("Step").OnClick(onStepButton),
		),
		g.Row(
			g.Label("Regs: "),
			g.Table().FastMode(true).Columns(regColumns...).Size(curSizeX, 20),
		),
		g.SplitLayout(g.DirectionVertical, &splitPos,
			g.Column(
				g.Label("Stack: "),
				g.Table().FastMode(true).Rows(hramRows...),
			),
			g.TabBar().TabItems(

				g.TabItem("Bank0").Layout(
					g.Table().Flags(0).FastMode(true).Rows(bank0Rows...),
				),
				g.TabItem("B").Layout(
					g.Label("This is second tab"),
				),
			),
		),
	)

}

func makeHramTableFromSlice(slice []uint8) []*g.TableRowWidget {
	rows := make([]*g.TableRowWidget, len(slice))
	start := 0xff80
	for i, v := range slice {
		rows[i] = g.TableRow(g.Labelf("0x%04x:", start), g.Labelf("0x%02x ", v))
		start++
	}
	return rows
}

func makeHexTable(slice []uint8, add_offset uint32) []*g.TableRowWidget {
	//TODO row num of table
	var rowLen int = len(slice) / 16
	rows := make([]*g.TableRowWidget, rowLen)
	offset := 0x0000
	rowInd := 0

	for i := 0; i < len(slice); i += 16 {

		rows[rowInd] = g.TableRow(
			g.Selectablef("0x%04x: ", offset+int(add_offset)+i),
			// g.Selectablef("0x%02x ", slice[i+0]).Selected(false).OnClick(OnDClickMemView),
			g.Selectablef("0x%02x ", slice[i+1]),
			g.Selectablef("0x%02x ", slice[i+2]),
			g.Selectablef("0x%02x ", slice[i+3]),
			g.Selectablef("0x%02x ", slice[i+4]),
			g.Selectablef("0x%02x ", slice[i+5]),
			g.Selectablef("0x%02x ", slice[i+6]),
			g.Selectablef("0x%02x ", slice[i+7]),
			g.Selectablef("0x%02x ", slice[i+8]),
			g.Selectablef("0x%02x ", slice[i+9]),
			g.Selectablef("0x%02x ", slice[i+10]),
			g.Selectablef("0x%02x ", slice[i+11]),
			g.Selectablef("0x%02x ", slice[i+12]),
			g.Selectablef("0x%02x ", slice[i+13]),
			g.Selectablef("0x%02x ", slice[i+14]),
			g.Selectablef("0x%02x ", slice[i+15]),
		)

		rowInd++
	}
	return rows
}

// func OnDClickMemView() {
// 	c.ToggleBP(slice[i+0])
// }

func makeRegColumns() []*g.TableColumnWidget {
	regColumns := make([]*g.TableColumnWidget, 10)
	regColumns[0] = g.TableColumn(fmt.Sprintf("PC: 0x%04x", c.PC))
	regColumns[1] = g.TableColumn(fmt.Sprintf("SP: 0x%04x", c.SP))
	regColumns[2] = g.TableColumn(fmt.Sprintf("A: 0x%02x", c.A))
	regColumns[3] = g.TableColumn(fmt.Sprintf("F: 0x%02x", c.F))
	regColumns[4] = g.TableColumn(fmt.Sprintf("B: 0x%02x", c.B))
	regColumns[5] = g.TableColumn(fmt.Sprintf("C: 0x%02x", c.C))
	regColumns[6] = g.TableColumn(fmt.Sprintf("D: 0x%02x", c.D))
	regColumns[7] = g.TableColumn(fmt.Sprintf("E: 0x%02x", c.E))
	regColumns[8] = g.TableColumn(fmt.Sprintf("H: 0x%02x", c.H))
	regColumns[9] = g.TableColumn(fmt.Sprintf("L: 0x%02x", c.L))
	return regColumns
}

func main() {

	go func() {
		wnd := g.NewMasterWindow("GB Debugger", 800, 800, g.MasterWindowFlagsMaximized)
		wnd.Run(loop)
	}()
	emulate()

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
