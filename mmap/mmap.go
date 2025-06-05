package mmap

import (
	"go-boy/ioregs"
)

type Ioregs = ioregs.Ioregs
type Mmap struct {
	bank0 [0x4000]uint8 // 16 KiB ROM bank 00
	bankN [0x4000]uint8 // 16 KiB ROM Bank 01–NN

	vram   [0x2000]uint8 // 8 KiB Video RAM (VRAM)
	extram [0x2000]uint8 // 8 KiB External RAM

	wram1 [0x1000]uint8 // 4 KiB Work RAM (WRAM)
	wram2 [0x1000]uint8 // 4 KiB Work RAM (WRAM)

	echoram [0x1e00]uint8 // Echo Ram (mirror of C000–DDFF)

	Oam  [0xa0]uint8 //Object attribute memory (OAM)
	nu   [0x60]uint8 //not usable
	Io   Ioregs      // I/O Reg
	Hram [0x7f]uint8 //high ram
	Ie   uint8       //interrupt enable reg
}

func (m *Mmap) SetValue(address uint16, value uint8) {

	switch {
	case address < 0x4000:
		m.bank0[address] = value

	case address < 0x8000:
		m.bankN[address-0x4000] = value

	case address < 0xA000:
		m.vram[address-0x8000] = value

	case address < 0xC000:
		m.extram[address-0xA000] = value

	case address < 0xD000:
		m.wram1[address-0xC000] = value

	case address < 0xE000:
		m.wram2[address-0xD000] = value

	case address < 0xFE00:
		m.echoram[address-0xE000] = value

	case address < 0xFEA0:
		m.Oam[address-0xFE00] = value

	case address < 0xFF00:
		m.nu[address-0xFEA0] = value

	case address < 0xFF80:
		m.Io.Regs[address-0xFF00] = value

	case address < 0xFFFF:
		m.Hram[address-0xFF80] = value

	case address == 0xFFFF:
		m.Ie = value
	}

}

func (m *Mmap) Read16At(address uint16) (data uint16, numReadBytes uint16) {

	switch {
	case address < 0x4000-1:
		a1 := uint16(m.bank0[address])
		a2 := uint16(m.bank0[address+1])
		return uint16(a1 | a2<<8), 2

	case address < 0x8000-1:
		a1 := uint16(m.bankN[address-0x4000])
		a2 := uint16(m.bankN[address-0x4000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xA000-1:
		a1 := uint16(m.vram[address-0x8000])
		a2 := uint16(m.vram[address-0x8000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xC000-1:
		a1 := uint16(m.extram[address-0xa000])
		a2 := uint16(m.extram[address-0xa000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xD000-1:
		a1 := uint16(m.wram1[address-0xc000])
		a2 := uint16(m.wram1[address-0xc000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xE000-1:
		a1 := uint16(m.wram2[address-0xd000])
		a2 := uint16(m.wram2[address-0xd000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFE00-1:
		a1 := uint16(m.echoram[address-0xe000])
		a2 := uint16(m.echoram[address-0xe000+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFEA0-1:
		a1 := uint16(m.Oam[address-0xFE00])
		a2 := uint16(m.Oam[address-0xFE00+1])
		return uint16(a1 | a2<<8), 2

	case address < 0xFF00-1:
		a1 := uint16(m.nu[address-0xFEA0])
		a2 := uint16(m.nu[address-0xFEA0+1])
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
		return m.bank0[address], 1

	case address < 0x8000:
		return m.bankN[address-0x4000], 1

	case address < 0xA000:
		return m.vram[address-0x8000], 1

	case address < 0xC000:
		return m.extram[address-0xA000], 1

	case address < 0xD000:
		return m.wram1[address-0xC000], 1

	case address < 0xE000:
		return m.wram2[address-0xD000], 1

	case address < 0xFE00:
		return m.echoram[address-0xE000], 1

	case address < 0xFEA0:
		return m.Oam[address-0xFE00], 1

	case address < 0xFF00:
		return m.nu[address-0xFEA0], 1

	case address < 0xFF80:
		return m.Io.Regs[address-0xFF00], 1

	case address < 0xFFFF:
		return m.Hram[address-0xFF80], 1

	case address == 0xFFFF:
		return m.Ie, 1
	}
	return 0, 0
}

func SetBit(ptr *uint8, bit uint8, cond bool) {
	if cond {
		*ptr |= (1 << bit)
	} else {
		*ptr &^= (1 << bit)
	}
}

func GetBit(num uint8, bit uint8) bool {
	res := (num >> bit) & 1
	return res != 0
}

func (m *Mmap) SetInterruptEnabledBit(bit ioregs.InterruptFlags, cond bool) {
	SetBit(&m.Ie, uint8(bit), cond)
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

// func (m *Mmap) Dump() {
// 	m.DumpBank0()
// 	m.DumpBank1()
// 	m.DumpVram()
// 	m.DumpExtram()
// 	m.DumpWram1()
// 	m.DumpWram2()
// 	m.DumpEchoram()
// 	m.DumpOam()
// 	m.DumpNu()
// 	m.DumpIo()
// 	m.DumpHram()
// 	m.DumpIe()
// }

// func (m *Mmap) DumpBank0() {
// 	fmt.Println("bank0:")
// 	m.dumpMemory(m.Bank0[:], 0x0000)
// }

// func (m *Mmap) DumpBank1() {
// 	fmt.Println("bank1:")
// 	m.dumpMemory(m.Bank1[:], 0x4000)
// }

// func (m *Mmap) DumpVram() {
// 	fmt.Println("vram:")
// 	m.dumpMemory(m.Vram[:], 0x8000)
// }

// func (m *Mmap) DumpExtram() {
// 	fmt.Println("extram:")
// 	m.dumpMemory(m.Extram[:], 0xa000)
// }

// func (m *Mmap) DumpWram1() {
// 	fmt.Println("wram1:")
// 	m.dumpMemory(m.Wram1[:], 0xc000)
// }

// func (m *Mmap) DumpWram2() {
// 	fmt.Println("wram2:")
// 	m.dumpMemory(m.Wram2[:], 0xd000)
// }

// func (m *Mmap) DumpEchoram() {
// 	fmt.Println("echoram:")
// 	m.dumpMemory(m.Echoram[:], 0xe000)
// }

// func (m *Mmap) DumpOam() {
// 	fmt.Println("oam:")
// 	m.dumpMemory(m.Oam[:], 0xfe00)
// }

// func (m *Mmap) DumpNu() {
// 	fmt.Println("nu:")
// 	m.dumpMemory(m.Nu[:], 0xff00)
// }

// func (m *Mmap) DumpIo() {
// 	fmt.Println("io:")
// 	m.Io.Dump()
// }

// func (m *Mmap) DumpHram() {
// 	fmt.Println("hram:")
// 	m.dumpMemory(m.Hram[:], 0xff80)
// }

// func (m *Mmap) DumpIe() {
// 	fmt.Printf("ie: %02x\n", m.Ie)
// }

// func (m *Mmap) dumpMemory(memory []uint8, baseAddress uint16) {
// 	for i := 0; i < len(memory); i += 16 {
// 		fmt.Printf("%04x ", baseAddress+uint16(i))
// 		for j := 0; j < 16 && i+j < len(memory); j++ {
// 			fmt.Printf("%02x ", memory[i+j])
// 		}
// 		fmt.Println()
// 	}
// }

// Getters for memory-mapped regions
func (m *Mmap) GetBank0() []uint8 {
	return m.bank0[:]
}

func (m *Mmap) GetBank1() []uint8 {
	return m.bankN[:]
}

func (m *Mmap) GetVram() []uint8 {
	return m.vram[:]
}

func (m *Mmap) GetExtram() []uint8 {
	return m.extram[:]
}

func (m *Mmap) GetWram1() []uint8 {
	return m.wram1[:]
}

func (m *Mmap) GetWram2() []uint8 {
	return m.wram2[:]
}

func (m *Mmap) GetEchoram() []uint8 {
	return m.echoram[:]
}

func (m *Mmap) GetOam() []uint8 {
	return m.Oam[:]
}

func (m *Mmap) GetNu() []uint8 {
	return m.nu[:]
}

func (m *Mmap) GetIo() Ioregs {
	return m.Io
}

func (m *Mmap) GetHram() []uint8 {
	return m.Hram[:]
}

func (m *Mmap) GetIe() uint8 {
	return m.Ie
}
