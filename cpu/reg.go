package cpu

type Reg16 int

const (
	REG_AF Reg16 = iota
	REG_BC
	REG_DE
	REG_HL
	REG_SP
	REG_PC
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
