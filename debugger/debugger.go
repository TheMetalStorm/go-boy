package debugger

import (
	"fmt"
	"go-boy/emulator"
	"slices"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

type Emulator = emulator.Emulator

type Debugger struct {
	autorun     bool
	doStep      bool
	breakpoints []uint16
	lastBPHit   int

	e *Emulator
}

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

func (d *Debugger) SetEmu(emu *Emulator) {
	d.reset()
	d.e = emu
}

var lastFrameTime time.Time

// RunEmulator spins freely to execute CPU instructions at normal speed
func (d *Debugger) RunEmulator() {
	for {
		d.e.SerialOut()
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
			} else {
				// Prevent maxing a core just waiting for an un-paused state
				time.Sleep(time.Millisecond * 5)
			}
		}
	}
}

// Render should be called in the cimgui-go main loop (every frame)
func (d *Debugger) Render() {

	// Limit framerate to ~60FPS to prevent the unlocked SDL backend from spinning at >1000fps and eating the CPU.
	elapsed := time.Since(lastFrameTime)
	if elapsed < time.Millisecond*16 {
		time.Sleep(time.Millisecond*16 - elapsed)
	}
	lastFrameTime = time.Now()

	// Main UI
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPosV(viewport.Pos(), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(viewport.Size(), imgui.CondFirstUseEver)

	imgui.Begin("GB Debugger")
	// Control Row
	imgui.Text("Control: ")
	imgui.SameLine()
	if imgui.Button("Start") {
		d.autorun = true
	}
	imgui.SameLine()
	if imgui.Button("Stop") {
		d.autorun = false
	}
	imgui.SameLine()
	if imgui.Button("Step") {
		d.doStep = true
	}
	imgui.SameLine()
	if imgui.Button("Restart Machine") {
		d.autorun = false
		d.e.Restart()
	}

	// Registers Table
	imgui.Text("Regs: ")
	if imgui.BeginTable("Regs", 13) {
		imgui.TableSetupColumn(fmt.Sprintf("PC: 0x%04x", d.e.Cpu.PC))
		imgui.TableSetupColumn(fmt.Sprintf("SP: 0x%04x", d.e.Cpu.SP))
		imgui.TableSetupColumn(fmt.Sprintf("A: 0x%02x", d.e.Cpu.A))
		imgui.TableSetupColumn(fmt.Sprintf("F: 0x%02x", d.e.Cpu.F))
		imgui.TableSetupColumn(fmt.Sprintf("B: 0x%02x", d.e.Cpu.B))
		imgui.TableSetupColumn(fmt.Sprintf("C: 0x%02x", d.e.Cpu.C))
		imgui.TableSetupColumn(fmt.Sprintf("D: 0x%02x", d.e.Cpu.D))
		imgui.TableSetupColumn(fmt.Sprintf("E: 0x%02x", d.e.Cpu.E))
		imgui.TableSetupColumn(fmt.Sprintf("H: 0x%02x", d.e.Cpu.H))
		imgui.TableSetupColumn(fmt.Sprintf("L: 0x%02x", d.e.Cpu.L))
		imgui.TableSetupColumn(fmt.Sprintf("DIV: 0x%02x", d.e.Cpu.Memory.Io.GetDIV()))
		imgui.TableSetupColumn(fmt.Sprintf("TIMA: 0x%02x", d.e.Cpu.Memory.Io.GetTIMA()))
		imgui.TableSetupColumn(fmt.Sprintf("HALT: %t", d.e.Cpu.Halt))
		imgui.TableHeadersRow()
		imgui.EndTable()
	}

	// Side-by-side Layout: Stack and Tabs
	imgui.BeginChildStrV("StackRegion", imgui.NewVec2(200, 0), 0, 0)
	imgui.Text("Stack: (HRAM)")
	if imgui.BeginTable("Hram Stack", 2) {
		hram := d.e.Cpu.Memory.Hram[:]
		start := 0xff80
		rowCount := int32(len(hram))
		clipper := imgui.NewListClipper()
		defer clipper.Destroy()
		clipper.Begin(rowCount)
		for clipper.Step() {
			for r := clipper.DisplayStart(); r < clipper.DisplayEnd(); r++ {
				i := len(hram) - 1 - int(r)
				imgui.TableNextRow()
				imgui.TableNextColumn()
				imgui.Text(fmt.Sprintf("0x%04x:", start+i))
				imgui.TableNextColumn()
				imgui.Text(fmt.Sprintf("0x%02x", hram[i]))
			}
		}
		imgui.EndTable()
	}
	imgui.EndChild()

	imgui.SameLine()

	imgui.BeginChildStrV("TabsRegion", imgui.NewVec2(0, 0), 0, 0)
	// Tabs
	if imgui.BeginTabBar("Memory Regions") {
		if imgui.BeginTabItem("Bank0") {
			d.RenderMemoryTable("Bank0", d.e.Cpu.Memory.GetBank0(), 0, true)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("BankN") {
			d.RenderMemoryTable("BankN", d.e.Cpu.Memory.GetBank1(), 0x4000, true)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Vram") {
			d.RenderMemoryTable("Vram", d.e.Cpu.Memory.GetVram(), 0x8000, false)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Extram") {
			d.RenderMemoryTable("Extram", d.e.Cpu.Memory.GetExtram(), 0xa000, false)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Wram1") {
			d.RenderMemoryTable("Wram1", d.e.Cpu.Memory.GetWram1(), 0xc000, true)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Wram2") {
			d.RenderMemoryTable("Wram2", d.e.Cpu.Memory.GetWram2(), 0xd000, true)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Oam") {
			d.RenderMemoryTable("Oam", d.e.Cpu.Memory.GetOam(), 0xfe00, false)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Io") {
			d.RenderMemoryTable("Io", d.e.Cpu.Memory.Io.Regs[:], 0xff00, false)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Hram") {
			d.RenderMemoryTable("Hram", d.e.Cpu.Memory.GetHram(), 0xff80, false)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Ie") {
			imgui.Text(fmt.Sprintf("0xFFFF: %02x", d.e.Cpu.Memory.Ie))
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Game Code") {
			d.RenderMemoryTable("Game Code", d.e.GetCurrentGame(), 0, true)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}
	imgui.EndChild()

	imgui.End()
}

func (d *Debugger) RenderMemoryTable(id string, slice []uint8, offset uint32, debuggable bool) {
	if imgui.BeginTable(id, 17) {
		rowCount := int32((len(slice) + 15) / 16)
		clipper := imgui.NewListClipper()
		defer clipper.Destroy()

		clipper.Begin(rowCount)
		for clipper.Step() {
			for r := clipper.DisplayStart(); r < clipper.DisplayEnd(); r++ {
				i := int(r) * 16
				imgui.TableNextRow()
				imgui.TableNextColumn()
				imgui.Text(fmt.Sprintf("0x%04x:", int(offset)+i))

				for j := 0; j < 16; j++ {
					imgui.TableNextColumn()
					localAddr := uint16(i + j)
					if int(localAddr) < len(slice) {
						globalAddr := uint16(offset) + localAddr
						if debuggable {
							pushed := 0
							if d.isCurrentPC(globalAddr) {
								imgui.PushStyleColorVec4(imgui.ColText, imgui.NewVec4(0, 1, 0, 1))
								pushed++
							} else if d.isInDebug(globalAddr) {
								imgui.PushStyleColorVec4(imgui.ColText, imgui.NewVec4(0, 0.7, 0.7, 1))
								pushed++
							}

							// Disable button framing to look like giu's styling
							if imgui.SelectableBool(fmt.Sprintf("%02x##%04x", slice[localAddr], globalAddr)) {
								d.ToggleBP(globalAddr)
							}

							for pushed > 0 {
								imgui.PopStyleColor()
								pushed--
							}
						} else {
							if imgui.SelectableBool(fmt.Sprintf("%02x##%04x", slice[localAddr], globalAddr)) {
								// Just visual
							}
						}
					}
				}
			}
		}
		imgui.EndTable()
	}
}

func (d *Debugger) isCurrentPC(addr uint16) bool {
	return d.e.Cpu.PC == addr
}

func (d *Debugger) isInDebug(addr uint16) bool {
	return slices.Contains(d.breakpoints, addr)
}
