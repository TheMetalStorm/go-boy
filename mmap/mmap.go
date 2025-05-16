package mmap

import "go-boy/ioregs"

type Ioregs = ioregs.Ioregs
type Mmap struct {
	bank0 [0x4000]uint8 // 16 KiB ROM bank 00
	bank1 [0x4000]uint8 // 16 KiB ROM Bank 01–NN

	vram   [0x2000]uint8 // 8 KiB Video RAM (VRAM)
	extram [0x2000]uint8 // 8 KiB External RAM

	wram1 [0x1000]uint8 // 4 KiB Work RAM (WRAM)
	wram2 [0x1000]uint8 // 4 KiB Work RAM (WRAM)

	echoram [0x1e00]uint8 // Echo Ram (mirror of C000–DDFF)

	oam  [0xa0]uint8 //Object attribute memory (OAM)
	nu   [0x60]uint8 //not usable
	io   Ioregs      // I/O Reg
	hram [0x7e]uint8 //high ram
	ie   uint8       //interruot enable reg
}
