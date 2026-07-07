package cpu

type State struct {
	B, C, D, E, H, L, A     uint8  // 8 bit regs
	IX, IY, PC, SP          uint16 // 16 bit regs
	BC2, DE2, HL2, AF2      uint16 // alternate regs
	FS, FZ, FH, FPV, FN, FC bool   // flags
}

func GetState() State {
	return State{
		B:   b.Rd(),
		C:   c.Rd(),
		D:   d.Rd(),
		E:   e.Rd(),
		H:   h.Rd(),
		L:   l.Rd(),
		A:   a.Rd(),
		IX:  ix.Rd16(),
		IY:  iy.Rd16(),
		PC:  pc.Rd16(),
		SP:  sp.Rd16(),
		BC2: bc2.Rd16(),
		DE2: de2.Rd16(),
		HL2: hl2.Rd16(),
		AF2: af2.Rd16(),
		FS:  flagS.get(),
		FZ:  flagZ.get(),
		FH:  flagH.get(),
		FPV: flagPV.get(),
		FN:  flagN.get(),
		FC:  flagC.get(),
	}
}

func Disassemble(pc uint16, trace []uint8) string {
	switch trace[0] {
	case 0xdd, 0xfd:
		hlAlias := "hl"
		if trace[0] == 0xdd {
			hlAlias = "ix"
		} else {
			hlAlias = "iy"
		}
		if trace[1] == 0xcb {
			return bitTable.text(pc, trace[3], trace[2:], hlAlias)
		}
		return coreTable.text(pc, trace[1], trace[2:], hlAlias)

	case 0xed:
		return miscTable.text(pc, trace[1], trace[2:], "hl")
	case 0xcb:
		return bitTable.text(pc, trace[1], trace[2:], "hl")
	default:
		return coreTable.text(pc, trace[0], trace[1:], "hl")
	}
}
