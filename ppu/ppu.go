package ppu

import (
	"fmt"
	_ "fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"go-boy/cpu"
	"go-boy/mmap"
)

type Ppu struct {
	running bool
	window  *pixelgl.Window
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF

const (
	color1 = 0x75a973
	color2 = 0x72a381
	color3 = 0x749989
	color4 = 0x77978a
)

type Tile struct {
	lines [8]uint16
}

func NewPpu() *Ppu {
	ppu := Ppu{}
	ppu.Restart()

	return &ppu

}

func (p *Ppu) Restart() {
	if p.window == nil {

		cfg := pixelgl.WindowConfig{
			Title:     "GoBoy!",
			Bounds:    pixel.R(0, 0, 600, 800),
			Invisible: false,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			fmt.Printf("Could not create window: %v", err)
		}
		p.window = win
	}

}

func (p *Ppu) Step(cpu *cpu.Cpu) {
	p.window.Clear(pixel.RGB(0, 0, 0))
	tile := p.readTile(0x8000, cpu, true)
	p.renderTile(tile)
	p.window.Update()
}

func (p *Ppu) readTile(tileDataOffset uint16, cpu *cpu.Cpu, isObject bool) Tile {

	var tileStart uint16

	if isObject {
		// Object
		tileStart = 0x8000 + tileDataOffset
	} else {
		addressingMode := mmap.GetBit(cpu.Memory.Io.GetLCDC(), 4)
		if addressingMode { // LCD.4 = 1
			// same as Object
			tileStart = 0x8000 + tileDataOffset
		} else {                      // LCD.4 = 0
			if tileDataOffset > 127 { //128-255 start at 0x8800
				tileStart = 0x8800 + tileDataOffset
			} else { //0-127 start at 0x9000
				tileStart = 0x9000 + tileDataOffset
			}
		}
	}
	var tile Tile
	for i := 0; i < 8; i++ {
		tile.lines[i], _ = cpu.Memory.Read16At(tileStart + uint16(i))
	}
	return tile
}

func (p *Ppu) renderTile(tile Tile) {

}
