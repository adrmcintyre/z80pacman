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
		fmt.Fprintf(os.Stderr, "Could not open %s: %v\n", path, err)
		os.Exit(1)
	}
	copy(programROM[:], data)
}

func wireCPU() {
	z80.HookBusRead = busRead
	z80.HookBusWrite = busWrite
	z80.HookIoWrite = ioWrite

	// abort is a helper that aborts the program with the given message,
	// and displays the disassembly of the last few instructions executed.
	z80.HookAbort = func(msg string) {
		fmt.Printf("\n")
		dumpTraceLocs(16)
		fmt.Printf("last instruction:")
		for _, op := range z80.DebugTrace {
			fmt.Printf(" %02x", op)
		}
		panic(msg)
	}
}

func powerOn() {
	resetMachine()
	ioInit()
	vblankStart()
}

func runCPU() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
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

			select {
			case a := <-sig:
				switch a {
				case syscall.SIGINT:
					fmt.Println("\nCtrl+C")
				}
				os.Exit(1)
			default:
			}
		}
		os.Exit(1)
	}()
}

func resetMachine() {
	z80.Reset()
	resetDevices()
}

func resetDevices() {
	z80.ResetAssertPin.Store(false)
	z80.IrqAssertPin.Store(false)

	irqLowRegister.Store(0)
	watchdogReset()
}
