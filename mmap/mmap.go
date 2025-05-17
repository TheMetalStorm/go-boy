package mmap

import (
	"fmt"
	"go-boy/ioregs"
)

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

func (m *Mmap) SetValue(address uint16, value uint8) {

	switch {
	case address < 0x4000:
		m.bank0[address] = value
	case address < 0x8000:
		m.bank1[address-0x4000] = value
	case address < 0xa000:
		m.vram[address-0x8000] = value
	case address < 0xc000:
		m.extram[address-0xa000] = value
	case address < 0xd000:
		m.wram1[address-0xc000] = value
	case address < 0xe000:
		m.wram2[address-0xd000] = value
	case address < 0xfe00:
		m.echoram[address-0xe000] = value
	case address < 0xff00:
		m.oam[address-0xfe00] = value
	case address < 0xff80:
		m.nu[address-0xff00] = value
	case address < 0xfffe:
		m.io.SetAtAdress(address-0xff80, value)
	case address == 0xfffe:
		m.hram[address-0xff80] = value
	case address == 0xffff:
		m.ie = value
	}
}

func (m *Mmap) Read16At(address uint16) (data uint16, numReadBytes uint16) {
	switch {
	case address < 0x4000:
		a1 := uint16(m.bank0[address])
		a2 := uint16(m.bank0[address+1])
		return uint16(a1 | a2<<8), 2
	case address < 0x8000:
		a1 := uint16(m.bank1[address-0x4000])
		a2 := uint16(m.bank1[address-0x4000+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xa000:
		a1 := uint16(m.vram[address-0x8000])
		a2 := uint16(m.vram[address-0x8000+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xc000:
		a1 := uint16(m.extram[address-0xa000])
		a2 := uint16(m.extram[address-0xa000+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xd000:
		a1 := uint16(m.wram1[address-0xc000])
		a2 := uint16(m.wram1[address-0xc000+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xe000:
		a1 := uint16(m.wram2[address-0xd000])
		a2 := uint16(m.wram2[address-0xd000+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xfe00:
		a1 := uint16(m.echoram[address-0xe000])
		a2 := uint16(m.echoram[address-0xe000+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xff00:
		a1 := uint16(m.oam[address-0xfe00])
		a2 := uint16(m.oam[address-0xfe00+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xff80:
		a1 := uint16(m.nu[address-0xff00])
		a2 := uint16(m.nu[address-0xff00+1])
		return uint16(a1 | a2<<8), 2
	case address < 0xfffe:
		a1 := uint16(m.io.Regs[address-0xff80])
		a2 := uint16(m.io.Regs[address-0xff80+1])
		return uint16(a1 | a2<<8), 2
	case address == 0xfffe:
		a1 := uint16(m.hram[address-0xff80])
		a2 := uint16(m.hram[address-0xff80+1])
		return uint16(a1 | a2<<8), 2
	}
	return 0, 2
}

func (m *Mmap) ReadByteAt(address uint16) (val uint8, bytesRead uint16) {

	switch {
	case address < 0x4000:
		return m.bank0[address], 1
	case address < 0x8000:
		return m.bank1[address-0x4000], 1
	case address < 0xa000:
		return m.vram[address-0x8000], 1
	case address < 0xc000:
		return m.extram[address-0xa000], 1
	case address < 0xd000:
		return m.wram1[address-0xc000], 1
	case address < 0xe000:
		return m.wram2[address-0xd000], 1
	case address < 0xfe00:
		return m.echoram[address-0xe000], 1
	case address < 0xff00:
		return m.oam[address-0xfe00], 1
	case address < 0xff80:
		return m.nu[address-0xff00], 1
	case address < 0xfffe:
		return m.io.Regs[address-0xff80], 1
	case address == 0xfffe:
		return m.hram[address-0xff80], 1
	case address == 0xffff:
		return m.ie, 1
	}
	return 0, 0
}

// Dump

func (m *Mmap) Dump() {
	m.DumpBank0()
	m.DumpBank1()
	m.DumpVram()
	m.DumpExtram()
	m.DumpWram1()
	m.DumpWram2()
	m.DumpEchoram()
	m.DumpOam()
	m.DumpNu()
	m.DumpIo()
	m.DumpHram()
	m.DumpIe()
}

func (m *Mmap) DumpBank0() {
	fmt.Println("bank0:")
	m.dumpMemory(m.bank0[:], 0x0000)
}

func (m *Mmap) DumpBank1() {
	fmt.Println("bank1:")
	m.dumpMemory(m.bank1[:], 0x4000)
}

func (m *Mmap) DumpVram() {
	fmt.Println("vram:")
	m.dumpMemory(m.vram[:], 0x8000)
}

func (m *Mmap) DumpExtram() {
	fmt.Println("extram:")
	m.dumpMemory(m.extram[:], 0xa000)
}

func (m *Mmap) DumpWram1() {
	fmt.Println("wram1:")
	m.dumpMemory(m.wram1[:], 0xc000)
}

func (m *Mmap) DumpWram2() {
	fmt.Println("wram2:")
	m.dumpMemory(m.wram2[:], 0xd000)
}

func (m *Mmap) DumpEchoram() {
	fmt.Println("echoram:")
	m.dumpMemory(m.echoram[:], 0xe000)
}

func (m *Mmap) DumpOam() {
	fmt.Println("oam:")
	m.dumpMemory(m.oam[:], 0xfe00)
}

func (m *Mmap) DumpNu() {
	fmt.Println("nu:")
	m.dumpMemory(m.nu[:], 0xff00)
}

func (m *Mmap) DumpIo() {
	fmt.Println("io:")
	m.io.Dump()
}

func (m *Mmap) DumpHram() {
	fmt.Println("hram:")
	m.dumpMemory(m.hram[:], 0xff80)
}

func (m *Mmap) DumpIe() {
	fmt.Printf("ie: %02x\n", m.ie)
}

func (m *Mmap) dumpMemory(memory []uint8, baseAddress uint16) {
	for i := 0; i < len(memory); i += 16 {
		fmt.Printf("%04x ", baseAddress+uint16(i))
		for j := 0; j < 16 && i+j < len(memory); j++ {
			fmt.Printf("%02x ", memory[i+j])
		}
		fmt.Println()
	}
}
