# z80pacman

This is an emulator for Namco's pacman, written in go, utilising the ebiten gaming library for I/O, and running the original z80 code in an emulated processor. Please see [z80/README.md](z80/README.md) for more
information on the z80 emulator.

## Running

Legally obtain a Midway Pacman ROM set and place them in the roms/ directory, see [roms/README.md](roms/README.md).

```
go build && ./z80pacman
```

> [!NOTE]
> Do not worry if you see an error like `[CAMetalLayer nextDrawable] returning nil because allocation failed.` under OSX.
> This is a known issue with the drivers used by ebiten and does not indicate an underlying problem.

## Options

### Dip Switches
The following options configure virtual dip-switches, as used in the original hardware:
* `-dip-alternate` use alternate ghost names.
* `-dip-bonus` score at which the one and only bonus life is awarded: 0 (none), 10000 (default), 15000, 20000.
* `-dip-cocktail` enables cocktail mode: inverts P2's screen.
* `-dip-coins` number of coins per credit:
    * 0: free play
    * 1: 1 coin per credit (default)
    * 2: 2 coins per credit
    * 3: 2 credits per coin
* `-dip-hard` increases difficulty: faster ghosts.
* `-dip-lives` sets how many lives pacman starts with: 1, 2, 3 (default), 5.
* `-dip-rack` enables rack-test mode: cycles through every screen once a game is started.
* `-dip-test` enables test mode (memory test, displays configuration, sound hardware triggered by player inputs).

### Debugging
These options can be used for debugging the emulator itself:

* `-no-watchdog` disables watchdog reset.
* `-debug-test` loads specified built in test program.
* `-trace` displays a trace of executed instructions.
* `-trace-mem` displays a memory r/w trace: specify a comma separate list of address or address-ranges, e.g. 4400-4bff,5040

# Implementation Notes

This is a free-running emulator - no attempt is made to match the rate of emulated instruction execution to the correct clock frequency. We can get away with this for pacman, as the game is ticked at about 60Hz by an interrupt issued before every display refresh (VBLANK), which we do emulate. We just get through each frame's work a lot faster than on real hardware.

The sound hardware is emulated by generating 1/60s of audio on every VBLANK and writing it into a buffer, which is then consumed by the ebiten audio processing thread approx every 1/20s. My first attempt only generated the audio when ebiten requested it, which resulted in poor quality audio as the emulated registers are updated every frame for certain sound effects, and we end up missing around 75% of them.
