package cpu

var HookBusRead = func(uint16) uint8 { panic("BusRead not hooked") }
var HookBusWrite = func(uint16, uint8) { panic("BusWrite not hooked") }
var HookIoWrite = func() { panic("IoWrite not hooked") }
var HookAbort = func(msg string) { panic(msg) }

// illegal panics with an "illegal" message when executed.
func illegal() {
	HookAbort("illegal instruction")
}

// TODO aborts with a "TODO" message when executed.
func TODO() {
	HookAbort("TODO")
}

// unimplemented aborts with an "unimplemented" message when executed.
func unimplemented() {
	HookAbort("unimplemented instruction")
}
