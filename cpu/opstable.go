package cpu

import (
	"fmt"
	"strings"
)

// An opTable defines a set of opcodes.
type opTable struct {
	fn  [256]func()     // each entry is a closure that executes the opcode at that index
	tpl [256]opTemplate // each entry is a template to help disassemble the opcode at that index
}

// An opTemplate contains templated text for disassembling an opcode.
type opTemplate struct {
	txt string // templated text for the opcode
	arg [2]int // argument values used when expanding txt
}

// newTable returns a new opTable, with all instructions initially declared illegal.
func newTable() *opTable {
	t := &opTable{}
	for i := range t.fn {
		t.fn[i] = illegal
		t.tpl[i] = opTemplate{txt: "illegal_opcode"}
	}
	return t
}

// def defines an opcode.
// <opcode> is the opcode;
// <fn> is the closure to run during execution;
// <txt> as a template used during disassembly
// <args> is 0, 1 or 2 arguments for the template.
func (t *opTable) def(opcode int, fn func(), txt string, args ...int) {
	tpl := opTemplate{txt: txt}
	copy(tpl.arg[:], args)
	t.fn[opcode] = fn
	t.tpl[opcode] = tpl
}

// text returns the disassembly for an opcode.
// <pc> is the program counter for the opcode, or the first prefix byte if any.
// <opcode> is the opcode to disassemble.
// <opargs> is any immediate bytes that follow the opcode.
// <hlAlias> is "hl", "ix", or "iy" according to whether a DD or FD prefix is present.
//
// The following placeholders are expanded from the template txt:
// %b - bit number (template arg)
// %c - condition code (template arg)
// %d - 16-bit register bc, de, hl/ix/iy, or sp (template arg)
// %e - relative address (oparg)
// %h - hl/ix/iy
// %m - interrupt mode 0, 1, or 2 (template arg)
// %n - immediate byte literal (oparg)
// %N - immediate word literal (oparg)
// %p - restart index (template arg)
// %q - 16-bit register bc, de, hl/ix/iy, or af (template arg)
// %r - 8-bit register b, c, d, e, h, l, (hl)/(ix+#d)/(iy+#d), or a
// .... (template arg, plus oparg for an index displacement)
func (t *opTable) text(pc uint16, opcode uint8, opargs []uint8, hlAlias string) string {
	res := ""
	k := 0
	tpl := t.tpl[opcode]
	j := 0
	pos := 0
	for {
		i := pos + strings.Index(tpl.txt[pos:], "%")
		if i < pos {
			return res + tpl.txt[pos:]
		}
		res += tpl.txt[pos:i]
		pos = i + 2
		switch tpl.txt[i+1] {
		case 'b':
			res += fmt.Sprintf("%d", tpl.arg[j])
			j++
		case 'c':
			conds := [8]string{"nz", "z", "nc", "c", "po", "pe", "p", "m"}
			res += conds[tpl.arg[j]]
			j++
		case 'd':
			regs := [4]string{"bc", "de", "hl", "sp"}
			dd := tpl.arg[j]
			j++
			if dd == 2 {
				res += hlAlias
			} else {
				res += regs[dd]
			}
		case 'e':
			// relative
			n := uint16(opargs[k])
			if n >= 0x80 {
				n -= 0x100
			}
			k++
			res += fmt.Sprintf("#%04x", pc+2+n)
		case 'h':
			res += hlAlias
		case 'm':
			// interrupt mode
			res += fmt.Sprintf("%d", tpl.arg[j])
			j++
		case 'n':
			n := opargs[k]
			k++
			if n < 10 {
				res += fmt.Sprintf("#%d", n)
			} else {
				res += fmt.Sprintf("#%02x", n)
			}
		case 'N':
			nn := word(opargs[k+1], opargs[k])
			k += 2
			res += fmt.Sprintf("#%04x", nn)
		case 'p':
			// reset
			res += fmt.Sprintf("#%02x", tpl.arg[j]*8)
			j++
		case 'q':
			regs := [4]string{"bc", "de", "hl", "af"}
			qq := tpl.arg[j]
			j++
			if qq == 2 {
				res += hlAlias
			} else {
				res += regs[qq]
			}
		case 'r':
			regs := [8]string{"b", "c", "d", "e", "h", "l", "(hl)", "a"}
			rrr := tpl.arg[j]
			j++
			if rrr == 6 && hlAlias != "hl" {
				offset := opargs[k]
				k++
				if offset < 10 {
					res += fmt.Sprintf("(%s+#%d)", hlAlias, offset)
				} else {
					res += fmt.Sprintf("(%s+#%02x)", hlAlias, offset)
				}
			} else {
				res += regs[rrr]
			}
		}
	}
}
