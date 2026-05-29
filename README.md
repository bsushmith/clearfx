# clearfx

`clearfx` is a Go CLI that clears the terminal with a short animated
transition. The default style is fire, with additional styles for moving ocean
waves, a Great Wave-inspired curl, matrix rain, glitch, starfield, black-hole,
meteor shower, lightning, page-burn, and typewriter-erase effects.

## Run

```sh
go run ./cmd/clearfx
```

Try a specific style:

```sh
go run ./cmd/clearfx --style ocean-wave
go run ./cmd/clearfx --style great-wave
go run ./cmd/clearfx --style matrix-rain
go run ./cmd/clearfx --style glitch
go run ./cmd/clearfx --style starfield
go run ./cmd/clearfx --style black-hole
go run ./cmd/clearfx --style meteor-shower
go run ./cmd/clearfx --style lightning
go run ./cmd/clearfx --style page-burn
go run ./cmd/clearfx --style typewriter-erase
go run ./cmd/clearfx --style random
go run ./cmd/clearfx --random-style
```

Experimental styles are kept behind a build tag while they are refined:

```sh
go run -tags experimental ./cmd/clearfx --style dune-worm
go run -tags experimental ./cmd/clearfx --style sandstorm
go run -tags experimental ./cmd/clearfx --style snowfall
go run -tags experimental ./cmd/clearfx --style ink-drop
go run -tags experimental ./cmd/clearfx --style shatter
go run -tags experimental ./cmd/clearfx --style aurora
go run -tags experimental ./cmd/clearfx --style laser-grid
go run -tags experimental ./cmd/clearfx --style rainstorm
```

List styles:

```sh
go run ./cmd/clearfx --list-styles
```

Install locally:

```sh
go install ./cmd/clearfx
```

Then run:

```sh
clearfx --style fire
```

## Flags

```text
--duration 700ms
--fps 30
--style fire
--intensity medium
--palette classic
--no-animation
--list-styles
--random-style
--force-ansi
--version
```

Environment defaults:

```sh
CLEARFX_STYLE=ocean-wave
CLEARFX_DURATION=1200ms
CLEARFX_FPS=45
CLEARFX_INTENSITY=medium
CLEARFX_PALETTE=classic
```

`--list-styles` prints each style with category, recommended duration, and
recommended FPS. If `--duration` or `--fps` is not set explicitly, `clearfx`
uses the selected style's recommendation.

## Benchmarks

Measured on macOS `darwin/arm64`, Apple M4, with `120x40` frames and medium
intensity. The end-to-end benchmark generates each animation frame and renders
it through the terminal renderer to `io.Discard`, so these numbers measure
Go-side frame/render cost rather than terminal emulator GPU/CPU cost.

Run the benchmarks:

```sh
GOCACHE=$PWD/.go-cache go test -bench=. -benchmem ./internal/animation ./internal/terminal
```

Latest measured end-to-end results:

| Style | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| lightning | 9,877 | 41,022 | 1 |
| meteor-shower | 10,027 | 41,027 | 1 |
| matrix-rain | 12,805 | 41,002 | 1 |
| starfield | 18,254 | 41,028 | 1 |
| typewriter-erase | 20,522 | 41,064 | 1 |
| great-wave | 35,820 | 41,075 | 1 |
| ocean-wave | 41,512 | 41,070 | 1 |
| black-hole | 52,071 | 41,042 | 1 |
| glitch | 52,807 | 41,633 | 1 |
| page-burn | 53,279 | 41,024 | 1 |
| fire | 61,163 | 41,058 | 1 |

The shared terminal renderer alone measured `16,816 ns/op`, `1 B/op`, and
`0 allocs/op`. At `30 FPS`, each frame has about `33.3ms` available, so the
slowest measured path is comfortably under the frame budget before real terminal
emulator rendering overhead.

## Development

```sh
go test ./...
go build ./...
```
