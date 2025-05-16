package cpu

type Reg8 int
type Reg16 int

const (
	REG_A Reg8 = iota
	REG_F
	REG_B
	REG_C
	REG_D
	REG_E
	REG_H
	REG_L
)

const (
	REG_AF Reg16 = iota
	REG_BC
	REG_DE
	REG_HL
	REG_SP
	REG_PC
)

var reg8Name = map[Reg8]string{
	REG_A: "Register A",
	REG_F: "Register F",
	REG_B: "Register B",
	REG_C: "Register C",
	REG_D: "Register D",
	REG_E: "Register E",
	REG_H: "Register H",
	REG_L: "Register L",
}

var reg16Name = map[Reg16]string{
	REG_AF: "Register AF",
	REG_BC: "Register BC",
	REG_DE: "Register DE",
	REG_HL: "Register HL",
	REG_SP: "Register SP",
	REG_PC: "Register PC",
}

func (ss Reg8) String() string {
	return reg8Name[ss]
}

func (ss Reg16) String() string {
	return reg16Name[ss]
}
