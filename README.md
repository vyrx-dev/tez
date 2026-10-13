<div align="center">

# Tez

A typing game with a stretchy blob.

https://github.com/user-attachments/assets/a63bb5c8-beb8-4719-b9a8-664ddb1c8bf6

<br>

</div>

> **Heads up:** I'm not a game developer, so there are probably things done
> wrong or clumsy in here. Be gentle, ha ha — always open to learning!

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
  with `1 - exp(-speed * dt)`. The simpler `dt * speed` step (which I first tested
  in [blob-physics](https://github.com/vyrx-dev/blob-physics)) drifts: on high refresh
  screens the blob moves at different speeds, and on a dropped frame it can jump past its target.
  The exponential form never overshoots.
- **The stretch is capped.** The dome can only get so long, because the blob's
  resting spot teleports to whatever word it just ate, so uncapped it would
  smear the whole screen. It stays as wide as the blob the whole stretch, so
  it reads as one fat blob being pulled rather than a rubber band.

## Installation

### Arch Linux (AUR)
```sh
yay -S tez-bin
```

### Nix
```sh
nix run github:vyrx-dev/tez
```

### Prebuilt binaries
Download the latest version for your platform from [Releases](https://github.com/vyrx-dev/tez/releases):
* **Linux**: `.tar.gz`, `.deb` (Debian/Ubuntu), or `.rpm` (Fedora/RHEL)
* **macOS**: `tez_*_darwin_all.tar.gz` (Universal)
* **Windows**: `tez_*_windows_amd64.zip` (Standalone `tez.exe`)

## Building from source

You just need Go 1.24+. No C compiler or raylib setup required:

```sh
CGO_ENABLED=0 go run .
```

To build a binary:
```sh
CGO_ENABLED=0 go build -o tez .
```

### Nix
```sh
nix run .
# or in the dev shell:
nix-shell --run "go run ."
```

### Tests
The blob maths doesn't need raylib or a window, so it's tested on its own:
```sh
go test -tags test ./...
```

### Audio troubleshooting (Linux)
raylib picks its audio backend at runtime. If it cannot find PulseAudio or ALSA, it silently falls back to `Null` (plays nothing, with no error). Check your startup terminal log:

```
INFO:     > Backend:       miniaudio | PulseAudio    <- sound works
INFO:     > Backend:       miniaudio | Null          <- silent fallback
```

If you see `Null`:
* **Standard Linux**: Make sure `alsa-lib` (Arch/Fedora) or `libasound2` (Ubuntu/Debian) is installed.
* **NixOS**: Run with `steam-run ./tez`, `nix-shell --run "./tez"`, or `nix run .`.

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

## TODO

- [ ] **Death mode**: one wrong letter and it's over, no hearts and no second chances
- [ ] more words
- [ ] difficulty that isn't just "everything got faster"
- [ ] a combo score for typing words without a slip

## Credit

The stretchy blob is inspired by [Tsoding's blob](https://gist.github.com/rexim/c1e3c1b99cb824ab6c817c3f45d75564) (Rust + SDL3). I first prototyped and verified the mechanics in Go as [blob-physics](https://github.com/vyrx-dev/blob-physics) before building this game on top of it.

## License

[MIT](LICENSE)
