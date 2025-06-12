package debugger

import (
	"fmt"
	"go-boy/emulator"
	"image/color"
	"slices"

	g "github.com/AllenDang/giu"
)

type Emulator = emulator.Emulator

type Debugger struct {
	autorun     bool
	doStep      bool
	breakpoints []uint16
	lastBPHit   int

	e *Emulator
}

var splitPos float32 = 200

func NewDebugger() *Debugger {
	dbg := &Debugger{}
	dbg.reset()
	return dbg
}

func (d *Debugger) reset() {
	d.autorun = false
	d.doStep = false
	d.lastBPHit = -1
	d.breakpoints = nil
}

func (d *Debugger) GetBreakpoints() []uint16 {
	return d.breakpoints
}

func (d *Debugger) ToggleBP(addr uint16) {
	//No point in setting BP on Mem Adress 0 since we start here with Autorun == false
	if addr == 0 {
		return
	}
	for i, b := range d.breakpoints {
		if b == addr {
			d.breakpoints = append(d.breakpoints[:i], d.breakpoints[i+1:]...)
			return
		}
	}
	d.breakpoints = append(d.breakpoints, addr)
}

func (d *Debugger) onStartButton() {
	d.autorun = true
}

func (d *Debugger) onStopButton() {
	d.autorun = false
}

func (d *Debugger) onStepButton() {
	d.doStep = true
}

func (d *Debugger) onRestartButton() {
	d.autorun = false
	d.e.Restart()
}

func StartLoop(d *Debugger) func() {
	return func() {
		curSizeX, _ := g.SingleWindow().CurrentSize()
		hramRows := makeHramTableFromSlice(d.e.Cpu.Memory.Hram[:])
		slices.Reverse(hramRows)

		regColumns := d.makeRegColumns()

		g.SingleWindow().Layout(
			g.Row(
				g.Label("Control: "),
				g.Button("Start").OnClick(d.onStartButton),
				g.Button("Stop").OnClick(d.onStopButton),
				g.Button("Step").OnClick(d.onStepButton),
				g.Button("Restart Machine").OnClick(d.onRestartButton),
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
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRowsDebuggable(d.e.Cpu.Memory.GetBank0(), 0)...),
					),

					g.TabItem("BankN").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRowsDebuggable(d.e.Cpu.Memory.GetBank1(), 0x4000)...),
					),

					g.TabItem("Vram").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRows(d.e.Cpu.Memory.GetVram(), 0x8000)...),
					),

					g.TabItem("Extram").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRows(d.e.Cpu.Memory.GetExtram(), 0xa000)...),
					),

					g.TabItem("Wram1").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRowsDebuggable(d.e.Cpu.Memory.GetWram1(), 0xc000)...),
					),

					g.TabItem("Wram2").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRowsDebuggable(d.e.Cpu.Memory.GetWram2(), 0xd000)...),
					),

					g.TabItem("Oam").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRows(d.e.Cpu.Memory.GetOam(), 0xfe00)...),
					),

					// Todo better view for Io with description
					g.TabItem("Io").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRows(d.e.Cpu.Memory.Io.Regs[:], 0xff00)...),
					),

					g.TabItem("Hram").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRows(d.e.Cpu.Memory.GetHram(), 0xff80)...),
					),

					g.TabItem("Ie").Layout(
						g.Labelf("0xFFFF: %02x", d.e.Cpu.Memory.Ie),
					),

					g.TabItem("Game Code").Layout(
						g.Table().Flags(g.TableFlagsRowBg).FastMode(true).Rows(d.makeHexTableRows(d.e.GetCurrentGame(), 0)...),
					),
				),
			),
		)
	}
}

func Loop() {

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

func (d *Debugger) makeHexTableRows(slice []uint8, regionOffset uint32) []*g.TableRowWidget {
	var rowLen int = (len(slice) / 16) + 1
	rows := make([]*g.TableRowWidget, rowLen)
	rowInd := 0

	for hexOffset := 0; hexOffset < len(slice); hexOffset += 16 {

		rows[rowInd] = g.TableRow(
			g.Selectablef("0x%04x: ", int(regionOffset)+hexOffset),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 0),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 1),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 2),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 3),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 4),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 5),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 6),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 7),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 8),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 9),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 10),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 11),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 12),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 13),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 14),
			d.makeHexRowCell(slice, regionOffset, hexOffset, 15),
		)
		rowInd++
	}

	if rows[len(rows)-1] == nil {
		rows[len(rows)-1] = g.TableRow()
	}
	return rows
}

func (d *Debugger) makeHexTableRowsDebuggable(slice []uint8, regionOffset uint32) []*g.TableRowWidget {
	var rowLen int = len(slice) / 16
	rows := make([]*g.TableRowWidget, rowLen)
	rowInd := 0

	for hexOffset := 0; hexOffset < len(slice); hexOffset += 16 {
		rows[rowInd] = g.TableRow(
			g.Selectablef("0x%04x: ", int(regionOffset)+hexOffset),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 0),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 1),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 2),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 3),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 4),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 5),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 6),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 7),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 8),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 9),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 10),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 11),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 12),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 13),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 14),
			d.makeHexRowCellDebugabble(slice, regionOffset, hexOffset, 15),
		)
		rowInd++
	}
	return rows
}

func (d *Debugger) makeHexRowCell(slice []uint8, regionOffset uint32, hexOffset int, rowOffset int) g.Widget {
	localAddr := uint16(hexOffset) + uint16(rowOffset)

	if int(localAddr) > len(slice)-1 {
		return g.Label(" ")
	}
	return g.Selectablef("0x%02x ", slice[hexOffset+rowOffset])

}

func (d *Debugger) makeHexRowCellDebugabble(slice []uint8, regionOffset uint32, hexOffset int, rowOffset int) g.Widget {
	localAddr := uint16(hexOffset) + uint16(rowOffset)
	if int(localAddr) > len(slice)-1 {
		return g.Label(" ")
	}
	globalAddr := uint16(regionOffset) + uint16(hexOffset) + uint16(rowOffset)

	style := g.Style()

	if d.isInDebug(globalAddr) {
		style.SetColor(g.StyleColorButton, color.RGBA{0x00, 0xbb, 0xaa, 0xff})

	}
	if d.isCurrentPC(globalAddr) {
		style.SetColor(g.StyleColorText, color.RGBA{0x00, 0xff, 0x00, 0xff})

	}
	style.SetStyle(g.StyleVarFrameBorderSize, 0, 0)
	style.SetColor(g.StyleColorButtonHovered, color.RGBA{0x00, 0xff, 0x00, 0x00})
	style.SetColor(g.StyleColorButtonActive, color.RGBA{0x00, 0xff, 0x00, 0x00})

	button := g.Buttonf("0x%02x ", slice[localAddr]).OnClick(d.OnClickMemView(globalAddr))
	return style.To(button)

}

func (d *Debugger) isCurrentPC(addr uint16) bool {
	return d.e.Cpu.PC == addr
}

func (d *Debugger) isInDebug(addr uint16) bool {
	return slices.Contains(d.breakpoints, addr)
}

func (d *Debugger) OnClickMemView(addr uint16) func() {
	return func() {

		d.ToggleBP(addr)

	}
}

func (d *Debugger) makeRegColumns() []*g.TableColumnWidget {
	regColumns := make([]*g.TableColumnWidget, 12)
	regColumns[0] = g.TableColumn(fmt.Sprintf("PC: 0x%04x", d.e.Cpu.PC))
	regColumns[1] = g.TableColumn(fmt.Sprintf("SP: 0x%04x", d.e.Cpu.SP))
	regColumns[2] = g.TableColumn(fmt.Sprintf("A: 0x%02x", d.e.Cpu.A))
	regColumns[3] = g.TableColumn(fmt.Sprintf("F: 0x%02x", d.e.Cpu.F))
	regColumns[4] = g.TableColumn(fmt.Sprintf("B: 0x%02x", d.e.Cpu.B))
	regColumns[5] = g.TableColumn(fmt.Sprintf("C: 0x%02x", d.e.Cpu.C))
	regColumns[6] = g.TableColumn(fmt.Sprintf("D: 0x%02x", d.e.Cpu.D))
	regColumns[7] = g.TableColumn(fmt.Sprintf("E: 0x%02x", d.e.Cpu.E))
	regColumns[8] = g.TableColumn(fmt.Sprintf("H: 0x%02x", d.e.Cpu.H))
	regColumns[9] = g.TableColumn(fmt.Sprintf("L: 0x%02x", d.e.Cpu.L))
	div := d.e.Cpu.Memory.Io.GetDIV()

	regColumns[10] = g.TableColumn(fmt.Sprintf("DIV: 0x%02x", div))

	tima := d.e.Cpu.Memory.Io.GetTIMA()
	regColumns[11] = g.TableColumn(fmt.Sprintf("TIMA: 0x%02x", tima))

	return regColumns
}

func (d *Debugger) SetEmu(emu *Emulator) {
	d.reset()
	d.e = emu
}

func (d *Debugger) RunEmulator() {
	for {
		d.e.SerialOut()

		// if d.e.Cpu.Stop {
		// 	continue
		// }
		if d.autorun {
			if slices.Contains(d.GetBreakpoints(), d.e.Cpu.PC) && d.e.Cpu.PC != uint16(d.lastBPHit) {
				d.autorun = false
				d.lastBPHit = int(d.e.Cpu.PC)
			} else {
				d.lastBPHit = -1

				d.e.Step()
			}
		} else {
			if d.doStep {
				d.lastBPHit = -1
				d.e.Step()
				d.doStep = false
			}
		}
	}

}
