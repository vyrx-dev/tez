<div align="center">

# Tez

A typing game with a stretchy blob.



https://github.com/user-attachments/assets/a63bb5c8-beb8-4719-b9a8-664ddb1c8bf6


</div>

## The blob

This is the part I like. There are just two dots:

- the **head**, which runs at the word you're typing
- the **tail**, which runs at the head, just slower

The head is quick so it races ahead. The tail can't keep up, so it trails
behind. The gap between them is what you see as the long stretchy bit. Move
fast and it stretches, stop and it catches up again. That's the whole trick,
and it's about ten lines in `logic.go`.

Two small things worth knowing:

- **It looks the same on any monitor.** Both dots ease toward their target
  with `1 - exp(-speed * dt)`. The simpler `dt * speed` that most people write
  first drifts: on a 144Hz screen the blob moves slower in real seconds, and
  on a slow frame it jumps past its target. The exponential one can never
  overshoot.
- **The stretch is capped.** The dome can only get so long, because the blob's
  resting spot teleports to whatever word it just ate, so uncapped it would
  smear the whole screen. It stays as wide as the blob the whole stretch, so
  it reads as one fat blob being pulled rather than a rubber band.

## Installing

The release binaries are not fully self-contained, because this is a graphical
game. On Linux they need, in addition to the binary itself:

- `libffi` (the library raylib-go loads raylib through)
- `libx11`, `wayland`, `mesa` (the window and OpenGL)
- a running display, so not over SSH without X forwarding
- PulseAudio or ALSA, otherwise the game runs silently

On Arch those are normally already installed. On Windows and macOS the binary
brings its own raylib and needs nothing extra.

The sound effects are embedded in the binary, so there is no data directory to
install and the game never writes to disk.

### Why the game might be silent

raylib picks its audio backend at runtime. If it cannot find PulseAudio or ALSA
it falls back to a backend called `Null`, which accepts the audio and plays
nothing, with no error message. The startup log tells you which one you got:

```
INFO:     > Backend:       miniaudio | PulseAudio    <- real output
INFO:     > Backend:       miniaudio | Null          <- silent, no error
```

So if the log says `Null`, the sounds are fine but nothing is being played.
Install PulseAudio or ALSA, or run the binary from inside the nix shell, which
puts those libraries on the search path.

## Building

Easiest is the nix shell, which now lives inside this folder:

```sh
nix-shell --run "go run ."
```

If you don't have raylib installed, you can skip C entirely. This bundles a
prebuilt libraylib and needs no compiler:

```sh
CGO_ENABLED=0 go run .
```

To build a binary you can run later, use `go build` instead of `go run`:

```sh
nix-shell --run "go build -o tez ." && ./tez
```

Run `./tez` from inside the nix shell, or use the `play.sh` wrapper, which
borrows the shell's environment. Outside either one, raylib cannot find
PulseAudio or ALSA and falls back to a silent `Null` audio backend.

The blob maths doesn't need raylib or a window, so it's tested on its own:

```sh
go test -tags test ./...
```

## What's where

| File | What it does |
| --- | --- |
| `main.go` | Opens the window, audio and the frame loop. Nothing else. |
| `game.go` | Game state, restarting, the boot and game over screens. |
| `update.go` | What happens each frame: keys, spawning, the blob. |
| `render.go` | Everything that gets drawn. |
| `logic.go` | The blob maths. No raylib in here at all. |
| `logic_test.go` | Tests for the maths. |
| `audio.go` | Loads the sound effects, which are embedded in the binary. |

Want more words? They all live in one list, `wordList` in `types.go`, split by
difficulty. Add or take away whatever you like there and the game picks it up
next build.

The sound effects are embedded in the binary from `assets/audio`, so the game
never writes anything to disk and makes no noise about it if it cannot.

> **Heads up:** I'm not a game developer, so there are probably things done
> wrong or clumsy in here. Be gentle, ha ha.

## TODO

- [ ] **Death mode**: one wrong letter and it's over, no hearts and no second
  chances
- [ ] more words
- [ ] difficulty that isn't just "everything got faster"
- [ ] a combo score for typing words without a slip

## Contributing

There are comments on the bits that are easy to get wrong: the blob chase in
`logic.go`, how the game moves between states in `update.go`, and why the
sounds are generated in `audio.go`. If you spot a mistake, or see something
that could be tidier, simpler or faster, open a PR. The blob is the fun part,
so ideas for it are very welcome.

## Credit

The stretchy blob is my Go version of [rexim's blob](https://gist.github.com/rexim/c1e3c1b99cb824ab6c817c3f45d75564)
(Rust + SDL3), which is where the two dots idea comes from.

## License

[MIT](LICENSE)
