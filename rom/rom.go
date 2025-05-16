package rom

import (
	"bufio"
	"fmt"
	"os"
)

type Rom struct {
	data []byte
}

func NewRom(path string) *Rom {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	// Create a new reader
	reader := bufio.NewReader(file)

	rom := &Rom{}
	// Read the file byte by byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		rom.data = append(rom.data, b)
	}
	return rom
}

func (r *Rom) ReadByte(address uint16) (byte, uint16) {
	return r.data[address], 1
}

func (r *Rom) Read16(address uint16) (uint16, uint16) {
	a1 := uint16(r.data[address])
	a2 := uint16(r.data[address+1])

	return uint16(a1 | a2<<8), 2
}
