package draw

import (
	"encoding/hex"
	"go-boy/cpu"
	"go-boy/mmap"
	"image/color"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	color1 = "75a973"
	color2 = "72a381"
	color3 = "749989"
	color4 = "77978a"
)

type Tile struct {
	Lines [8]uint16
}

var tileTexture rl.Texture2D

func CreateWindow(width, height int, name string) unsafe.Pointer {
	rl.InitWindow(int32(width), int32(height), name)
	return rl.GetWindowHandle()
}

func RenderTileToScreen(tile Tile, positionX int, positionY int, screen rl.RenderTexture2D) {

	if tileTexture.ID == 0 {
		tileTexture = rl.LoadTextureFromImage(rl.GenImageColor(8, 8, rl.White))
	}
	//img := rl.GenImageColor(8, 8, rl.White)
	img := []color.RGBA{}

	//Assign
	for y := range 8 {
		currentLine := tile.Lines[y]
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
			img = append(img, c)
		}
	}

	rl.UpdateTexture(tileTexture, img)

	//draw tex onto render Texture
	rl.BeginTextureMode(screen)
	rl.DrawTexture(tileTexture, int32(positionX), int32(positionY), rl.White)
	rl.EndTextureMode()

}
func ReadTile(tileDataOffset uint16, cpu *cpu.Cpu, isObject bool) Tile {

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
		tile.Lines[i], _ = cpu.Memory.Read16At(tileStart + uint16(i))
	}
	return tile
}

func getColor(bits int) color.RGBA {
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
		return color.RGBA{0, 0, 0, 0xFF}
	}

}
