package internal

import "github.com/go-gl/gl/v3.3-core/gl"

type Ppu struct {
	screenMultiplier int
	running          bool
	texture          uint32
	pixelBuffer      []uint8

	CurrentMode PpuMode
	CurrentDot  uint64

	Cpu *Cpu

	HandleGLUpdate bool
}

var TILE_DATA_START int = 0x8000
var TILE_DATA_END int = 0x97FF
var GB_WINDOW_WIDTH int = 160
var GB_WINDOW_HEIGHT int = 144

var quadVAO uint32
var quadVBO uint32

func NewPpu(screenMultiplier int) *Ppu {
	ppu := &Ppu{}
	ppu.screenMultiplier = screenMultiplier
	ppu.Restart(screenMultiplier)

	ppu.pixelBuffer = make([]uint8, GB_WINDOW_WIDTH*GB_WINDOW_HEIGHT*4)

	return ppu
}

func (p *Ppu) Restart(screenMultiplier int) {
	p.screenMultiplier = screenMultiplier
	p.running = true
	p.CurrentDot = 0
	p.CurrentMode = MODE_2
}

func (p *Ppu) Step(ranMCyclesThisStep uint64) {

	mode3Duration := uint64(172)
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

func (p *Ppu) drawLine() {

}

func (p *Ppu) Render(texture uint32) {
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(GB_WINDOW_WIDTH),
		int32(GB_WINDOW_HEIGHT),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(p.pixelBuffer))
}

func (p *Ppu) GetTexture() uint32 {
	return p.texture
}

func (p *Ppu) GetPixelBuffer() []uint8 {
	return p.pixelBuffer
}
