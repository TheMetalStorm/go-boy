package debugger

import (
	"fmt"
	"go-boy/emulator"
	"go-boy/internal"
	"slices"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/micahke/imgui-go"
)

type Emulator = emulator.Emulator

type Debugger struct {
	autorun     bool
	doStep      bool
	doFastStep  bool
	breakpoints []uint16
	lastBPHit   int

	window *glfw.Window

	oldRenderMode internal.PpuMode

	e *Emulator
}

func NewDebugger() *Debugger {
	dbg := &Debugger{}
	dbg.reset()

	return dbg
}

func (d *Debugger) reset() {
	d.oldRenderMode = internal.MODE_2
	d.autorun = true
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
	d.e.Ppu.HandleGLUpdate = false
}

var lastFrameTime time.Time

var lastLY uint8 = 0

func (d *Debugger) RunEmulator() {

	for !d.e.Window.ShouldClose() {

		d.e.Window.MakeContextCurrent()
		glfw.PollEvents()

		if d.e.Io.WantTextInput() {
			d.e.Impl.SetDefaultKeyCallback()
		} else {
			d.e.Window.SetKeyCallback(d.debugKeyCallback)
		}

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
				d.e.DoRender = true
				d.doStep = false
			}
			if d.doFastStep {
				d.lastBPHit = -1
				d.e.Step()
				d.e.DoRender = true
			}
		}

		if d.e.DoRender {
			gl.Clear(gl.COLOR_BUFFER_BIT)

			d.e.DoRender = false

			d.e.Ppu.Render()

			gl.UseProgram(d.e.Program)
			gl.BindVertexArray(d.e.Vao)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)

			d.e.Impl.NewFrame()
			d.Render()
			imgui.Render()
			d.e.Impl.Render(imgui.RenderedDrawData())

			d.e.Window.SwapBuffers()

		}

	}

}

func (d *Debugger) RenderTileViewer() {
	d.e.Ppu.RenderTileViewer()
}

func (d *Debugger) RenderBackgroundMapViewer() {
	d.e.Ppu.RenderBackgroundMapViewer()
}

func (d *Debugger) RenderWindowMapViewer() {
	d.e.Ppu.RenderWindowMapViewer()
}

func (d *Debugger) Render() {

	imgui.Begin("VRAM")
	if imgui.BeginTabBar("View ") {
		if imgui.BeginTabItem("TileViewer") {
			d.RenderTileViewer()
			imgui.Image(imgui.TextureID(d.e.Ppu.TileViewerTex), imgui.Vec2{X: 16 * 8 * 4, Y: 24 * 8 * 4})
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Background Map") {
			d.RenderBackgroundMapViewer()
			imgui.Image(imgui.TextureID(d.e.Ppu.BackgroundTex), imgui.Vec2{X: 32 * 8 * 3, Y: 32 * 8 * 3})
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Window Map") {
			d.RenderWindowMapViewer()
			imgui.Image(imgui.TextureID(d.e.Ppu.WindowTex), imgui.Vec2{X: 32 * 8 * 3, Y: 32 * 8 * 3})
			imgui.EndTabItem()
		}
	}
	imgui.EndTabBar()
	imgui.End()

	imgui.Begin("GB Debugger")

	imgui.Text("Control: ")
	imgui.SameLine()
	imgui.Text("F3=Start  F4=Stop  F5=Step  F6=FastStep  F7=Restart")

	imgui.Text("Regs: ")
	if imgui.BeginTable("Regs", 13, imgui.TableFlags_None, imgui.Vec2{X: 0, Y: 0}, 0.0) {
		imgui.TableSetupColumn(fmt.Sprintf("PC: 0x%04x", d.e.Cpu.PC), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("SP: 0x%04x", d.e.Cpu.SP), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("A: 0x%02x", d.e.Cpu.A), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("F: 0x%02x", d.e.Cpu.F), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("B: 0x%02x", d.e.Cpu.B), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("C: 0x%02x", d.e.Cpu.C), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("D: 0x%02x", d.e.Cpu.D), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("E: 0x%02x", d.e.Cpu.E), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("H: 0x%02x", d.e.Cpu.H), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("L: 0x%02x", d.e.Cpu.L), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("DIV: 0x%02x", d.e.Cpu.Memory.Io.GetDIV()), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("TIMA: 0x%02x", d.e.Cpu.Memory.Io.GetTIMA()), 0, 0, 0)
		imgui.TableSetupColumn(fmt.Sprintf("HALT: %t", d.e.Cpu.Halt), 0, 0, 0)
		imgui.TableHeadersRow()
		imgui.EndTable()
	}

	imgui.BeginChild("TabsRegion")
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

func (d *Debugger) debugKeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyF3:
		if action == glfw.Press {
			d.autorun = true
		}
	case glfw.KeyF4:
		if action == glfw.Press {
			d.autorun = false
		}
	case glfw.KeyF5:
		if action == glfw.Press {
			d.doStep = true
		}
	case glfw.KeyF6:
		if action == glfw.Press {
			d.doFastStep = true
		} else if action == glfw.Release {
			d.doFastStep = false
		}
	case glfw.KeyF7:
		if action == glfw.Press {
			d.autorun = false
			d.e.Restart()
			d.e.DoRender = true
		}
	}

	KeyCallback(w, key, scancode, action, mods)
}

func (d *Debugger) RenderMemoryTable(id string, slice []uint8, offset uint32, debuggable bool) {
	if imgui.BeginTable(id, 17, imgui.TableFlags_None, imgui.Vec2{X: 0, Y: 0}, 0.0) {
		for i := 0; i < len(slice); i += 16 {
			imgui.TableNextRow(0, 0.0)
			imgui.TableNextColumn()
			imgui.Text(fmt.Sprintf("0x%04x:", int(offset)+i))

			for j := 0; j < 16 && i+j < len(slice); j++ {
				imgui.TableNextColumn()
				globalAddr := uint16(offset) + uint16(i+j)
				if debuggable {
					pushed := 0
					if d.isCurrentPC(globalAddr) {
						imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0, Y: 1, Z: 0, W: 1})
						pushed++
					} else if d.isInDebug(globalAddr) {
						imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0, Y: 0.7, Z: 0.7, W: 1})
						pushed++
					}

					if imgui.Selectable(fmt.Sprintf("%02x##%04x", slice[i+j], globalAddr)) {
						d.ToggleBP(globalAddr)
					}

					for pushed > 0 {
						imgui.PopStyleColor()
						pushed--
					}
				} else {
					imgui.Text(fmt.Sprintf("%02x", slice[i+j]))
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

func KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

}
