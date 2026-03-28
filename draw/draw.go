package draw

import (
	"go-boy/cpu"
	"go-boy/mmap"
	"image/color"
)

var color1 = color.RGBA{0x75, 0xa9, 0x73, 0xFF}
var color2 = color.RGBA{0x72, 0xa3, 0x81, 0xFF}
var color3 = color.RGBA{0x74, 0x99, 0x89, 0xFF}
var color4 = color.RGBA{0x77, 0x97, 0x8a, 0xFF}

type Tile struct {
	Lines [8]uint16
}



// func RenderTileToScreen(tile Tile, positionX int, positionY int, screen rl.RenderTexture2D) {

// 	if tileTexture.ID == 0 {
// 		tileTexture = rl.LoadTextureFromImage(rl.GenImageColor(8, 8, rl.White))
// 	}
// 	//img := rl.GenImageColor(8, 8, rl.White)
// 	img := make([]color.RGBA, 64)

// 	//Assign
// 	for y := range 8 {
// 		currentLine := tile.Lines[y]
// 		for x := range 8 {
// 			colorLsb := 0
// 			if mmap.GetBit16(currentLine, uint8(1*x)) {
// 				colorLsb = 1
// 			}
// 			colorMsb := 0
// 			if mmap.GetBit16(currentLine, uint8(8+1*x)) {
// 				colorMsb = 1
// 			}
// 			colorBits := colorLsb & (colorMsb << 1)
// 			c := getColor(colorBits)
// 			img[y*8+x] = c
// 		}
// 	}

// 	rl.UpdateTexture(tileTexture, img)

// 	//draw tex onto render Texture
// 	rl.BeginTextureMode(screen)
// 	rl.DrawTexture(tileTexture, int32(positionX), int32(positionY), rl.White)
// 	rl.EndTextureMode()

// }

func RenderObjectsToScreen(objects []Tile, screen interface{}) {
	// Stubbed
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
		return color1
	case 1:
		return color2
	case 2:
		return color3
	case 3:
		return color4
	default:
		return color.RGBA{0, 0, 0, 0xFF}
	}

}
