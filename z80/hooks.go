package z80

var OnBusRead = func(uint16) uint8 { panic("BusRead not hooked") }
var OnBusWrite = func(uint16, uint8) { panic("BusWrite not hooked") }
var OnIoWrite = func(uint16, uint8) { panic("IoWrite not hooked") }
var OnAbort = func(msg string) { panic(msg) }

// illegal panics with an "illegal" message when executed.
func illegal() {
	OnAbort("illegal instruction")
}

// TODO aborts with a "TODO" message when executed.
func TODO() {
	OnAbort("TODO")
}

// unimplemented aborts with an "unimplemented" message when executed.
func unimplemented() {
	OnAbort("unimplemented instruction")
}
