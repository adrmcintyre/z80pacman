package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adrmcintyre/z80/audio"
	"github.com/adrmcintyre/z80/video"
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

func setWindowSize(aspectRatio float64, fillRatio float64) {
	w, h := ebiten.Monitor().Size()
	fw, fh := float64(w), float64(h)
	if fw/fh > aspectRatio {
		w = int(fh * aspectRatio)
	} else {
		h = int(fw / aspectRatio)
	}
	ebiten.SetWindowSize(int(float64(w)*fillRatio), int(float64(h)*fillRatio))
}

func main() {
	flag.Parse()
	debugInit()
	switch *flagCoins {
	case 0, 1, 2, 3:
	default:
		fmt.Printf("illegal -dip-coins")
	}
	switch *flagLives {
	case 1, 2, 3, 5:
	default:
		fmt.Printf("illegal -dip-lives")
	}
	switch *flagBonus {
	case 0, 10_000, 15_000, 20_000:
	default:
		fmt.Printf("illegal -dip-bonus")
	}

	ebiten.SetWindowTitle("ebiman")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	setWindowSize(28.0/36.0, 0.75)
	video.InitColors()
	video.InitTiles()
	video.InitSprites()

	au := audio.NewAudio()
	defer func() { _ = au.Close() }()

	// connect to host's audio
	if err := au.Connect(audio.LatencyLow); err != nil {
		panic(err)
	}

	runMachine()

	g := &Game{}
	ebiten.RunGame(g)
}

func runMachine() {
	if *flagDebugTest != "" {
		loadTestProgram(*flagDebugTest)
		startVblankTicker()
	} else {
		load("pacman.rom")
		startVblankTicker()
	}

	powerOn()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for stepCPU() {
			select {
			case a := <-sig:
				switch a {
				case syscall.SIGINT:
					fmt.Println("\nCtrl+C")
					os.Exit(1)
				}
			default:
			}
		}
		os.Exit(1)
	}()
}

func powerOn() {
	resetMachine()
	ioInit()
}

func resetMachine() {
	resetCPU()
	resetDevices()
}

func resetDevices() {
	resetAssertPin.Store(false)
	irqAssertPin.Store(false)

	irqLowRegister.Store(0)
	watchdogRegister.Store(15 << 28)
}

func load(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s: %v\n", path, err)
		os.Exit(1)
	}
	copy(programROM[:], data)

	setBreakpoint(0x0038, Breakpoint{
		dumpTiles:   false,
		dumpProgram: false,
		dumpTrace:   1000,
	})
}

type Game struct{}

func (g *Game) Update() error {
	put := func(bits uint8, bit uint8, v bool) uint8 {
		if v {
			bits |= bit
		} else {
			bits &= ^bit
		}
		return bits
	}

	var in0 uint8
	in0 = put(in0, In0_NotUp, !ebiten.IsKeyPressed(ebiten.KeyUp))
	in0 = put(in0, In0_NotLeft, !ebiten.IsKeyPressed(ebiten.KeyLeft))
	in0 = put(in0, In0_NotRight, !ebiten.IsKeyPressed(ebiten.KeyRight))
	in0 = put(in0, In0_NotDown, !ebiten.IsKeyPressed(ebiten.KeyDown))
	in0 = put(in0, In0_NotRackTest, !*flagRackTest)
	in0 = put(in0, In0_RisingCoin1, !false)
	in0 = put(in0, In0_RisingCoin2, !false)
	in0 = put(in0, In0_NotCredit, !ebiten.IsKeyPressed(ebiten.KeyC))
	in0_State.Store(uint32(in0))

	var in1 uint8
	in0 = put(in1, In1_NotUp, !false)
	in1 = put(in1, In1_NotLeft, !false)
	in1 = put(in1, In1_NotRight, !false)
	in1 = put(in1, In1_NotDown, !false)
	in1 = put(in1, In1_NotTest, (!*flagTest) != ebiten.IsKeyPressed(ebiten.KeyT))
	in1 = put(in1, In1_NotStart1, !ebiten.IsKeyPressed(ebiten.Key1))
	in1 = put(in1, In1_NotStart2, !ebiten.IsKeyPressed(ebiten.Key2))
	in1 = put(in1, In1_NotCocktailMode, !*flagCocktail)
	in1_State.Store(uint32(in1))

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	video.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	const (
		hTiles, vTiles        = 28, 36 // dimensions of display area in tiles
		tileWidth, tileHeight = 8, 8   // dimensions of a tile in simulated pixels
		border                = 8      // small border around display in simulated pixels

		// calculate logical output size
		logicalWidth  = float64(hTiles*tileWidth + 2*border)
		logicalHeight = float64(vTiles*tileHeight + 2*border)
		logicalAspect = float64(logicalWidth) / float64(logicalHeight)
	)

	var (
		fOutsideWidth  = float64(outsideWidth)
		fOutsideHeight = float64(outsideHeight)
		outputAspect   = fOutsideWidth / fOutsideHeight

		fScreenWidth  = logicalWidth
		fScreenHeight = logicalHeight
	)

	// centre output in window
	if outputAspect > logicalAspect {
		fScreenWidth = logicalHeight * outputAspect
	} else {
		fScreenHeight = logicalWidth / outputAspect
	}
	return int(fScreenWidth), int(fScreenHeight)
}
