package main

import (
	"fmt"
	"strings"
)

type Table struct {
	fn  [256]func()
	tpl [256]opTemplate
}

type opTemplate struct {
	txt string
	arg [2]int
}

func newTable() *Table {
	t := &Table{}
	for i := range t.fn {
		t.fn[i] = illegal
	}
	return t
}

func (t *Table) def(opcode int, fn func(), txt string, args ...int) {
	tpl := opTemplate{txt: txt}
	copy(tpl.arg[:], args)
	t.fn[opcode] = fn
	t.tpl[opcode] = tpl
}

func (t *Table) text(pc uint16, opcode uint8, opargs []uint8, hlAlias string) string {
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
			res += fmt.Sprintf("$%04x", pc+2+n)
		case 'h':
			res += hlAlias
		case 'm':
			// interrupt mode
			res += fmt.Sprintf("%d", tpl.arg[j])
			j++
		case 'n':
			n := opargs[k]
			k++
			res += fmt.Sprintf("$%02x", n)
		case 'N':
			nn := word(opargs[k+1], opargs[k])
			k += 2
			res += fmt.Sprintf("$%04x", nn)
		case 'p':
			// reset
			res += fmt.Sprintf("$%d", tpl.arg[j]*8)
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
				res += fmt.Sprintf("(%s+$%02x)", hlAlias, offset)
			} else {
				res += regs[rrr]
			}
		}
	}
}
