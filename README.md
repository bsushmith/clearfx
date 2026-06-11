# clearfx

`clearfx` is a Go CLI that clears the terminal with a short animated
transition. It includes multiple built-in effects, configurable palettes,
structured style discovery, JSON config presets, and a `run` mode that can
animate before executing another command.

## Run

```sh
go run ./cmd/clearfx
```

Try a specific style and palette:

```sh
go run ./cmd/clearfx --style ocean-wave --palette ocean
go run ./cmd/clearfx --style matrix-rain --palette matrix
go run ./cmd/clearfx --style glitch --palette aurora
go run ./cmd/clearfx --style random
go run ./cmd/clearfx --random-style
```

Run a command after the transition:

```sh
go run ./cmd/clearfx run --style great-wave --palette ocean -- printf 'ready\n'
```

Browse styles interactively:

```sh
go run ./cmd/clearfx preview --palette ocean
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

## Discovery

List styles:

```sh
go run ./cmd/clearfx --list-styles
go run ./cmd/clearfx --list-styles --json
```

List palettes:

```sh
go run ./cmd/clearfx --list-palettes
go run ./cmd/clearfx --list-palettes --json
```

List presets and inspect resolved config:

```sh
go run ./cmd/clearfx --list-presets
go run ./cmd/clearfx --list-presets --json
go run ./cmd/clearfx --show-config
go run ./cmd/clearfx --show-config --json
```

Install locally:

```sh
go install ./cmd/clearfx
```

Then run:

```sh
clearfx --style fire --palette ember
```

## Flags

```text
--duration 700ms
--fps 30
--style fire
--intensity medium
--palette classic
--preset ""
--config ~/.config/clearfx/config.json
--no-animation
--list-styles
--list-palettes
--list-presets
--show-config
--json
--random-style
--force-ansi
--version
```

`clearfx run` accepts the same flags as the base command and then requires a
command after `--`.

`clearfx preview` runs a line-oriented interactive gallery. It replays the
current style and lets you use `n`, `p`, `r`, `q`, or `Enter` to browse,
replay, quit, or accept the current style.

## Config

Config is loaded from `~/.config/clearfx/config.json` by default, or from
`CLEARFX_CONFIG` / `--config`.

Example:

```json
{
  "style": "ocean-wave",
  "palette": "ocean",
  "intensity": "medium",
  "preset": "cinematic",
  "presets": {
    "cinematic": {
      "style": "great-wave",
      "duration": "1400ms",
      "fps": 45,
      "palette": "ocean"
    },
    "fast": {
      "style": "glitch",
      "duration": "450ms",
      "fps": 45,
      "palette": "aurora"
    }
  }
}
```

Precedence is:

```text
flags > environment > config file > built-in defaults
```

Environment defaults:

```sh
CLEARFX_STYLE=ocean-wave
CLEARFX_DURATION=1200ms
CLEARFX_FPS=45
CLEARFX_INTENSITY=medium
CLEARFX_PALETTE=classic
CLEARFX_PRESET=cinematic
```

If `--duration` or `--fps` is not set explicitly, `clearfx` uses the selected
style's recommendation.

## Benchmarks

Measured on macOS `darwin/arm64`, Apple M4, with `120x40` frames and medium
intensity. The end-to-end benchmark generates each animation frame and renders
it through the terminal renderer to `io.Discard`, so these numbers measure
Go-side frame/render cost rather than terminal emulator GPU/CPU cost.

Run the benchmarks:

```sh
GOCACHE=$PWD/.go-cache go test -bench=. -benchmem ./internal/animation ./internal/terminal
```

## Development

```sh
GOCACHE=$PWD/.go-cache go test ./...
GOCACHE=$PWD/.go-cache go build ./...
```
