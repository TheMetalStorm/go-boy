package main

import (
	"fmt"
	"go-boy/cpu"
	"image/color"
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

func onRestartButton() {
	c.Restart()
}

func loop() {

	curSizeX, _ := g.SingleWindow().CurrentSize()
	hramRows := makeHramTableFromSlice(c.Memory.Hram[:])
	slices.Reverse(hramRows)

	regColumns := makeRegColumns()

	g.SingleWindow().Layout(
		g.Row(
			g.Label("Control: "),
			g.Button("Start").OnClick(onStartButton),
			g.Button("Stop").OnClick(onStopButton),
			g.Button("Step").OnClick(onStepButton),
			g.Button("Restart Machine").OnClick(onRestartButton),
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
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRowsDebuggable(c.Memory.Bank0[:], 0)...),
				),

				g.TabItem("BankN").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRowsDebuggable(c.Memory.Bank1[:], 0x4000)...),
				),

				g.TabItem("Vram").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Vram[:], 0x8000)...),
				),

				g.TabItem("Extram").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Extram[:], 0xa000)...),
				),

				g.TabItem("Wram1").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Wram1[:], 0xc000)...),
				),

				g.TabItem("Wram2").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Wram2[:], 0xd000)...),
				),

				g.TabItem("Oam").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Oam[:], 0xfe00)...),
				),

				// Todo better view for Io with description
				g.TabItem("Io").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Io.Regs[:], 0xff00)...),
				),

				g.TabItem("Hram").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.Memory.Hram[:], 0xff80)...),
				),

				g.TabItem("Ie").Layout(
					g.Labelf("0xFFFF: %02x", c.Memory.Ie),
				),

				g.TabItem("Game Code").Layout(
					g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(makeHexTableRows(c.GetCurrentGame(), 0)...),
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

func makeHexTableRows(slice []uint8, regionOffset uint32) []*g.TableRowWidget {
	var rowLen int = (len(slice) / 16) + 1
	rows := make([]*g.TableRowWidget, rowLen)
	rowInd := 0

	for hexOffset := 0; hexOffset < len(slice); hexOffset += 16 {

		rows[rowInd] = g.TableRow(
			g.Selectablef("0x%04x: ", int(regionOffset)+hexOffset),
			makeHexRowCell(slice, regionOffset, hexOffset, 0),
			makeHexRowCell(slice, regionOffset, hexOffset, 1),
			makeHexRowCell(slice, regionOffset, hexOffset, 2),
			makeHexRowCell(slice, regionOffset, hexOffset, 3),
			makeHexRowCell(slice, regionOffset, hexOffset, 4),
			makeHexRowCell(slice, regionOffset, hexOffset, 5),
			makeHexRowCell(slice, regionOffset, hexOffset, 6),
			makeHexRowCell(slice, regionOffset, hexOffset, 7),
			makeHexRowCell(slice, regionOffset, hexOffset, 8),
			makeHexRowCell(slice, regionOffset, hexOffset, 9),
			makeHexRowCell(slice, regionOffset, hexOffset, 10),
			makeHexRowCell(slice, regionOffset, hexOffset, 11),
			makeHexRowCell(slice, regionOffset, hexOffset, 12),
			makeHexRowCell(slice, regionOffset, hexOffset, 13),
			makeHexRowCell(slice, regionOffset, hexOffset, 14),
			makeHexRowCell(slice, regionOffset, hexOffset, 15),
		)
		rowInd++
	}

	if rows[len(rows)-1] == nil {
		rows[len(rows)-1] = g.TableRow()
	}
	return rows
}

func makeHexTableRowsDebuggable(slice []uint8, regionOffset uint32) []*g.TableRowWidget {
	var rowLen int = len(slice) / 16
	rows := make([]*g.TableRowWidget, rowLen)
	rowInd := 0

	for hexOffset := 0; hexOffset < len(slice); hexOffset += 16 {
		rows[rowInd] = g.TableRow(
			g.Selectablef("0x%04x: ", int(regionOffset)+hexOffset),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 0),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 1),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 2),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 3),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 4),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 5),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 6),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 7),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 8),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 9),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 10),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 11),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 12),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 13),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 14),
			makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 15),
		)
		rowInd++
	}
	return rows
}

func makeHexRowCell(slice []uint8, regionOffset uint32, hexOffset int, rowOffset int) g.Widget {
	localAddr := uint16(hexOffset) + uint16(rowOffset)

	if int(localAddr) > len(slice)-1 {
		return g.Label(" ")
	}
	return g.Selectablef("0x%02x ", slice[hexOffset+rowOffset])

}

func makeHexRowCellDebugabble(slice []uint8, regionOffset uint32, hexOffset int, rowOffset int) g.Widget {
	localAddr := uint16(hexOffset) + uint16(rowOffset)
	if int(localAddr) > len(slice)-1 {
		return g.Label(" ")
	}
	globalAddr := uint16(regionOffset) + uint16(hexOffset) + uint16(rowOffset)

	style := g.Style()
	if isInDebug(c.GetBreakpoints(), globalAddr) {
		style.SetColor(g.StyleColorButton, color.RGBA{0x00, 0xbb, 0xaa, 0xff})

	}
	if isCurrentPC(c.PC, globalAddr) {
		style.SetColor(g.StyleColorText, color.RGBA{0x00, 0xff, 0x00, 0xff})

	}
	style.SetStyle(g.StyleVarFrameBorderSize, 0, 0)
	style.SetColor(g.StyleColorButtonHovered, color.RGBA{0x00, 0xff, 0x00, 0x00})
	style.SetColor(g.StyleColorButtonActive, color.RGBA{0x00, 0xff, 0x00, 0x00})

	button := g.Buttonf("0x%02x ", slice[localAddr]).OnClick(OnClickMemView(globalAddr))
	return style.To(button)

}

func isCurrentPC(pc uint16, addr uint16) bool {
	return pc == addr
}

func isInDebug(breakpoints []uint16, addr uint16) bool {
	return slices.Contains(breakpoints, addr)
}

func OnClickMemView(addr uint16) func() {
	return func() {
		c.ToggleBP(addr)
	}
}

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

	c.Restart()
	c.Autorun = false
	c.Run()
}
