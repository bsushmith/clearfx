package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bsushmith/clearfx/internal/animation"
	"github.com/bsushmith/clearfx/internal/terminal"
)

func TestParseEnvDefaults(t *testing.T) {
	t.Setenv("CLEARFX_STYLE", "ocean-wave")
	t.Setenv("CLEARFX_DURATION", "1200ms")
	t.Setenv("CLEARFX_FPS", "45")
	t.Setenv("CLEARFX_INTENSITY", "high")
	t.Setenv("CLEARFX_PALETTE", "aurora")

	cfg, err := parse(nil, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.style != "ocean-wave" {
		t.Fatalf("style = %q", cfg.style)
	}
	if cfg.duration != 1200*time.Millisecond {
		t.Fatalf("duration = %s", cfg.duration)
	}
	if cfg.fps != 45 {
		t.Fatalf("fps = %d", cfg.fps)
	}
	if cfg.intensity != "high" {
		t.Fatalf("intensity = %q", cfg.intensity)
	}
	if cfg.palette != "aurora" {
		t.Fatalf("palette = %q", cfg.palette)
	}
}

func TestParseConfigPresetAndFlagOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
  "style": "fire",
  "preset": "cinematic",
  "presets": {
    "cinematic": {
      "style": "great-wave",
      "duration": "1400ms",
      "fps": 45,
      "palette": "ocean"
    }
  }
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLEARFX_CONFIG", configPath)

	cfg, err := parse([]string{"--palette", "monochrome"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.style != "great-wave" {
		t.Fatalf("style = %q", cfg.style)
	}
	if cfg.duration != 1400*time.Millisecond {
		t.Fatalf("duration = %s", cfg.duration)
	}
	if cfg.palette != "monochrome" {
		t.Fatalf("palette = %q", cfg.palette)
	}
}

func TestParseListPresetsAndShowConfig(t *testing.T) {
	cfg, err := parse([]string{"--list-presets", "--show-config"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.listPresets {
		t.Fatal("expected list-presets flag")
	}
	if !cfg.showConfig {
		t.Fatal("expected show-config flag")
	}
}

func TestParseRunMode(t *testing.T) {
	cfg, err := parse([]string{"run", "--style", "fire", "--", "printf", "ok"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.commandMode {
		t.Fatal("expected run mode")
	}
	if len(cfg.command) != 2 || cfg.command[0] != "printf" || cfg.command[1] != "ok" {
		t.Fatalf("command = %#v", cfg.command)
	}
}

func TestParsePreviewMode(t *testing.T) {
	cfg, err := parse([]string{"preview", "--style", "glitch"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.previewMode {
		t.Fatal("expected preview mode")
	}
	if cfg.style != "glitch" {
		t.Fatalf("style = %q", cfg.style)
	}
}

func TestParseRandomStyle(t *testing.T) {
	cfg, err := parse([]string{"--random-style"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.randomStyle {
		t.Fatal("expected random style flag")
	}

	cfg, err = parse([]string{"--style", "random"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.style != "random" {
		t.Fatalf("style = %q", cfg.style)
	}
}

func TestForceANSIClearsNonInteractiveOutput(t *testing.T) {
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()

	if err := Run([]string{"--force-ansi"}, stdout, stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := stdout.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 64)
	n, err := stdout.Read(buf)
	if err != nil && n == 0 {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != terminal.Reset+terminal.ClearScreen+terminal.Home {
		t.Fatalf("output = %q", got)
	}
}

func TestPrintStylesIncludesMetadata(t *testing.T) {
	var buf bytes.Buffer
	if err := printStyles(&buf, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"style", "category", "duration", "fps", "matrix-rain", "palettes"} {
		if !strings.Contains(out, want) {
			t.Fatalf("style output missing %q in %q", want, out)
		}
	}
}

func TestPrintStylesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := printStyles(&buf, true); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`"name": "fire"`, `"supported_palettes"`, `"experimental"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json style output missing %q in %q", want, out)
		}
	}
}

func TestPrintPalettes(t *testing.T) {
	var buf bytes.Buffer
	if err := printPalettes(&buf, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"classic", "aurora", "monochrome"} {
		if !strings.Contains(out, want) {
			t.Fatalf("palette output missing %q in %q", want, out)
		}
	}
}

func TestPrintPresets(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
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
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := printPresets(&buf, config{configPath: configPath}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"cinematic", "style=great-wave", "fast", "palette=aurora"} {
		if !strings.Contains(out, want) {
			t.Fatalf("preset output missing %q in %q", want, out)
		}
	}
}

func TestPrintPresetsJSON(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{
  "presets": {
    "cinematic": {
      "style": "great-wave",
      "duration": "1400ms",
      "fps": 45,
      "palette": "ocean"
    }
  }
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := printPresets(&buf, config{configPath: configPath, jsonOutput: true}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`"name": "cinematic"`, `"style": "great-wave"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("preset json output missing %q in %q", want, out)
		}
	}
}

func TestPrintResolvedConfig(t *testing.T) {
	var buf bytes.Buffer
	cfg := config{
		style:       "matrix-rain",
		duration:    900 * time.Millisecond,
		fps:         45,
		intensity:   "high",
		palette:     "matrix",
		preset:      "fast",
		configPath:  "/tmp/config.json",
		previewMode: true,
	}
	if err := printResolvedConfig(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"style: matrix-rain", "palette: matrix", "preset: fast", "preview_mode: true"} {
		if !strings.Contains(out, want) {
			t.Fatalf("resolved config output missing %q in %q", want, out)
		}
	}
}

func TestPrintResolvedConfigJSON(t *testing.T) {
	var buf bytes.Buffer
	cfg := config{
		style:      "fire",
		duration:   700 * time.Millisecond,
		fps:        30,
		intensity:  "medium",
		palette:    "ember",
		jsonOutput: true,
	}
	if err := printResolvedConfig(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`"style": "fire"`, `"palette": "ember"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("resolved config json output missing %q in %q", want, out)
		}
	}
}

func TestPreviewIndex(t *testing.T) {
	styles := animation.ListMetadata()
	index := previewIndex(styles, "matrix-rain")
	if index < 0 || index >= len(styles) {
		t.Fatalf("index out of range: %d", index)
	}
	if styles[index].Name != "matrix-rain" {
		t.Fatalf("selected style = %q", styles[index].Name)
	}
}

func TestPrintPreviewCard(t *testing.T) {
	var buf bytes.Buffer
	meta := animation.MetadataFor("fire")
	if err := printPreviewCard(&buf, meta, "ember", 0, 3); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"preview 1/3", "style: fire", "palette: ember", "commands:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("preview output missing %q in %q", want, out)
		}
	}
}

func TestEveryStyleHasMetadata(t *testing.T) {
	for _, name := range animation.Names() {
		meta := animation.MetadataFor(name)
		if meta.Category == "" {
			t.Fatalf("%s missing category", name)
		}
		if meta.RecommendedDuration <= 0 {
			t.Fatalf("%s missing recommended duration", name)
		}
		if meta.RecommendedFPS <= 0 {
			t.Fatalf("%s missing recommended fps", name)
		}
		if len(meta.SupportedPalettes) == 0 {
			t.Fatalf("%s missing supported palettes", name)
		}
	}
}
