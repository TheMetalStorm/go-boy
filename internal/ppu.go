package internal

import (
	"image/color"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Ppu struct {
	screenMultiplier int
	running          bool

	ViewPortTex   uint32
	BackgroundTex uint32
	WindowTex     uint32
	TileViewerTex uint32

	viewportBuf   []color.RGBA
	backgroundBuf []color.RGBA
	windowBuf     []color.RGBA
	tileViewerBuf []color.RGBA

	CurrentMode PpuMode
	CurrentDot  uint64

	Cpu *Cpu

	HandleGLUpdate bool
	Frame          uint64
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF
var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

const BG_WINDOW_X_Y int = 256

var quadVAO uint32
var quadVBO uint32

func NewPpu(screenMultiplier int) *Ppu {
	ppu := &Ppu{}
	ppu.screenMultiplier = screenMultiplier
	ppu.Restart(screenMultiplier)

	ppu.viewportBuf = make([]color.RGBA, GB_WINDOW_WIDTH*GB_WINDOW_HEIGHT)

	ppu.backgroundBuf = make([]color.RGBA, BG_WINDOW_X_Y*BG_WINDOW_X_Y)

	ppu.windowBuf = make([]color.RGBA, BG_WINDOW_X_Y*BG_WINDOW_X_Y)

	ppu.tileViewerBuf = make([]color.RGBA, 16*8*24*8)

	return ppu
}

func (p *Ppu) Restart(screenMultiplier int) {
	p.screenMultiplier = screenMultiplier
	p.running = true
	p.CurrentDot = 0
	p.CurrentMode = MODE_2

	//TODO: clear buffers (and tex)?
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

			//draw current line!
			if p.Cpu.Memory.Io.GetLY() < 144 {
				p.drawLine()
			}

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
		return color.RGBA{255, 255, 0, 255}
	}
}

func (p *Ppu) FillTileViewerData() {
	bufW := 16 * 8
	for tileNum := 0; tileNum < 384; tileNum++ {
		var lines [8]uint16
		for i := 0; i < 8; i++ {
			leftPart, _ := p.Cpu.Memory.ReadByteAtForced(uint16(TILE_DATA_START) + uint16(tileNum)*16 + uint16(i*2))
			rightPart, _ := p.Cpu.Memory.ReadByteAtForced(uint16(TILE_DATA_START) + uint16(tileNum)*16 + uint16(i*2+1))
			lines[i] = uint16(leftPart) | uint16(rightPart)<<8
		}

		tileX := tileNum % 16
		tileY := tileNum / 16

		for py := 0; py < 8; py++ {
			currentLine := lines[py]
			for px := 0; px < 8; px++ {
				colorLsb := 0
				colorMsb := 0
				if GetBit16(currentLine, uint8(px)) {
					colorLsb = 1
				}
				if GetBit16(currentLine, uint8(8+px)) {
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

func (p *Ppu) RenderBG() {
	gl.BindTexture(gl.TEXTURE_2D, p.BackgroundTex)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(BG_WINDOW_X_Y),
		int32(BG_WINDOW_X_Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(p.backgroundBuf))

}

func (p *Ppu) RenderTileViewer() {
	p.FillTileViewerData()
	gl.BindTexture(gl.TEXTURE_2D, p.TileViewerTex)

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
func (p *Ppu) Render() {

	//for now just render background
	p.RenderBG()
	return
	// // Clear to black
	// for i := range p.viewportBuf {
	// 	p.viewportBuf[i] = color.RGBA{0, 0, 0, 255}
	// }

	// // Moving box
	// boxW, boxH := 20, 20
	// speed := 2
	// x := int(p.Frame*uint64(speed)) % (GB_WINDOW_WIDTH - boxW)
	// y := int(p.Frame*uint64(speed)) % (GB_WINDOW_HEIGHT - boxH)

	// for py := y; py < y+boxH; py++ {
	// 	for px := x; px < x+boxW; px++ {
	// 		p.viewportBuf[py*GB_WINDOW_WIDTH+px] = color.RGBA{255, 0, 0, 255}
	// 	}
	// }

	// p.Frame++

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
		gl.Ptr(p.viewportBuf))
}

func (p *Ppu) GetTexture() uint32 {
	return p.ViewPortTex
}

func (p *Ppu) GetPixelBuffer() []color.RGBA {
	return p.viewportBuf
}
