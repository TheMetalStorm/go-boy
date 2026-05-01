package internal

import (
	"image/color"
	"math/bits"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Ppu struct {
	screenMultiplier int
	running          bool

	ViewPortTex    uint32
	ViewPortObjTex uint32
	BackgroundTex  uint32
	WindowTex      uint32
	TileViewerTex  uint32
	ObjOverviewTex uint32

	viewportBuf    []color.RGBA
	viewportObjBuf []color.RGBA
	backgroundBuf  []color.RGBA
	windowBuf      []color.RGBA
	tileViewerBuf  []color.RGBA
	objOverviewBuf []color.RGBA

	CurrentMode PpuMode
	CurrentDot  uint64

	Cpu *Cpu

	HandleGLUpdate bool
	Frame          uint64
}

type Tile struct {
	Lines [8][2]uint8
}

// Flip in hardware later? - Texture
func (t *Tile) FlipY() {
	for i := 0; i < len(t.Lines)/2; i++ {
		j := len(t.Lines) - 1 - i
		t.Lines[i], t.Lines[j] = t.Lines[j], t.Lines[i]
	}
}

func (t *Tile) FlipX() {
	for i := 0; i < len(t.Lines); i++ {
		t.Lines[i][0] = bits.Reverse8(t.Lines[i][0])
		t.Lines[i][1] = bits.Reverse8(t.Lines[i][1])
	}
}

type Object struct {
	yPos       uint8
	xPos       uint8
	tileInd    uint8
	attributes uint8
}

var TILE_DATA_START = 0x8000
var TILE_DATA_END int = 0x97FF
var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

const TILE_X_Y int = 8
const BG_WINDOW_X_Y int = 256

var quadVAO uint32
var quadVBO uint32

func NewPpu(screenMultiplier int) *Ppu {
	ppu := &Ppu{}
	ppu.screenMultiplier = screenMultiplier
	ppu.Restart(screenMultiplier)

	ppu.viewportBuf = make([]color.RGBA, GB_WINDOW_WIDTH*GB_WINDOW_HEIGHT)
	ppu.viewportObjBuf = make([]color.RGBA, GB_WINDOW_WIDTH*GB_WINDOW_HEIGHT)
	ppu.backgroundBuf = make([]color.RGBA, BG_WINDOW_X_Y*BG_WINDOW_X_Y)

	ppu.windowBuf = make([]color.RGBA, BG_WINDOW_X_Y*BG_WINDOW_X_Y)

	ppu.objOverviewBuf = make([]color.RGBA, 8*TILE_X_Y*10*16)

	ppu.tileViewerBuf = make([]color.RGBA, 16*TILE_X_Y*24*TILE_X_Y)

	return ppu
}

func (p *Ppu) Restart(screenMultiplier int) {

	clear(p.viewportBuf)
	clear(p.viewportObjBuf)
	clear(p.backgroundBuf)
	clear(p.windowBuf)
	clear(p.objOverviewBuf)
	clear(p.tileViewerBuf)

	p.screenMultiplier = screenMultiplier
	p.running = true
	p.CurrentDot = 0
	p.CurrentMode = MODE_2

}

func (p *Ppu) Step(ranMCyclesThisStep uint64) {

	mode3Duration := uint64(172) + uint64(p.Cpu.Memory.Io.GetSCX())%8 //+ Num Sprites*8
	mode0Duration := uint64(376 - mode3Duration)
	mode2Duration := uint64(80)

	ranDotsThisCPUStep := ranMCyclesThisStep * 4
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

			// TODO: first we implement a simple version where we just render the whole screen at once
			//draw current line!
			// if p.Cpu.Memory.Io.GetLY() < 144 {
			// 	p.drawLine()
			// }

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

var tileColor1 = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
var tileColor2 = color.RGBA{0xAA, 0xAA, 0xAA, 0xFF}
var tileColor3 = color.RGBA{0x55, 0x55, 0x55, 0xFF}
var tileColor4 = color.RGBA{0x00, 0x00, 0x00, 0xFF}

// TODO: implement palette selection
// TODO: when obj, some bits are transparent
func getTileColor(bits int) color.RGBA {

	switch bits {
	case 0:
		return tileColor1
	case 1:
		return tileColor2
	case 2:
		return tileColor3
	case 3:
		return tileColor4
	default:
		return color.RGBA{255, 0, 0, 255}
	}
}

func (p *Ppu) FillWindowMapData() {
	clear(p.backgroundBuf)

	var areaStart uint16
	if !GetBit(p.Cpu.Memory.Io.GetLCDC(), 6) {
		areaStart = 0x9800
	} else {
		areaStart = 0x9C00
	}

	var indices [1024]uint8
	for i := 0; i < 1024; i++ {
		idx, _ := p.Cpu.Memory.ReadByteAtForced(areaStart + uint16(i))
		indices[i] = idx
	}

	bufW := 32 * 8

	for i, tileInd := range indices {
		tile := ReadTileForLayers(uint16(tileInd), p.Cpu)

		var lines = tile.Lines

		tileX := i % 32
		tileY := i / 32

		for py := 0; py < 8; py++ {
			currentLine := lines[py]
			for px := 0; px < 8; px++ {
				colorLsb := 0
				colorMsb := 0
				if GetBit(currentLine[0], uint8(px)) {
					colorLsb = 1
				}
				if GetBit(currentLine[1], uint8(px)) {
					colorMsb = 1
				}
				colorBits := colorLsb | (colorMsb << 1)
				bufIdx := (tileY*8+py)*bufW + (tileX*8 + px)
				p.windowBuf[bufIdx] = getTileColor(colorBits)
			}
		}
	}

}

func (p *Ppu) FillBackgroundMapData() {

	clear(p.backgroundBuf)
	var areaStart uint16
	if !GetBit(p.Cpu.Memory.Io.GetLCDC(), 3) {
		areaStart = 0x9800
	} else {
		areaStart = 0x9C00
	}

	var indices [1024]uint8
	for i := 0; i < 1024; i++ {
		idx, _ := p.Cpu.Memory.ReadByteAt(areaStart + uint16(i))
		indices[i] = idx
	}

	bufW := 32 * 8

	for i, tileInd := range indices {
		tile := ReadTileForLayers(uint16(tileInd), p.Cpu)

		var lines = tile.Lines

		tileX := i % 32
		tileY := i / 32

		for py := 0; py < 8; py++ {
			currentLine := lines[py]
			for px := 0; px < 8; px++ {
				colorLsb := 0
				colorMsb := 0
				if GetBit(currentLine[0], uint8(px)) {
					colorLsb = 1
				}
				if GetBit(currentLine[1], uint8(px)) {
					colorMsb = 1
				}
				colorBits := colorLsb | (colorMsb << 1)
				bufIdx := (tileY*8+py)*bufW + (tileX*8 + px)
				p.backgroundBuf[bufIdx] = getTileColor(colorBits)
			}
		}
	}

}

// TODO: instead of revers do this?
// This is a quirk of Game Boy LCD hardware - bit 0 is actually the LEFTMOST pixel, not rightmost.
// When iterating px = 0..7, you might be doing:
// colorLsb = (byte >> px) & 1  // px=0 gets bit 0 (LEFT pixel on screen)
// But since bit 0 = leftmost, you need:
// colorLsb = (byte >> (7-px)) & 1  // px=0 gets bit 7 (RIGHT pixel on screen)

func ReadTileDataBypass(absTileNumber uint16, c *Cpu) Tile {
	var tile Tile

	for i := 0; i < 8; i++ {
		leftPart, _ := c.Memory.ReadByteAtForced(uint16(TILE_DATA_START) + absTileNumber*16 + uint16(i*2))
		rightPart, _ := c.Memory.ReadByteAtForced(uint16(TILE_DATA_START) + absTileNumber*16 + uint16(i*2+1))

		tile.Lines[i] = [2]uint8{bits.Reverse8(leftPart), bits.Reverse8(rightPart)}

	}

	return tile
}

func ReadTileForLayers(relTileNumber uint16, c *Cpu) Tile {
	var tileStart uint16

	addressingMode := GetBit(c.Memory.Io.GetLCDC(), 4)
	if addressingMode { // LCD.4 = 1
		// same as Object
		tileStart = 0x8000 + relTileNumber*16
	} else { // LCD.4 = 0
		if relTileNumber > 127 { //128-255 start at 0x8800
			tileStart = 0x8800 + (relTileNumber-128)*16
		} else { //0-127 start at 0x9000
			tileStart = 0x9000 + relTileNumber*16
		}
	}

	var tile Tile
	for i := 0; i < 8; i++ {
		leftPart, _ := c.Memory.ReadByteAtForced(tileStart + uint16(i*2))
		rightPart, _ := c.Memory.ReadByteAtForced(tileStart + uint16(i*2+1))

		tile.Lines[i] = [2]uint8{bits.Reverse8(leftPart), bits.Reverse8(rightPart)}
	}
	return tile
}

func ReadTileForObjects(relTileNumber uint16, c *Cpu) Tile {
	var tileStart uint16 = 0x8000 + relTileNumber*16

	var tile Tile
	for i := 0; i < 8; i++ {
		leftPart, _ := c.Memory.ReadByteAtForced(tileStart + uint16(i*2))
		rightPart, _ := c.Memory.ReadByteAtForced(tileStart + uint16(i*2+1))

		tile.Lines[i] = [2]uint8{bits.Reverse8(leftPart), bits.Reverse8(rightPart)}
	}
	return tile
}

func ReadTile(tileDataOffset uint16, c *Cpu, isObject bool) Tile {

	// TODO check if WORKS

	var tileStart uint16
	_ = tileStart
	if isObject {
		// Object
		tileStart = 0x8000 + tileDataOffset
	} else {
		addressingMode := GetBit(c.Memory.Io.GetLCDC(), 4)
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
		leftPart, _ := c.Memory.ReadByteAt(tileStart + tileDataOffset + uint16(i*2))
		rightPart, _ := c.Memory.ReadByteAt(tileStart + tileDataOffset + uint16(i*2+1))

		tile.Lines[i] = [2]uint8{bits.Reverse8(leftPart), bits.Reverse8(rightPart)}
	}
	return tile
}

func (p *Ppu) FillTileViewerData() {
	bufW := 16 * 8
	for tileNum := 0; tileNum < 384; tileNum++ {

		tile := ReadTileDataBypass(uint16(tileNum), p.Cpu)

		var lines = tile.Lines

		tileX := tileNum % 16
		tileY := tileNum / 16

		for py := 0; py < 8; py++ {
			currentLine := lines[py]
			for px := 0; px < 8; px++ {
				colorLsb := 0
				colorMsb := 0
				if GetBit(currentLine[0], uint8(px)) {
					colorLsb = 1
				}
				if GetBit(currentLine[1], uint8(px)) {
					colorMsb = 1
				}
				colorBits := colorLsb | (colorMsb << 1)
				bufIdx := (tileY*8+py)*bufW + (tileX*8 + px)
				p.tileViewerBuf[bufIdx] = getTileColor(colorBits)
			}
		}
	}
}

func (p *Ppu) drawLine() {

}

func (p *Ppu) RenderObjOverview() {

	gl.BindTexture(gl.TEXTURE_2D, p.ObjOverviewTex)

	gl.Clear(gl.COLOR_BUFFER_BIT)

	// if GetBit(p.Cpu.Memory.Io.GetLCDC(), 1) {
	p.FillObjOverviewData()
	// }

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(8*8),
		int32(5*16),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(p.objOverviewBuf))
}

var BitMask = [...]uint8{0x1, 0x2, 0x4, 0x8, 0x10, 0x20, 0x40, 0x80}

func (p *Ppu) GenObjArray() []Object {

	all := make([]Object, 40)
	for i := 0; i < 40; i++ {
		oamBase := 0xFE00 + uint16(i*4)
		yPos, _ := p.Cpu.Memory.ReadByteAtForced(oamBase)
		xPos, _ := p.Cpu.Memory.ReadByteAtForced(oamBase + 1)
		tileInd, _ := p.Cpu.Memory.ReadByteAtForced(oamBase + 2)
		attributes, _ := p.Cpu.Memory.ReadByteAtForced(oamBase + 3)

		obj := Object{
			yPos:       yPos,
			xPos:       xPos,
			tileInd:    tileInd,
			attributes: attributes,
		}
		all[i] = obj
	}
	return all
}
func (p *Ppu) FillObjOverviewData() {
	isY16 := GetBit(p.Cpu.Memory.Io.GetLCDC(), 2) //0=8x8, 1=8x16
	// var objYSize uint8 = 8
	// if isY16 {
	// 	objYSize = 16
	// }
	objs := p.GenObjArray()
	for i := 0; i < len(objs); i++ {
		obj := objs[i]
		tile := ReadTileForObjects(uint16(obj.tileInd), p.Cpu)
		var tile2 Tile

		if isY16 {
			tile2 = ReadTileForObjects(uint16(obj.tileInd+16), p.Cpu)
		}

		if GetBit(obj.attributes, 6) {
			tile.FlipY()
			if isY16 {
				tile2.FlipY()
				tile2, tile = tile, tile2
			}
		}
		if GetBit(obj.attributes, 5) {
			tile.FlipX()
			if isY16 {
				tile2.FlipX()
			}
		}

		// TODO: implement attributes:
		// Priority : 0 = No, 1 = BG and Window color indices 1–3 are drawn over this OBJ

		var lines = tile.Lines
		var lines2 = tile2.Lines

		tileX := i % 8
		tileY := i / 8

		bufW := 8 * 8

		for py := 0; py < 16; py++ {
			var currentLine [2]uint8
			if py >= 8 {
				currentLine = lines2[py-8]
			} else {
				currentLine = lines[py]
			}
			for px := 0; px < 8; px++ {
				colorLsb := 0
				colorMsb := 0
				if GetBit(currentLine[0], uint8(px)) {
					colorLsb = 1
				}
				if GetBit(currentLine[1], uint8(px)) {
					colorMsb = 1
				}
				colorBits := colorLsb | (colorMsb << 1)
				bufIdx := (tileY*16+py)*bufW + (tileX*8 + px)
				p.objOverviewBuf[bufIdx] = getTileColor(colorBits)
			}
		}
	}

	// render overview of sprites in objOverviewBuf, maybe 4 sprites per tile, with some indication of priority and palette?
}

func (p *Ppu) RenderTileViewer() {
	gl.BindTexture(gl.TEXTURE_2D, p.TileViewerTex)

	gl.Clear(gl.COLOR_BUFFER_BIT)

	p.FillTileViewerData()

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(16*8),
		int32(24*8),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(p.tileViewerBuf))
}

func (p *Ppu) RenderWindowMapViewer() {

	gl.BindTexture(gl.TEXTURE_2D, p.WindowTex)

	gl.Clear(gl.COLOR_BUFFER_BIT)

	if GetBit(p.Cpu.Memory.Io.GetLCDC(), 5) { //&& GetBit(p.Cpu.Memory.Io.GetLCDC(), 0) {
		p.FillWindowMapData()
	}

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(256),
		int32(256),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(p.windowBuf))
}

func (p *Ppu) RenderBackgroundMapViewer() {
	gl.BindTexture(gl.TEXTURE_2D, p.BackgroundTex)

	gl.Clear(gl.COLOR_BUFFER_BIT)

	// if GetBit(p.Cpu.Memory.Io.GetLCDC(), 0) {
	p.FillBackgroundMapData()
	// }

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(256),
		int32(256),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(p.backgroundBuf))
}

var texData = []uint8{255, 255, 255, 255}

func (p *Ppu) Render() {
	clear(p.viewportBuf)
	// TODO: with current rendering this shows white screen 99% of the time
	// maybe we need to set ly = 0 when LCDC.7 is set to off? or this just doesnt work
	// when rendering like this
	// if GetBit(p.Cpu.Memory.Io.GetLCDC(), 7) {
	// 	gl.BindTexture(gl.TEXTURE_2D, p.ViewPortTex)
	// 	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, (gl.Ptr(texData)))
	// 	return
	// }

	// BG and window rendering to viewportTex
	// TODO: palette selection for BG and window
	if GetBit(p.Cpu.Memory.Io.GetLCDC(), 0) {
		//Fill BG
		p.FillBackgroundMapData()

		//probably faster way to to do this in gpu but this is not final way of rendering anyway
		bgSubImage := extractRect(p.backgroundBuf,
			int(p.Cpu.Memory.Io.GetSCX()),
			int(p.Cpu.Memory.Io.GetSCY()),
			GB_WINDOW_WIDTH,
			GB_WINDOW_HEIGHT,
			256)

		gl.BindTexture(gl.TEXTURE_2D, p.ViewPortTex)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(GB_WINDOW_WIDTH),
			int32(GB_WINDOW_HEIGHT),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(bgSubImage))

		if GetBit(p.Cpu.Memory.Io.GetLCDC(), 5) {

			// if this doesnt work:
			// render viewporetWinTex like viewportTex
			// then blend like with obj tex
			p.FillWindowMapData()
			windowSubImage := extractRect(p.windowBuf,
				0,
				0,
				GB_WINDOW_WIDTH,
				GB_WINDOW_HEIGHT,
				256)

			gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 256)

			xofs := int32(p.Cpu.Memory.Io.GetWX() - 7)
			yofs := int32(p.Cpu.Memory.Io.GetWY())
			gl.TexSubImage2D(gl.TEXTURE_2D, 0, xofs, yofs, int32(GB_WINDOW_WIDTH), int32(GB_WINDOW_HEIGHT), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(windowSubImage))

			gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
		}
	}

	// OBJ rendering to own tex
	// TODO: obj priority and palette selection
	if GetBit(p.Cpu.Memory.Io.GetLCDC(), 1) {
		clear(p.viewportObjBuf)

		isY16 := GetBit(p.Cpu.Memory.Io.GetLCDC(), 2) //0=8x8, 1=8x16
		objs := p.GenObjArray()
		for i := 0; i < len(objs); i++ {
			obj := objs[i]
			tile := ReadTileForObjects(uint16(obj.tileInd), p.Cpu)
			var tile2 Tile

			if isY16 {
				tile2 = ReadTileForObjects(uint16(obj.tileInd+16), p.Cpu)
			}

			if GetBit(obj.attributes, 6) {
				tile.FlipY()
				if isY16 {
					tile2.FlipY()
					tile2, tile = tile, tile2
				}
			}
			if GetBit(obj.attributes, 5) {
				tile.FlipX()
				if isY16 {
					tile2.FlipX()
				}
			}

			xPos := int32(obj.xPos - 8)
			yPos := int32(obj.yPos - 16)

			var lines = tile.Lines
			var lines2 = tile2.Lines

			var bufW int32 = int32(GB_WINDOW_WIDTH)

			for py := int32(0); py < 16; py++ {
				var currentLine [2]uint8
				if py >= 8 {
					currentLine = lines2[py-8]
				} else {
					currentLine = lines[py]
				}
				for px := int32(0); px < 8; px++ {
					colorLsb := 0
					colorMsb := 0
					if GetBit(currentLine[0], uint8(px)) {
						colorLsb = 1
					}
					if GetBit(currentLine[1], uint8(px)) {
						colorMsb = 1
					}
					colorBits := colorLsb | (colorMsb << 1)
					// bufIdx := (tileY*16+py)*bufW + (tileX*8 + px)
					bufIdx := (yPos+py)*bufW + xPos + px
					if bufIdx >= 0 && bufIdx < int32(len(p.viewportObjBuf)) {
						p.viewportObjBuf[bufIdx] = getTileColor(colorBits)
					}
				}
			}
		}
		gl.BindTexture(gl.TEXTURE_2D, p.ViewPortObjTex)

		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(GB_WINDOW_WIDTH),
			int32(GB_WINDOW_HEIGHT),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(p.viewportObjBuf))
	}

	// 1. Draw the first layer (Viewport)
	gl.BindTexture(gl.TEXTURE_2D, p.ViewPortTex)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	// Because gl.BLEND is on, this will draw OVER the first one
	gl.BindTexture(gl.TEXTURE_2D, p.ViewPortObjTex)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)

}

func extractRect(src []color.RGBA, x, y, w, h, stride int) []color.RGBA {
	dest := make([]color.RGBA, w*h)
	for cH := 0; cH < h; cH++ {
		srcY := y*stride + cH*stride
		for cW := 0; cW < w; cW++ {
			srcX := x + cW
			destIdx := cH*w + cW
			srcIdx := srcY + srcX
			dest[destIdx] = src[srcIdx]
			// srcIdx := (y+cH)*stride + (x + cW)
			// destIdx := (cH*w + cW)
			// dest[destIdx] = src[srcIdx]

		}
	}
	return dest
}

func (p *Ppu) GetTexture() uint32 {
	return p.ViewPortTex
}

func (p *Ppu) GetPixelBuffer() []color.RGBA {
	return p.viewportBuf
}
