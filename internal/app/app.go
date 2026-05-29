package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bsushmith/clearfx/internal/animation"
	"github.com/bsushmith/clearfx/internal/terminal"
)

const version = "0.1.0"

func Run(args []string, stdout, stderr *os.File) error {
	cfg, err := parse(args, stderr)
	if err != nil {
		return err
	}
	if cfg.help {
		return nil
	}
	if cfg.version {
		fmt.Fprintln(stdout, version)
		return nil
	}
	if cfg.listStyles {
		printStyles(stdout)
		return nil
	}
	if cfg.randomStyle || cfg.style == "random" {
		cfg.style = animation.RandomName()
		if cfg.style == "" {
			return fmt.Errorf("no animation styles are registered")
		}
	}

	if !terminal.IsTerminal(stdout) {
		if cfg.forceANSI || cfg.noAnimation {
			return terminal.Clear(stdout)
		}
		return nil
	}

	width, height := terminal.Size(stdout)
	if cfg.noAnimation || width < 20 || height < 8 {
		return terminal.Clear(stdout)
	}

	style, err := animation.Get(cfg.style)
	if err != nil {
		return err
	}
	if !cfg.durationSet {
		if recommended := animation.MetadataFor(style.Name()).RecommendedDuration; recommended > 0 {
			cfg.duration = recommended
		}
	}
	if !cfg.fpsSet {
		if recommended := animation.MetadataFor(style.Name()).RecommendedFPS; recommended > 0 {
			cfg.fps = recommended
		}
	}
	anim := style.New(width, height, animation.Options{
		Duration:  cfg.duration,
		FPS:       cfg.fps,
		Intensity: cfg.intensity,
		Palette:   cfg.palette,
	})

	signals := make(chan os.Signal, 4)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGWINCH)
	defer signal.Stop(signals)

	cleanup := func() {
		io.WriteString(stdout, terminal.Reset+terminal.ShowCursor)
	}
	defer cleanup()
	defer func() {
		if r := recover(); r != nil {
			cleanup()
			panic(r)
		}
	}()

	if _, err := io.WriteString(stdout, terminal.HideCursor+terminal.ClearScreen+terminal.Home); err != nil {
		return err
	}

	frames := int(cfg.duration.Seconds() * float64(cfg.fps))
	if frames < 1 {
		frames = 1
	}
	delay := time.Second / time.Duration(cfg.fps)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for i := 0; i <= frames; i++ {
		resized := false
		for {
			select {
			case sig := <-signals:
				if sig == syscall.SIGWINCH {
					resized = true
					continue
				}
				cleanup()
				os.Exit(130)
			default:
			}
			break
		}
		if resized {
			newWidth, newHeight := terminal.Size(stdout)
			if newWidth != width || newHeight != height {
				width, height = newWidth, newHeight
				if width < 20 || height < 8 {
					return terminal.Clear(stdout)
				}
				anim = style.New(width, height, animation.Options{
					Duration:  cfg.duration,
					FPS:       cfg.fps,
					Intensity: cfg.intensity,
					Palette:   cfg.palette,
				})
				if _, err := io.WriteString(stdout, terminal.ClearScreen+terminal.Home); err != nil {
					return err
				}
			}
		}
		t := float64(i) / float64(frames)
		if err := terminal.RenderFrame(stdout, anim.Frame(t)); err != nil {
			return err
		}
		if i < frames {
			<-ticker.C
		}
	}
	return terminal.Clear(stdout)
}

type config struct {
	duration    time.Duration
	fps         int
	style       string
	intensity   string
	palette     string
	noAnimation bool
	listStyles  bool
	randomStyle bool
	forceANSI   bool
	version     bool
	help        bool
	durationSet bool
	fpsSet      bool
}

func parse(args []string, stderr io.Writer) (config, error) {
	cfg := config{
		duration:  700 * time.Millisecond,
		fps:       30,
		style:     "fire",
		intensity: "medium",
		palette:   "classic",
	}
	if value := os.Getenv("CLEARFX_STYLE"); value != "" {
		cfg.style = value
	}
	if value := os.Getenv("CLEARFX_DURATION"); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return cfg, fmt.Errorf("CLEARFX_DURATION: %w", err)
		}
		cfg.duration = duration
		cfg.durationSet = true
	}
	if value := os.Getenv("CLEARFX_FPS"); value != "" {
		fps, err := strconv.Atoi(value)
		if err != nil {
			return cfg, fmt.Errorf("CLEARFX_FPS: %w", err)
		}
		cfg.fps = fps
		cfg.fpsSet = true
	}
	if value := os.Getenv("CLEARFX_INTENSITY"); value != "" {
		cfg.intensity = value
	}
	if value := os.Getenv("CLEARFX_PALETTE"); value != "" {
		cfg.palette = value
	}
	fs := flag.NewFlagSet("clearfx", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.DurationVar(&cfg.duration, "duration", cfg.duration, "animation duration")
	fs.IntVar(&cfg.fps, "fps", cfg.fps, "animation frames per second")
	fs.StringVar(&cfg.style, "style", cfg.style, "animation style")
	fs.StringVar(&cfg.intensity, "intensity", cfg.intensity, "animation intensity: low, medium, high")
	fs.StringVar(&cfg.palette, "palette", cfg.palette, "color palette")
	fs.BoolVar(&cfg.noAnimation, "no-animation", false, "clear immediately without animation")
	fs.BoolVar(&cfg.listStyles, "list-styles", false, "print available styles and exit")
	fs.BoolVar(&cfg.randomStyle, "random-style", false, "pick a random available style")
	fs.BoolVar(&cfg.forceANSI, "force-ansi", false, "emit ANSI clear even when stdout is not a terminal")
	fs.BoolVar(&cfg.version, "version", false, "print version and exit")
	fs.BoolVar(&cfg.help, "help", false, "print help and exit")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: clearfx [flags]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "duration":
			cfg.durationSet = true
		case "fps":
			cfg.fpsSet = true
		}
	})
	if cfg.help {
		fs.Usage()
		return cfg, nil
	}
	if cfg.duration < 50*time.Millisecond {
		return cfg, fmt.Errorf("duration must be at least 50ms")
	}
	if cfg.fps < 1 || cfg.fps > 120 {
		return cfg, fmt.Errorf("fps must be between 1 and 120")
	}
	if !validIntensity(cfg.intensity) {
		return cfg, fmt.Errorf("intensity must be one of: low, medium, high")
	}
	cfg.style = strings.ToLower(cfg.style)
	if _, err := animation.Get(cfg.style); err != nil && cfg.style != "random" && !cfg.randomStyle && !cfg.noAnimation && !cfg.listStyles && !cfg.version {
		return cfg, err
	}
	return cfg, nil
}

func validIntensity(v string) bool {
	return v == "low" || v == "medium" || v == "high"
}

func printStyles(w io.Writer) {
	styles := animation.ListMetadata()
	nameWidth := len("style")
	categoryWidth := len("category")
	for _, style := range styles {
		if len(style.Name) > nameWidth {
			nameWidth = len(style.Name)
		}
		if len(style.Category) > categoryWidth {
			categoryWidth = len(style.Category)
		}
	}
	fmt.Fprintf(w, "%-*s  %-*s  %-8s  %-3s  %s\n", nameWidth, "style", categoryWidth, "category", "duration", "fps", "description")
	for _, style := range styles {
		fmt.Fprintf(
			w,
			"%-*s  %-*s  %-8s  %-3d  %s\n",
			nameWidth,
			style.Name,
			categoryWidth,
			style.Category,
			style.RecommendedDuration,
			style.RecommendedFPS,
			style.Description,
		)
	}
}
