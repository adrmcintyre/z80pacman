package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adrmcintyre/z80pacman/audio"
	"github.com/adrmcintyre/z80pacman/video"
	"github.com/adrmcintyre/z80pacman/z80"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	// dip switches
	flagCoins     = flag.Int("dip-coins", 1, "coins per play: 0, 1, 2; or 3 is 2 plays per coin")
	flagLives     = flag.Int("dip-lives", 3, "lives: 1, 2, 3, 5")
	flagBonus     = flag.Int("dip-bonus", 10_000, "bonus: 0, 10000, 15000, 20000")
	flagDifficult = flag.Bool("dip-hard", false, "increase difficulty (faster ghosts)")
	flagAltGhosts = flag.Bool("dip-alternate", false, "alternate ghost names")
	flagTest      = flag.Bool("dip-test", false, "test mode")
	flagRackTest  = flag.Bool("dip-rack", false, "rack test mode")
	flagCocktail  = flag.Bool("dip-cocktail", false, "cocktail mode")

	// debug flags
	flagTrace      = flag.Bool("trace", false, "display instruction trace")
	flagTraceMem   = flag.String("trace-mem", "", "display memory r/w trace; e.g. 4400-4bff,5040")
	flagNoWatchdog = flag.Bool("no-watchdog", false, "disable watchdog reset")
	flagDebugTest  = flag.String("debug-test", "", "load specified test program")
)

func main() {
	parseFlags()
	machineInit()
	defer audio.Shutdown()

	powerOn()
	runCPU()
	ebiten.RunGame(&Game{})
}

func parseFlags() {
	flag.Parse()
	ioParseFlags()
	debugParseFlags()
}

func machineInit() {
	programInit()
	videoInit()
	audioInit()
	wireCPU()
}

func programInit() {
	if *flagDebugTest != "" {
		loadTestProgram(*flagDebugTest)
	} else {
		romSet{
			"roms/pacman.6e": programROM[0x0000:0x1000],
			"roms/pacman.6f": programROM[0x1000:0x2000],
			"roms/pacman.6h": programROM[0x2000:0x3000],
			"roms/pacman.6j": programROM[0x3000:0x4000],
		}.Load()
	}
}

func videoInit() {
	var (
		paletteROM [0x100]uint8
		colorROM   [0x20]uint8
		tileROM    [0x1000]uint8
		spriteROM  [0x1000]uint8
	)
	romSet{
		"roms/82s126.4a": paletteROM[:],
		"roms/82s123.7f": colorROM[:],
		"roms/pacman.5e": tileROM[:],
		"roms/pacman.5f": spriteROM[:],
	}.Load()
	video.Init(paletteROM[:], colorROM[:], tileROM[:], spriteROM[:])
}

func audioInit() {
	var (
		waveROM [0x100]uint8
	)
	romSet{
		"roms/82s126.1m": waveROM[:],
	}.Load()
	if err := audio.Init(waveROM[:]); err != nil {
		die("initialising audio: %v", err)
	}
}

func wireCPU() {
	z80.OnBusRead = busRead
	z80.OnBusWrite = busWrite
	z80.OnIoWrite = ioWrite
	z80.OnAbort = cpuAbort
}

func powerOn() {
	fmt.Println("[POWER ON]")
	resetMachine()
	ioInit()
	vblankStart()
}

var (
	cpuResetCh = make(chan struct{}, 1)
	cpuNmiCh   = make(chan struct{}, 1)
	cpuIrqCh   = make(chan uint8, 1)
)

func runCPU() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-cpuResetCh:
				z80.Reset()
			case <-cpuNmiCh:
				z80.NMI()
			case irqData := <-cpuIrqCh:
				z80.IRQ(irqData)
			case s := <-sigCh:
				switch s {
				case syscall.SIGINT:
					die("\nCtrl+C")
				default:
					die("\nQUIT")
				}
			default:
			}

			pc := z80.Step()

			if breakpointed(pc) {
				break
			}
			debugTraceCPU(pc)
		}
		// TODO - kill ebiten and exit cleanly
		die("terminated by breakpoint")
	}()
}

func resetMachine() {
	fmt.Println("[RESET]")
	cpuResetCh <- struct{}{}
	resetDevices()
}

func resetDevices() {
	irqLowRegister.Store(0)
	watchdogReset()
}

// cpuAbort aborts the program with the given message, and displays
//
//	the disassembly of the last few instructions executed.
func cpuAbort(msg string) {
	fmt.Println()
	dumpTraceLocs(16)
	fmt.Printf("last instruction:")
	for _, op := range z80.DebugTrace {
		fmt.Printf(" %02x", op)
	}
	die("abort: %s", msg)
}

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
