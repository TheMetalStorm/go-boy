package ppu

import (
	_ "fmt"
	"go-boy/cpu"
	"unsafe"
)

type Ppu struct {
	screenMultiplier int
	running          bool
	screen           interface{}
	window           unsafe.Pointer
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF
var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

func NewPpu(screenMultiplier int) *Ppu {
	ppu := Ppu{}
	ppu.Restart(screenMultiplier)

	return &ppu

}

func (p *Ppu) Restart(screenMultiplier int) {
	if p.window == nil {
		p.screenMultiplier = screenMultiplier
	}
	p.running = true
}

func (p *Ppu) Step(cpu *cpu.Cpu) {

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
