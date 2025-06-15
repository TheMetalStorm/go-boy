package ppu

import (
	"encoding/hex"
	"fmt"
	_ "fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"go-boy/cpu"
	"go-boy/mmap"
	"image"
	"image/color"
)

type Ppu struct {
	screenMultiplier float64
	running          bool
	window           *pixelgl.Window // window screen
	canvas           *pixelgl.Canvas //gb screen
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF

const (
	color1 = "75a973"
	color2 = "72a381"
	color3 = "749989"
	color4 = "77978a"
)

type Tile struct {
	lines [8]uint16
}

func NewPpu(screenMultiplier float64) *Ppu {
	ppu := Ppu{}
	ppu.Restart(screenMultiplier)

	return &ppu

}

func (p *Ppu) Restart(screenMultiplier float64) {
	if p.window == nil {

		cfg := pixelgl.WindowConfig{
			Title:     "GoBoy!",
			Bounds:    pixel.R(0, 0, 160*screenMultiplier, 144*screenMultiplier),
			Resizable: true,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			fmt.Printf("Could not create window: %v", err)
		}
		p.screenMultiplier = screenMultiplier
		p.window = win
		p.canvas = pixelgl.NewCanvas(pixel.R(0, 0, 160, 144))

	}

}

func (p *Ppu) Step(cpu *cpu.Cpu) {
	p.canvas.Clear(pixel.RGB(0, 0, 0))
	p.window.Clear(pixel.RGB(0, 0, 0))

	for x := range 18 {
		for y := range 20 {
			tile := p.readTile(uint16(x*18+y), cpu, true)
			p.renderTile(tile, x*8, y*8)
		}
	}

	matrix := pixel.IM.
		Scaled(pixel.ZV, p.screenMultiplier). // Skalierung
		Moved(p.window.Bounds().Center())

	p.canvas.Draw(p.window, matrix)
	p.window.Update()
}

func (p *Ppu) renderTile(tile Tile, positionX int, positionY int) {

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))

	//Assign
	for y := range 8 {
		currentLine := tile.lines[y]
		for x := range 8 {
			colorLsb := 0
			if mmap.GetBit16(currentLine, uint8(1*x)) {
				colorLsb = 1
			}
			colorMsb := 0
			if mmap.GetBit16(currentLine, uint8(8+1*x)) {
				colorMsb = 1
			}
			colorBits := colorLsb & (colorMsb << 1)
			c := getColor(colorBits)
			img.Set(x, y, c)
		}
	}

	pic := pixel.PictureDataFromImage(img)
	tileSprite := pixel.NewSprite(pic, pic.Bounds())
	matrix := pixel.IM.Moved(pixel.V(float64(positionX+4), float64(positionY+4)))
	tileSprite.Draw(p.canvas, matrix)
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

func getColor(bits int) color.Color {
	//TODO: Get According to Palette
	switch bits {
	case 0:
		b, _ := hex.DecodeString(color1)
		return color.RGBA{b[0], b[1], b[2], 0xFF}
	case 1:
		b, _ := hex.DecodeString(color2)
		return color.RGBA{b[0], b[1], b[2], 0xFF}
	case 2:
		b, _ := hex.DecodeString(color3)
		return color.RGBA{b[0], b[1], b[2], 0xFF}
	case 3:
		b, _ := hex.DecodeString(color4)
		return color.RGBA{b[0], b[1], b[2], 0xFF}
	default:
		return color.Black
	}

}
