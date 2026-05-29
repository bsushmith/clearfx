package app

import (
	"bytes"
	"os"
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
	t.Setenv("CLEARFX_PALETTE", "classic")

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
	printStyles(&buf)
	out := buf.String()
	for _, want := range []string{"style", "category", "duration", "fps", "matrix-rain"} {
		if !strings.Contains(out, want) {
			t.Fatalf("style output missing %q in %q", want, out)
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
	}
}
