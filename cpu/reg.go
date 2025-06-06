package cpu

type Reg16 int
type Reg8 int

const (
	REG_AF Reg16 = iota
	REG_BC
	REG_DE
	REG_HL
	REG_SP
	REG_PC
)

const (
	REG_B Reg8 = iota
	REG_C
	REG_D
	REG_E
	REG_H
	REG_L
	REG_MEM_HL
	REG_A
)

var reg16Name = map[Reg16]string{
	REG_AF: "Register AF",
	REG_BC: "Register BC",
	REG_DE: "Register DE",
	REG_HL: "Register HL",
	REG_SP: "Register SP",
	REG_PC: "Register PC",
}

func (ss Reg16) String() string {
	return reg16Name[ss]
}
