package cpu

type Reg int

const (
	REG_A Reg = iota
	REG_F
	REG_B
	REG_C
	REG_D
	REG_E
	REG_H
	REG_L

	REG_AF
	REG_BC
	REG_DE
	REG_HL
	REG_SP
	REG_PC
)

var regName = map[Reg]string{
	REG_A:  "Register A",
	REG_F:  "Register F",
	REG_B:  "Register B",
	REG_C:  "Register C",
	REG_D:  "Register D",
	REG_E:  "Register E",
	REG_H:  "Register H",
	REG_L:  "Register L",
	REG_AF: "Register AF",
	REG_BC: "Register BC",
	REG_DE: "Register DE",
	REG_HL: "Register HL",
	REG_SP: "Register SP",
	REG_PC: "Register PC",
}

func (ss Reg) String() string {
	return regName[ss]
}
