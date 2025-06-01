package mmap

import (
	"fmt"
	"go-boy/ioregs"
)

type Ioregs = ioregs.Ioregs
type Mmap struct {
	Bank0 [0x4000]uint8 // 16 KiB ROM bank 00
	Bank1 [0x4000]uint8 // 16 KiB ROM Bank 01–NN

	Vram   [0x2000]uint8 // 8 KiB Video RAM (VRAM)
	Extram [0x2000]uint8 // 8 KiB External RAM

	Wram1 [0x1000]uint8 // 4 KiB Work RAM (WRAM)
	Wram2 [0x1000]uint8 // 4 KiB Work RAM (WRAM)

	Echoram [0x1e00]uint8 // Echo Ram (mirror of C000–DDFF)

	Oam  [0xa0]uint8 //Object attribute memory (OAM)
	Nu   [0x60]uint8 //not usable
	Io   Ioregs      // I/O Reg
	Hram [0x7f]uint8 //high ram
	Ie   uint8       //interrupt enable reg
}

func (m *Mmap) SetValue(address uint16, value uint8) {

	switch {
	case address < 0x4000:
		m.Bank0[address] = value

	case address < 0x8000:
		m.Bank1[address-0x4000] = value

	case address < 0xA000:
		m.Vram[address-0x8000] = value

	case address < 0xC000:
		m.Extram[address-0xA000] = value

	case address < 0xD000:
		m.Wram1[address-0xC000] = value

	case address < 0xE000:
		m.Wram2[address-0xD000] = value

	case address < 0xFE00:
		m.Echoram[address-0xE000] = value

	case address < 0xFEA0:
		m.Oam[address-0xFE00] = value

	case address < 0xFF00:
		m.Nu[address-0xFEA0] = value

	case address < 0xFF80:
		m.Hram[address-0xFF00] = value

	case address < 0xFFFF:
		m.Hram[address-0xFF80] = value

	case address == 0xFFFF:
		m.Ie = value
	}

}

func (m *Mmap) Read16At(address uint16) (data uint16, numReadBytes uint16) {

	switch {
	case address < 0x4000-1:
		a1 := uint16(m.Bank0[address])
		a2 := uint16(m.Bank0[address+1])
		return uint16(a1 | a2<<8), 2

	case address < 0x8000-1:
		a1 := uint16(m.Bank1[address-0x4000])
		a2 := uint16(m.Bank1[address-0x4000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xA000-1:
		a1 := uint16(m.Vram[address-0x8000])
		a2 := uint16(m.Vram[address-0x8000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xC000-1:
		a1 := uint16(m.Extram[address-0xa000])
		a2 := uint16(m.Extram[address-0xa000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xD000-1:
		a1 := uint16(m.Wram1[address-0xc000])
		a2 := uint16(m.Wram1[address-0xc000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xE000-1:
		a1 := uint16(m.Wram2[address-0xd000])
		a2 := uint16(m.Wram2[address-0xd000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFE00-1:
		a1 := uint16(m.Echoram[address-0xe000])
		a2 := uint16(m.Echoram[address-0xe000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFEA0-1:
		a1 := uint16(m.Oam[address-0xFE00])
		a2 := uint16(m.Oam[address-0xFE00+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFF00-1:
		a1 := uint16(m.Nu[address-0xFEA0])
		a2 := uint16(m.Nu[address-0xFEA0+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFF80-1:
		a1 := uint16(m.Io.Regs[address-0xFF00])
		a2 := uint16(m.Io.Regs[address-0xFF00+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFFFE-1:
		a1 := uint16(m.Hram[address-0xff80])
		a2 := uint16(m.Hram[address-0xff80+1])
		return uint16(a1 | a2<<8), 2

	}
	return 0, 0
}

func (m *Mmap) ReadByteAt(address uint16) (val uint8, bytesRead uint16) {

	switch {
	case address < 0x4000:
		return m.Bank0[address], 1

	case address < 0x8000:
		return m.Bank1[address-0x4000], 1

	case address < 0xA000:
		return m.Vram[address-0x8000], 1

	case address < 0xC000:
		return m.Extram[address-0xA000], 1

	case address < 0xD000:
		return m.Wram1[address-0xC000], 1

	case address < 0xE000:
		return m.Wram2[address-0xD000], 1

	case address < 0xFE00:
		return m.Echoram[address-0xE000], 1

	case address < 0xFEA0:
		return m.Oam[address-0xFE00], 1

	case address < 0xFF00:
		return m.Nu[address-0xFEA0], 1

	case address < 0xFF80:
		return m.Hram[address-0xFF00], 1

	case address < 0xFFFF:
		return m.Hram[address-0xFF80], 1

	case address == 0xFFFF:
		return m.Ie, 1
	}
	return 0, 0
}

func (m *Mmap) SetInterruptEnabledBit(bit ioregs.InterruptFlags, cond bool) {
	if cond {
		m.Ie |= (1 << bit)
	} else {
		m.Ie &^= (1 << bit)
	}
}

func (m *Mmap) GetInterruptEnabledBit(bit ioregs.InterruptFlags) bool {
	ie := m.Ie
	sisSet := (ie >> bit) & 0x1
	if sisSet == 1 {
		return true
	} else {
		return false
	}
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
	m.dumpMemory(m.Bank0[:], 0x0000)
}

func (m *Mmap) DumpBank1() {
	fmt.Println("bank1:")
	m.dumpMemory(m.Bank1[:], 0x4000)
}

func (m *Mmap) DumpVram() {
	fmt.Println("vram:")
	m.dumpMemory(m.Vram[:], 0x8000)
}

func (m *Mmap) DumpExtram() {
	fmt.Println("extram:")
	m.dumpMemory(m.Extram[:], 0xa000)
}

func (m *Mmap) DumpWram1() {
	fmt.Println("wram1:")
	m.dumpMemory(m.Wram1[:], 0xc000)
}

func (m *Mmap) DumpWram2() {
	fmt.Println("wram2:")
	m.dumpMemory(m.Wram2[:], 0xd000)
}

func (m *Mmap) DumpEchoram() {
	fmt.Println("echoram:")
	m.dumpMemory(m.Echoram[:], 0xe000)
}

func (m *Mmap) DumpOam() {
	fmt.Println("oam:")
	m.dumpMemory(m.Oam[:], 0xfe00)
}

func (m *Mmap) DumpNu() {
	fmt.Println("nu:")
	m.dumpMemory(m.Nu[:], 0xff00)
}

func (m *Mmap) DumpIo() {
	fmt.Println("io:")
	m.Io.Dump()
}

func (m *Mmap) DumpHram() {
	fmt.Println("hram:")
	m.dumpMemory(m.Hram[:], 0xff80)
}

func (m *Mmap) DumpIe() {
	fmt.Printf("ie: %02x\n", m.Ie)
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
