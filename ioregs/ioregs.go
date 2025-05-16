package ioregs

import "fmt"

type Ioregs struct {
	Regs [0x80]uint8
}

func (i *Ioregs) GetJOYP() uint8 {
	return i.Regs[0x00]
}

func (i *Ioregs) SetAtAdress(add uint16, val uint8) {
	i.Regs[add] = val
}

func (i *Ioregs) SetJOYP(value uint8) {
	i.Regs[0x00] = value
}

func (i *Ioregs) GetSB() uint8 {
	return i.Regs[0x01]
}

func (i *Ioregs) SetSB(value uint8) {
	i.Regs[0x01] = value
}

func (i *Ioregs) GetSC() uint8 {
	return i.Regs[0x02]
}

func (i *Ioregs) SetSC(value uint8) {
	i.Regs[0x02] = value
}

func (i *Ioregs) GetDIV() uint8 {
	return i.Regs[0x04]
}

func (i *Ioregs) SetDIV(value uint8) {
	i.Regs[0x04] = value
}

func (i *Ioregs) GetTIMA() uint8 {
	return i.Regs[0x05]
}

func (i *Ioregs) SetTIMA(value uint8) {
	i.Regs[0x05] = value
}

func (i *Ioregs) GetTMA() uint8 {
	return i.Regs[0x06]
}

func (i *Ioregs) SetTMA(value uint8) {
	i.Regs[0x06] = value
}

func (i *Ioregs) GetTAC() uint8 {
	return i.Regs[0x07]
}

func (i *Ioregs) SetTAC(value uint8) {
	i.Regs[0x07] = value
}

func (i *Ioregs) GetIF() uint8 {
	return i.Regs[0x0F]
}

func (i *Ioregs) SetIF(value uint8) {
	i.Regs[0x0F] = value
}

func (i *Ioregs) GetNR10() uint8 {
	return i.Regs[0x10]
}

func (i *Ioregs) SetNR10(value uint8) {
	i.Regs[0x10] = value
}

func (i *Ioregs) GetNR11() uint8 {
	return i.Regs[0x11]
}

func (i *Ioregs) SetNR11(value uint8) {
	i.Regs[0x11] = value
}

func (i *Ioregs) GetNR12() uint8 {
	return i.Regs[0x12]
}

func (i *Ioregs) SetNR12(value uint8) {
	i.Regs[0x12] = value
}

func (i *Ioregs) SetNR13(value uint8) {
	i.Regs[0x13] = value
}

func (i *Ioregs) GetNR14() uint8 {
	return i.Regs[0x14]
}

func (i *Ioregs) SetNR14(value uint8) {
	i.Regs[0x14] = value
}

func (i *Ioregs) GetNR21() uint8 {
	return i.Regs[0x16]
}

func (i *Ioregs) SetNR21(value uint8) {
	i.Regs[0x16] = value
}

func (i *Ioregs) GetNR22() uint8 {
	return i.Regs[0x17]
}

func (i *Ioregs) SetNR22(value uint8) {
	i.Regs[0x17] = value
}

func (i *Ioregs) SetNR23(value uint8) {
	i.Regs[0x18] = value
}

func (i *Ioregs) GetNR24() uint8 {
	return i.Regs[0x19]
}

func (i *Ioregs) SetNR24(value uint8) {
	i.Regs[0x19] = value
}

func (i *Ioregs) GetNR30() uint8 {
	return i.Regs[0x1A]
}

func (i *Ioregs) SetNR30(value uint8) {
	i.Regs[0x1A] = value
}

func (i *Ioregs) SetNR31(value uint8) {
	i.Regs[0x1B] = value
}

func (i *Ioregs) GetNR32() uint8 {
	return i.Regs[0x1C]
}

func (i *Ioregs) SetNR32(value uint8) {
	i.Regs[0x1C] = value
}

func (i *Ioregs) SetNR33(value uint8) {
	i.Regs[0x1D] = value
}

func (i *Ioregs) GetNR34() uint8 {
	return i.Regs[0x1E]
}

func (i *Ioregs) SetNR34(value uint8) {
	i.Regs[0x1E] = value
}

func (i *Ioregs) SetNR41(value uint8) {
	i.Regs[0x20] = value
}

func (i *Ioregs) GetNR42() uint8 {
	return i.Regs[0x21]
}

func (i *Ioregs) SetNR42(value uint8) {
	i.Regs[0x21] = value
}

func (i *Ioregs) GetNR43() uint8 {
	return i.Regs[0x22]
}

func (i *Ioregs) SetNR43(value uint8) {
	i.Regs[0x22] = value
}

func (i *Ioregs) GetNR44() uint8 {
	return i.Regs[0x23]
}

func (i *Ioregs) SetNR44(value uint8) {
	i.Regs[0x23] = value
}

func (i *Ioregs) GetNR50() uint8 {
	return i.Regs[0x24]
}

func (i *Ioregs) SetNR50(value uint8) {
	i.Regs[0x24] = value
}

func (i *Ioregs) GetNR51() uint8 {
	return i.Regs[0x25]
}

func (i *Ioregs) SetNR51(value uint8) {
	i.Regs[0x25] = value
}

func (i *Ioregs) GetNR52() uint8 {
	return i.Regs[0x26]
}

func (i *Ioregs) SetNR52(value uint8) {
	i.Regs[0x26] = value
}

func (i *Ioregs) GetWaveRAM(index uint8) uint8 {
	return i.Regs[0x30+index]
}

func (i *Ioregs) SetWaveRAM(index uint8, value uint8) {
	i.Regs[0x30+index] = value
}

func (i *Ioregs) GetLCDC() uint8 {
	return i.Regs[0x40]
}

func (i *Ioregs) SetLCDC(value uint8) {
	i.Regs[0x40] = value
}

func (i *Ioregs) GetSTAT() uint8 {
	return i.Regs[0x41]
}

func (i *Ioregs) SetSTAT(value uint8) {
	i.Regs[0x41] = value
}

func (i *Ioregs) GetSCY() uint8 {
	return i.Regs[0x42]
}

func (i *Ioregs) SetSCY(value uint8) {
	i.Regs[0x42] = value
}

func (i *Ioregs) GetSCX() uint8 {
	return i.Regs[0x43]
}

func (i *Ioregs) SetSCX(value uint8) {
	i.Regs[0x43] = value
}

func (i *Ioregs) GetLY() uint8 {
	return i.Regs[0x44]
}

func (i *Ioregs) GetLYC() uint8 {
	return i.Regs[0x45]
}

func (i *Ioregs) SetLYC(value uint8) {
	i.Regs[0x45] = value
}

func (i *Ioregs) GetDMA() uint8 {
	return i.Regs[0x46]
}

func (i *Ioregs) SetDMA(value uint8) {
	i.Regs[0x46] = value
}

func (i *Ioregs) GetBGP() uint8 {
	return i.Regs[0x47]
}

func (i *Ioregs) SetBGP(value uint8) {
	i.Regs[0x47] = value
}

func (i *Ioregs) GetOBP0() uint8 {
	return i.Regs[0x48]
}

func (i *Ioregs) SetOBP0(value uint8) {
	i.Regs[0x48] = value
}

func (i *Ioregs) GetOBP1() uint8 {
	return i.Regs[0x49]
}

func (i *Ioregs) SetOBP1(value uint8) {
	i.Regs[0x49] = value
}

func (i *Ioregs) GetWY() uint8 {
	return i.Regs[0x4A]
}

func (i *Ioregs) SetWY(value uint8) {
	i.Regs[0x4A] = value
}

func (i *Ioregs) GetWX() uint8 {
	return i.Regs[0x4B]
}

func (i *Ioregs) SetWX(value uint8) {
	i.Regs[0x4B] = value
}

func (i *Ioregs) GetKEY1() uint8 {
	return i.Regs[0x4D]
}

func (i *Ioregs) SetKEY1(value uint8) {
	i.Regs[0x4D] = value
}

func (i *Ioregs) GetVBK() uint8 {
	return i.Regs[0x4F]
}

func (i *Ioregs) SetVBK(value uint8) {
	i.Regs[0x4F] = value
}

func (i *Ioregs) SetHDMA1(value uint8) {
	i.Regs[0x51] = value
}

func (i *Ioregs) SetHDMA2(value uint8) {
	i.Regs[0x52] = value
}

func (i *Ioregs) SetHDMA3(value uint8) {
	i.Regs[0x53] = value
}

func (i *Ioregs) SetHDMA4(value uint8) {
	i.Regs[0x54] = value
}

func (i *Ioregs) GetHDMA5() uint8 {
	return i.Regs[0x55]
}

func (i *Ioregs) SetHDMA5(value uint8) {
	i.Regs[0x55] = value
}

func (i *Ioregs) GetRP() uint8 {
	return i.Regs[0x56]
}

func (i *Ioregs) SetRP(value uint8) {
	i.Regs[0x56] = value
}

func (i *Ioregs) GetBCPS() uint8 {
	return i.Regs[0x68]
}

func (i *Ioregs) SetBCPS(value uint8) {
	i.Regs[0x68] = value
}

func (i *Ioregs) GetBCPD() uint8 {
	return i.Regs[0x69]
}

func (i *Ioregs) SetBCPD(value uint8) {
	i.Regs[0x69] = value
}

func (i *Ioregs) GetOCPS() uint8 {
	return i.Regs[0x6A]
}

func (i *Ioregs) SetOCPS(value uint8) {
	i.Regs[0x6A] = value
}

func (i *Ioregs) GetOCPD() uint8 {
	return i.Regs[0x6B]
}

func (i *Ioregs) SetOCPD(value uint8) {
	i.Regs[0x6B] = value
}

func (i *Ioregs) GetOPRI() uint8 {
	return i.Regs[0x6C]
}

func (i *Ioregs) SetOPRI(value uint8) {
	i.Regs[0x6C] = value
}

func (i *Ioregs) GetSVBK() uint8 {
	return i.Regs[0x70]
}

func (i *Ioregs) SetSVBK(value uint8) {
	i.Regs[0x70] = value
}

func (i *Ioregs) GetPCM12() uint8 {
	return i.Regs[0x76]
}

func (i *Ioregs) GetPCM34() uint8 {
	return i.Regs[0x77]
}

func (i *Ioregs) GetIE() uint8 {
	return i.Regs[0x7F]
}

func (i *Ioregs) SetIE(value uint8) {
	i.Regs[0x7F] = value
}

// Dump

func (i *Ioregs) Dump() {
	i.dumpMemory(i.Regs[:], 0x00)
}

func (i *Ioregs) dumpMemory(memory []uint8, baseAddress uint16) {
	for idx := 0; idx < len(memory); idx += 16 {
		fmt.Printf("%04x ", baseAddress+uint16(idx))
		for j := 0; j < 16 && idx+j < len(memory); j++ {
			fmt.Printf("%02x ", memory[idx+j])
		}
		fmt.Println()
	}
}
