package ppu

import (
	"encoding/hex"
	_ "fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"go-boy/cpu"
	"go-boy/mmap"
	"image"
	"image/color"
	"unsafe"
)

type Ppu struct {
	screenMultiplier int
	running          bool
	screen           rl.Image
	window           unsafe.Pointer
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF
var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

const (
	color1 = "75a973"
	color2 = "72a381"
	color3 = "749989"
	color4 = "77978a"
)

type Tile struct {
	lines [8]uint16
}

func NewPpu(screenMultiplier int) *Ppu {
	ppu := Ppu{}
	ppu.Restart(screenMultiplier)

	return &ppu

}

func (p *Ppu) Restart(screenMultiplier int) {

	if p.window == nil {
		//setup
		p.screenMultiplier = screenMultiplier
		rl.InitWindow(int32(GB_WINDOW_WIDTH*p.screenMultiplier), int32(GB_WINDOW_HEIGHT*p.screenMultiplier), "raygui - button")
		rl.ClearColor(0, 0, 0, 255)

		p.window = rl.GetWindowHandle()
		//p.screen = rl.NewImage()
	} else {
		//only clear
		rl.ClearColor(0, 0, 0, 255)
		//p.screen = rl.NewImage()

	}
}

func (p *Ppu) Step(cpu *cpu.Cpu) {

	//clear screen and Image
	rl.ClearColor(0, 0, 0, 255)
	//p.screen = rl.NewImage()

	//Draw on Imnage
	for x := range 18 {
		for y := range 20 {
			tile := p.readTile(uint16(x*18+y), cpu, true)
			p.renderTile(tile, x*8, y*8)
		}
	}

	// Draw Texture(image) on window, scaled up to right sizeMult
	texture := rl.LoadTextureFromImage(&p.screen)
	rl.DrawTextureEx(texture, rl.NewVector2(0, 0), 0, float32(p.screenMultiplier), rl.White)
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

	//draw to p.screen
	//pic := pixel.PictureDataFromImage(img)
	//tileSprite := pixel.NewSprite(pic, pic.Bounds())
	//matrix := pixel.IM.Moved(pixel.V(float64(positionX+4), float64(positionY+4)))

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
		} else { // LCD.4 = 0
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
