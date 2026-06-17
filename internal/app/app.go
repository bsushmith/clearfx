package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bsushmith/clearfx/internal/animation"
	"github.com/bsushmith/clearfx/internal/terminal"
)

const version = "0.1.0"

type fileConfig struct {
	Style     string                `json:"style"`
	Duration  string                `json:"duration"`
	FPS       int                   `json:"fps"`
	Intensity string                `json:"intensity"`
	Palette   string                `json:"palette"`
	Preset    string                `json:"preset"`
	Presets   map[string]fileConfig `json:"presets"`
}

type config struct {
	duration       time.Duration
	fps            int
	style          string
	intensity      string
	palette        string
	preset         string
	configPath     string
	command        []string
	previewMode    bool
	noAnimation    bool
	listStyles     bool
	listPalettes   bool
	listPresets    bool
	showConfig     bool
	randomStyle    bool
	forceANSI      bool
	version        bool
	help           bool
	jsonOutput     bool
	durationSet    bool
	fpsSet         bool
	commandMode    bool
	explicitConfig bool
}

type jsonStyle struct {
	Name                string   `json:"name"`
	Category            string   `json:"category"`
	Description         string   `json:"description"`
	RecommendedDuration string   `json:"recommended_duration"`
	RecommendedFPS      int      `json:"recommended_fps"`
	SupportedPalettes   []string `json:"supported_palettes"`
	Experimental        bool     `json:"experimental"`
}

type jsonPalette struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonPreset struct {
	Name      string `json:"name"`
	Style     string `json:"style,omitempty"`
	Duration  string `json:"duration,omitempty"`
	FPS       int    `json:"fps,omitempty"`
	Intensity string `json:"intensity,omitempty"`
	Palette   string `json:"palette,omitempty"`
}

type resolvedConfig struct {
	Style       string   `json:"style"`
	Duration    string   `json:"duration"`
	FPS         int      `json:"fps"`
	Intensity   string   `json:"intensity"`
	Palette     string   `json:"palette"`
	Preset      string   `json:"preset,omitempty"`
	ConfigPath  string   `json:"config_path,omitempty"`
	CommandMode bool     `json:"command_mode"`
	PreviewMode bool     `json:"preview_mode"`
	NoAnimation bool     `json:"no_animation"`
	ForceANSI   bool     `json:"force_ansi"`
	Command     []string `json:"command,omitempty"`
}

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
		return printStyles(stdout, cfg.jsonOutput)
	}
	if cfg.listPalettes {
		return printPalettes(stdout, cfg.jsonOutput)
	}
	if cfg.listPresets {
		return printPresets(stdout, cfg)
	}
	if cfg.showConfig {
		return printResolvedConfig(stdout, cfg)
	}
	if cfg.previewMode {
		return runPreview(cfg, stdout, stderr)
	}
	if cfg.randomStyle || cfg.style == "random" {
		cfg.style = animation.RandomName()
		if cfg.style == "" {
			return fmt.Errorf("no animation styles are registered")
		}
	}

	if err := runAnimation(cfg, stdout); err != nil {
		return err
	}
	if cfg.commandMode {
		return runCommand(cfg.command)
	}
	return nil
}

func runAnimation(cfg config, stdout *os.File) error {
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

func runCommand(command []string) error {
	if len(command) == 0 {
		return errors.New("run mode requires a command after --")
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runPreview(cfg config, stdout, stderr *os.File) error {
	if !terminal.IsTerminal(stdout) {
		return fmt.Errorf("preview mode requires a terminal")
	}
	styles := animation.ListMetadata()
	if len(styles) == 0 {
		return fmt.Errorf("no animation styles are registered")
	}
	index := previewIndex(styles, cfg.style)
	reader := bufio.NewReader(os.Stdin)

	for {
		style := styles[index]
		cfg.style = style.Name
		cfg.duration = style.RecommendedDuration
		cfg.fps = style.RecommendedFPS
		cfg.durationSet = true
		cfg.fpsSet = true

		if err := runAnimation(cfg, stdout); err != nil {
			return err
		}
		if err := printPreviewCard(stderr, style, cfg.palette, index, len(styles)); err != nil {
			return err
		}

		input, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		command := strings.TrimSpace(strings.ToLower(input))
		switch command {
		case "", "s", "select":
			return nil
		case "n", "next":
			index = (index + 1) % len(styles)
		case "p", "prev", "previous":
			index = (index - 1 + len(styles)) % len(styles)
		case "r", "replay":
		case "q", "quit", "exit":
			return nil
		default:
			fmt.Fprintln(stderr, "commands: [enter] select  [n] next  [p] previous  [r] replay  [q] quit")
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

func previewIndex(styles []animation.Metadata, selected string) int {
	for i, style := range styles {
		if style.Name == selected {
			return i
		}
	}
	return 0
}

func printPreviewCard(w io.Writer, style animation.Metadata, palette string, index, total int) error {
	_, err := fmt.Fprintf(
		w,
		"preview %d/%d\nstyle: %s\ncategory: %s\npalette: %s\nduration: %s\nfps: %d\nexperimental: %t\ndescription: %s\ncommands: [enter] select  [n] next  [p] previous  [r] replay  [q] quit\n",
		index+1,
		total,
		style.Name,
		style.Category,
		palette,
		style.RecommendedDuration,
		style.RecommendedFPS,
		style.Experimental,
		style.Description,
	)
	return err
}

func printPresets(w io.Writer, cfg config) error {
	fileCfg, err := loadConfigFile(cfg.configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if cfg.jsonOutput {
				return writeJSON(w, []jsonPreset{})
			}
			_, err = fmt.Fprintln(w, "no presets configured")
			return err
		}
		return err
	}
	names := make([]string, 0, len(fileCfg.Presets))
	for name := range fileCfg.Presets {
		names = append(names, name)
	}
	sort.Strings(names)
	if cfg.jsonOutput {
		out := make([]jsonPreset, 0, len(names))
		for _, name := range names {
			preset := fileCfg.Presets[name]
			out = append(out, jsonPreset{
				Name:      name,
				Style:     preset.Style,
				Duration:  preset.Duration,
				FPS:       preset.FPS,
				Intensity: preset.Intensity,
				Palette:   preset.Palette,
			})
		}
		return writeJSON(w, out)
	}
	if len(names) == 0 {
		_, err := fmt.Fprintln(w, "no presets configured")
		return err
	}
	for _, name := range names {
		preset := fileCfg.Presets[name]
		duration := preset.Duration
		if duration == "" {
			duration = "-"
		}
		style := preset.Style
		if style == "" {
			style = "-"
		}
		palette := preset.Palette
		if palette == "" {
			palette = "-"
		}
		intensity := preset.Intensity
		if intensity == "" {
			intensity = "-"
		}
		_, err := fmt.Fprintf(w, "%s  style=%s  duration=%s  fps=%d  intensity=%s  palette=%s\n", name, style, duration, preset.FPS, intensity, palette)
		if err != nil {
			return err
		}
	}
	return nil
}

func printResolvedConfig(w io.Writer, cfg config) error {
	out := resolvedConfig{
		Style:       cfg.style,
		Duration:    cfg.duration.String(),
		FPS:         cfg.fps,
		Intensity:   cfg.intensity,
		Palette:     cfg.palette,
		Preset:      cfg.preset,
		ConfigPath:  cfg.configPath,
		CommandMode: cfg.commandMode,
		PreviewMode: cfg.previewMode,
		NoAnimation: cfg.noAnimation,
		ForceANSI:   cfg.forceANSI,
		Command:     cfg.command,
	}
	if cfg.jsonOutput {
		return writeJSON(w, out)
	}
	_, err := fmt.Fprintf(
		w,
		"style: %s\nduration: %s\nfps: %d\nintensity: %s\npalette: %s\npreset: %s\nconfig: %s\ncommand_mode: %t\npreview_mode: %t\nno_animation: %t\nforce_ansi: %t\n",
		out.Style,
		out.Duration,
		out.FPS,
		out.Intensity,
		out.Palette,
		emptyDash(out.Preset),
		emptyDash(out.ConfigPath),
		out.CommandMode,
		out.PreviewMode,
		out.NoAnimation,
		out.ForceANSI,
	)
	return err
}

func parse(args []string, stderr io.Writer) (config, error) {
	cfg := config{
		duration:  700 * time.Millisecond,
		fps:       30,
		style:     "fire",
		intensity: "medium",
		palette:   "classic",
	}
	if len(args) > 0 && args[0] == "run" {
		cfg.commandMode = true
		args = args[1:]
	} else if len(args) > 0 && args[0] == "preview" {
		cfg.previewMode = true
		args = args[1:]
	}
	flagArgs, commandArgs := splitCommandArgs(args)
	cfg.command = commandArgs

	var configFile fileConfig
	configPath := os.Getenv("CLEARFX_CONFIG")
	if configPath == "" {
		configPath = defaultConfigPath()
	} else {
		cfg.explicitConfig = true
	}
	cfg.configPath = configPath
	if loaded, err := loadConfigFile(configPath); err == nil {
		configFile = loaded
		applyFileConfig(&cfg, loaded)
	} else if !errors.Is(err, os.ErrNotExist) && cfg.explicitConfig {
		return cfg, err
	}

	applyEnvConfig(&cfg)

	fs := flag.NewFlagSet("clearfx", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.DurationVar(&cfg.duration, "duration", cfg.duration, "animation duration")
	fs.IntVar(&cfg.fps, "fps", cfg.fps, "animation frames per second")
	fs.StringVar(&cfg.style, "style", cfg.style, "animation style")
	fs.StringVar(&cfg.intensity, "intensity", cfg.intensity, "animation intensity: low, medium, high")
	fs.StringVar(&cfg.palette, "palette", cfg.palette, "color palette")
	fs.StringVar(&cfg.preset, "preset", cfg.preset, "config preset name")
	fs.StringVar(&cfg.configPath, "config", cfg.configPath, "path to config file")
	fs.BoolVar(&cfg.noAnimation, "no-animation", false, "clear immediately without animation")
	fs.BoolVar(&cfg.listStyles, "list-styles", false, "print available styles and exit")
	fs.BoolVar(&cfg.listPalettes, "list-palettes", false, "print available palettes and exit")
	fs.BoolVar(&cfg.listPresets, "list-presets", false, "print configured presets and exit")
	fs.BoolVar(&cfg.showConfig, "show-config", false, "print resolved config and exit")
	fs.BoolVar(&cfg.randomStyle, "random-style", false, "pick a random available style")
	fs.BoolVar(&cfg.forceANSI, "force-ansi", false, "emit ANSI clear even when stdout is not a terminal")
	fs.BoolVar(&cfg.version, "version", false, "print version and exit")
	fs.BoolVar(&cfg.help, "help", false, "print help and exit")
	fs.BoolVar(&cfg.jsonOutput, "json", false, "print structured JSON output for listing commands")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: clearfx [flags]")
		fmt.Fprintln(stderr, "       clearfx run [flags] -- <command> [args]")
		fmt.Fprintln(stderr, "       clearfx preview [flags]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(flagArgs); err != nil {
		return cfg, err
	}
	visited := map[string]string{}
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = f.Value.String()
		switch f.Name {
		case "duration":
			cfg.durationSet = true
		case "fps":
			cfg.fpsSet = true
		case "config":
			cfg.explicitConfig = true
		}
	})
	if cfg.configPath != configPath {
		loaded, err := loadConfigFile(cfg.configPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return cfg, err
		}
		if err == nil {
			configFile = loaded
			cfg = config{
				duration:       700 * time.Millisecond,
				fps:            30,
				style:          "fire",
				intensity:      "medium",
				palette:        "classic",
				configPath:     cfg.configPath,
				commandMode:    cfg.commandMode,
				command:        cfg.command,
				explicitConfig: true,
			}
			applyFileConfig(&cfg, loaded)
			applyEnvConfig(&cfg)
			fs = flag.NewFlagSet("clearfx", flag.ContinueOnError)
			fs.SetOutput(stderr)
			fs.DurationVar(&cfg.duration, "duration", cfg.duration, "animation duration")
			fs.IntVar(&cfg.fps, "fps", cfg.fps, "animation frames per second")
			fs.StringVar(&cfg.style, "style", cfg.style, "animation style")
			fs.StringVar(&cfg.intensity, "intensity", cfg.intensity, "animation intensity: low, medium, high")
			fs.StringVar(&cfg.palette, "palette", cfg.palette, "color palette")
			fs.StringVar(&cfg.preset, "preset", cfg.preset, "config preset name")
			fs.StringVar(&cfg.configPath, "config", cfg.configPath, "path to config file")
			fs.BoolVar(&cfg.noAnimation, "no-animation", false, "clear immediately without animation")
			fs.BoolVar(&cfg.listStyles, "list-styles", false, "print available styles and exit")
			fs.BoolVar(&cfg.listPalettes, "list-palettes", false, "print available palettes and exit")
			fs.BoolVar(&cfg.listPresets, "list-presets", false, "print configured presets and exit")
			fs.BoolVar(&cfg.showConfig, "show-config", false, "print resolved config and exit")
			fs.BoolVar(&cfg.randomStyle, "random-style", false, "pick a random available style")
			fs.BoolVar(&cfg.forceANSI, "force-ansi", false, "emit ANSI clear even when stdout is not a terminal")
			fs.BoolVar(&cfg.version, "version", false, "print version and exit")
			fs.BoolVar(&cfg.help, "help", false, "print help and exit")
			fs.BoolVar(&cfg.jsonOutput, "json", false, "print structured JSON output for listing commands")
			fs.Usage = func() {
				fmt.Fprintln(stderr, "Usage: clearfx [flags]")
				fmt.Fprintln(stderr, "       clearfx run [flags] -- <command> [args]")
				fmt.Fprintln(stderr, "       clearfx preview [flags]")
				fs.PrintDefaults()
			}
			if err := fs.Parse(flagArgs); err != nil {
				return cfg, err
			}
			visited = map[string]string{}
			fs.Visit(func(f *flag.Flag) {
				visited[f.Name] = f.Value.String()
				switch f.Name {
				case "duration":
					cfg.durationSet = true
				case "fps":
					cfg.fpsSet = true
				}
			})
		}
	}
	if cfg.preset != "" {
		if err := applyPreset(&cfg, configFile, cfg.preset); err != nil {
			return cfg, err
		}
		if err := reapplyVisitedFlags(&cfg, visited); err != nil {
			return cfg, err
		}
	}
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
	cfg.palette = strings.ToLower(cfg.palette)
	if _, ok := knownPalette(cfg.palette); !ok {
		return cfg, fmt.Errorf("unknown palette %q; available palettes: %s", cfg.palette, strings.Join(animation.PaletteNames(), ", "))
	}
	if _, err := animation.Get(cfg.style); err != nil && cfg.style != "random" && !cfg.randomStyle && !cfg.noAnimation && !cfg.listStyles && !cfg.listPalettes && !cfg.version {
		return cfg, err
	}
	if cfg.commandMode && len(cfg.command) == 0 {
		return cfg, fmt.Errorf("run mode requires a command after --")
	}
	if cfg.previewMode && cfg.commandMode {
		return cfg, fmt.Errorf("preview mode cannot be combined with run mode")
	}
	return cfg, nil
}

func applyEnvConfig(cfg *config) {
	if value := os.Getenv("CLEARFX_STYLE"); value != "" {
		cfg.style = value
	}
	if value := os.Getenv("CLEARFX_DURATION"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			cfg.duration = duration
			cfg.durationSet = true
		}
	}
	if value := os.Getenv("CLEARFX_FPS"); value != "" {
		if fps, err := strconv.Atoi(value); err == nil {
			cfg.fps = fps
			cfg.fpsSet = true
		}
	}
	if value := os.Getenv("CLEARFX_INTENSITY"); value != "" {
		cfg.intensity = value
	}
	if value := os.Getenv("CLEARFX_PALETTE"); value != "" {
		cfg.palette = value
	}
	if value := os.Getenv("CLEARFX_PRESET"); value != "" {
		cfg.preset = value
	}
}

func applyFileConfig(cfg *config, fc fileConfig) {
	if fc.Style != "" {
		cfg.style = fc.Style
	}
	if fc.Duration != "" {
		if duration, err := time.ParseDuration(fc.Duration); err == nil {
			cfg.duration = duration
		}
	}
	if fc.FPS > 0 {
		cfg.fps = fc.FPS
	}
	if fc.Intensity != "" {
		cfg.intensity = fc.Intensity
	}
	if fc.Palette != "" {
		cfg.palette = fc.Palette
	}
	if fc.Preset != "" {
		cfg.preset = fc.Preset
	}
}

func applyPreset(cfg *config, fc fileConfig, name string) error {
	preset, ok := fc.Presets[name]
	if !ok {
		return fmt.Errorf("unknown preset %q in %s", name, cfg.configPath)
	}
	if preset.Style != "" {
		cfg.style = preset.Style
	}
	if preset.Duration != "" && !cfg.durationSet {
		duration, err := time.ParseDuration(preset.Duration)
		if err != nil {
			return fmt.Errorf("preset %q duration: %w", name, err)
		}
		cfg.duration = duration
	}
	if preset.FPS > 0 && !cfg.fpsSet {
		cfg.fps = preset.FPS
	}
	if preset.Intensity != "" {
		cfg.intensity = preset.Intensity
	}
	if preset.Palette != "" {
		cfg.palette = preset.Palette
	}
	return nil
}

func loadConfigFile(path string) (fileConfig, error) {
	var cfg fileConfig
	if path == "" {
		return cfg, os.ErrNotExist
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func defaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "clearfx", "config.json")
}

func splitCommandArgs(args []string) ([]string, []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

func validIntensity(v string) bool {
	return v == "low" || v == "medium" || v == "high"
}

func knownPalette(name string) (animation.Palette, bool) {
	for _, paletteName := range animation.PaletteNames() {
		if paletteName == name {
			return animation.PaletteFor(name), true
		}
	}
	return animation.Palette{}, false
}

func emptyDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func reapplyVisitedFlags(cfg *config, visited map[string]string) error {
	for name, value := range visited {
		switch name {
		case "style":
			cfg.style = value
		case "duration":
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			cfg.duration = duration
		case "fps":
			fps, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			cfg.fps = fps
		case "intensity":
			cfg.intensity = value
		case "palette":
			cfg.palette = value
		case "preset":
			cfg.preset = value
		case "config":
			cfg.configPath = value
		case "no-animation":
			cfg.noAnimation = value == "true"
		case "list-styles":
			cfg.listStyles = value == "true"
		case "list-palettes":
			cfg.listPalettes = value == "true"
		case "list-presets":
			cfg.listPresets = value == "true"
		case "show-config":
			cfg.showConfig = value == "true"
		case "random-style":
			cfg.randomStyle = value == "true"
		case "force-ansi":
			cfg.forceANSI = value == "true"
		case "version":
			cfg.version = value == "true"
		case "help":
			cfg.help = value == "true"
		case "json":
			cfg.jsonOutput = value == "true"
		}
	}
	return nil
}

func printStyles(w io.Writer, jsonOutput bool) error {
	styles := animation.ListMetadata()
	if jsonOutput {
		out := make([]jsonStyle, 0, len(styles))
		for _, style := range styles {
			out = append(out, jsonStyle{
				Name:                style.Name,
				Category:            style.Category,
				Description:         style.Description,
				RecommendedDuration: style.RecommendedDuration.String(),
				RecommendedFPS:      style.RecommendedFPS,
				SupportedPalettes:   style.SupportedPalettes,
				Experimental:        style.Experimental,
			})
		}
		return writeJSON(w, out)
	}
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
	fmt.Fprintf(w, "%-*s  %-*s  %-8s  %-3s  %-12s  %s\n", nameWidth, "style", categoryWidth, "category", "duration", "fps", "palettes", "description")
	for _, style := range styles {
		fmt.Fprintf(w, "%-*s  %-*s  %-8s  %-3d  %-12s  %s\n",
			nameWidth,
			style.Name,
			categoryWidth,
			style.Category,
			style.RecommendedDuration,
			style.RecommendedFPS,
			strings.Join(style.SupportedPalettes, ","),
			style.Description,
		)
	}
	return nil
}

func printPalettes(w io.Writer, jsonOutput bool) error {
	names := animation.PaletteNames()
	if jsonOutput {
		out := make([]jsonPalette, 0, len(names))
		for _, name := range names {
			palette := animation.PaletteFor(name)
			out = append(out, jsonPalette{Name: palette.Name, Description: palette.Description})
		}
		return writeJSON(w, out)
	}
	for _, name := range names {
		palette := animation.PaletteFor(name)
		fmt.Fprintf(w, "%-10s  %s\n", palette.Name, palette.Description)
	}
	return nil
}

func writeJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
