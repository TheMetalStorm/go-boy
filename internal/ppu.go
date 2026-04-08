package internal

import (
	_ "fmt"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Ppu struct {
	screenMultiplier int
	running          bool
	Surface          *sdl.Surface
	Window           *sdl.Window

	CurrentMode PpuMode
	CurrentDot  uint64

	Cpu *Cpu
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF
var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

func NewPpu(screenMultiplier int) *Ppu {
	ppu := Ppu{}
	window, err := sdl.CreateWindow("go-boy!", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ppu.Window = window

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	ppu.Surface = surface
	ppu.Restart(screenMultiplier)

	return &ppu

}

func (p *Ppu) Restart(screenMultiplier int) {
	if p.Window == nil {
		p.screenMultiplier = screenMultiplier
	}

	p.running = true
	p.CurrentDot = 0
	p.CurrentMode = MODE_2
}

func (p *Ppu) Step(ranMCyclesThisStep uint64) {

	//	TODO: figure out REAL mode 3 duration with penalties
	mode3Duration := uint64(172)
	mode0Duration := uint64(376 - mode3Duration)
	mode2Duration := uint64(80)

	ranDotsThisCPUStep := ranMCyclesThisStep * 4 //Double Speed Mode : * 2
	for i := uint64(0); i < ranDotsThisCPUStep; i++ {
		p.CurrentDot++
		if p.CurrentDot < mode2Duration && p.CurrentMode != MODE_2 {
			p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_LSB, false)
			p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_MSB, true)
			p.CurrentMode = MODE_2
			if p.Cpu.Memory.Io.GetSTATBit(STAT_MODE_2_INT) {
				p.Cpu.Memory.Io.SetInterruptFlagBit(LCD, true)
			}
		} else if p.CurrentDot >= mode2Duration && p.CurrentDot < mode2Duration+mode3Duration && p.CurrentMode != MODE_3 {
			//THIS IS WHERE WE DRAW THE Pixel for
			p.CurrentMode = MODE_3
			p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_LSB, true)
			p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_MSB, true)
		} else if p.CurrentDot >= mode2Duration+mode3Duration && p.CurrentDot < mode2Duration+mode3Duration+mode0Duration && p.CurrentDot < 456 && p.CurrentMode != MODE_0 {
			p.CurrentMode = MODE_0
			p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_LSB, false)
			p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_MSB, false)
			if p.Cpu.Memory.Io.GetSTATBit(STAT_MODE_0_INT) {
				p.Cpu.Memory.Io.SetInterruptFlagBit(LCD, true)
			}
		}
		if p.CurrentDot >= 456 {
			p.CurrentDot = 0
			p.Cpu.Memory.Io.SetLY(p.Cpu.Memory.Io.GetLY() + 1)

			//Enter VBlank and trigger VBlank interrupt
			if p.Cpu.Memory.Io.GetLY() == 144 {
				p.Cpu.Memory.Io.SetInterruptFlagBit(VBLANK, true)
			}

			if p.Cpu.Memory.Io.GetLY() > 143 && p.Cpu.Memory.Io.GetLY() < 154 && p.CurrentMode != MODE_1 {
				p.CurrentMode = MODE_1
				p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_LSB, true)
				p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_MSB, false)
				if p.Cpu.Memory.Io.GetSTATBit(STAT_MODE_1_INT) {
					p.Cpu.Memory.Io.SetInterruptFlagBit(LCD, true)
				}
			}

			if p.Cpu.Memory.Io.GetLY() > 153 {
				p.Cpu.Memory.Io.SetLY(0)
				p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_LSB, false)
				p.Cpu.Memory.Io.SetSTATBit(STAT_PPU_MODE_MSB, true)
				p.CurrentMode = MODE_2
			}

			p.Cpu.Memory.Io.SetSTATBit(STAT_LY_EQ_LYC, p.Cpu.Memory.Io.GetLY() == p.Cpu.Memory.Io.GetLYC())
			if p.Cpu.Memory.Io.GetSTATBit(STAT_LY_EQ_LYC) && p.Cpu.Memory.Io.GetSTATBit(STAT_LYC_INT) {
				p.Cpu.Memory.Io.SetInterruptFlagBit(LCD, true)
			}

		}
	}

}

func (p *Ppu) Render() {

	p.Surface.FillRect(nil, 0)
	rect := sdl.Rect{0, 0, int32(time.Now().Second()), 200}
	colour := sdl.Color{R: 255, G: 0, B: 255, A: 255} // purple
	format, _ := sdl.AllocFormat(sdl.PIXELFORMAT_RGBA8888)
	pixel := sdl.MapRGBA(format, colour.R, colour.G, colour.B, colour.A)
	p.Surface.FillRect(&rect, pixel)
	p.Window.UpdateSurface()
	// for p.running {

	// }
	//TODO
	//clear screen and Image
	// rl.ClearColor(0, 0, 0, 255)

	// //clear render texture
	// rl.BeginTextureMode(p.screen)
	// rl.ClearBackground(rl.Blank)
	// rl.EndTextureMode()

	// //Draw on RenderTexture
	// for x := range 18 {
	// 	for y := range 20 {
	// 		// tile := draw.ReadTile(uint16(x*18+y), cpu, true)
	// 		// draw.RenderTileToScreen(tile, x*8, y*8, p.screen)
	// 	}
	// }

	// Draw RenderTexture on window, scaled up to right sizeMult
	// rl.DrawTextureEx(p.screen.Texture, rl.NewVector2(0, 0), 0, float32(p.screenMultiplier), rl.White)
}
