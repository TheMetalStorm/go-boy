package rom

import (
	"bufio"
	"fmt"
	"os"
)

type Rom struct {
	data []byte
}

func (r *Rom) GetLength() int {
	return len(r.data)
}

func NewRom(path string) *Rom {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	rom := &Rom{}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		rom.data = append(rom.data, b)
	}
	return rom
}

func (r *Rom) Dump() {

	for idx := 0; idx < len(r.data); idx += 16 {
		fmt.Printf("%04x ", uint16(idx))
		for j := 0; j < 16 && idx+j < len(r.data); j++ {
			fmt.Printf("%02x ", r.data[idx+j])
		}
		fmt.Println()
	}

}

// returns data and instructions to increment PC by
func (r *Rom) ReadByteAt(address uint16) (data byte, numReadBytes uint16) {
	return r.data[address], 1
}
