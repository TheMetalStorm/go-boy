package ppu

import (
	_ "fmt"
	"go-boy/cpu"
	"go-boy/ioregs"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Ppu struct {
	screenMultiplier int
	running          bool
	Surface          *sdl.Surface
	Window           *sdl.Window

	currentDot uint64
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
	p.currentDot = 0
}

func (p *Ppu) Step(cpu *cpu.Cpu, ranMCyclesThisStep uint64) {

	ranDotsThisCPUStep := ranMCyclesThisStep * 4 //Double Speed Mode : * 2
	p.currentDot += ranDotsThisCPUStep

	if p.currentDot >= 456 {
		p.currentDot = 0
		cpu.Memory.Io.SetLY(cpu.Memory.Io.GetLY() + 1)
		if cpu.Memory.Io.GetLY() == 144 {
			cpu.Memory.Io.SetInterruptFlagBit(ioregs.VBLANK, true)
		}
		if cpu.Memory.Io.GetLY() > 153 {
			cpu.Memory.Io.SetLY(0)
		}
		cpu.Memory.Io.SetSTATBit(ioregs.STAT_LY_EQ_LYC, cpu.Memory.Io.GetLY() == cpu.Memory.Io.GetLYC())
		if cpu.Memory.Io.GetSTATBit(ioregs.STAT_LY_EQ_LYC) && cpu.Memory.Io.GetSTATBit(ioregs.STAT_LYC_INT) {
			cpu.Memory.Io.SetInterruptFlagBit(ioregs.LCD, true)
		}

	}

}

func (p *Ppu) Render(cpu *cpu.Cpu) {

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
