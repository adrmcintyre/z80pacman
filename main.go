package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	flagDelayHack  = flag.Bool("delay-hack", false, "simulate pause during pacman delay loop")
	flagTrace      = flag.Bool("trace", false, "display instruction trace")
	flagTraceMem   = flag.String("trace-mem", "", "display memory r/w trace; e.g. 4400-4bff,5040")
	flagNoWatchdog = flag.Bool("no-watchdog", false, "disable watchdog reset")
	flagDebugTest  = flag.String("debug-test", "", "load specified test program")
)

func main() {
	parseFlags()

	video.Init()

	err := audio.Init()
	defer audio.Shutdown()
	if err != nil {
		panic(err)
	}

	programInit()
	wireCPU()
	powerOn()
	runCPU()
	ebiten.RunGame(&Game{})
}

func parseFlags() {
	flag.Parse()
	ioParseFlags()
	debugParseFlags()
}

func programInit() {
	if *flagDebugTest != "" {
		loadTestProgram(*flagDebugTest)
	} else {
		loadProgram("pacman.rom")
	}
}

func loadProgram(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		die("Could not open %s: %v", path, err)
	}
	copy(programROM[:], data)
}

func wireCPU() {
	z80.OnBusRead = busRead
	z80.OnBusWrite = busWrite
	z80.OnIoWrite = ioWrite
	z80.OnAbort = cpuAbort
}

func powerOn() {
	resetMachine()
	ioInit()
	vblankStart()
}

var (
	irqCh   = make(chan uint8)
	nmiCh   = make(chan struct{})
	resetCh = make(chan struct{})
)

func runCPU() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-resetCh:
				resetMachine()
			case data := <-irqCh:
				z80.IRQ(data)
			case <-nmiCh:
				z80.NMI()
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

			if *flagDelayHack {
				if pc == 0x32ed {
					time.Sleep(100 * time.Millisecond)
				}
			}

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
	z80.Reset()
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
