package terminal

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bsushmith/clearfx/internal/animation"
)

func TestRenderFrameIncludesANSI(t *testing.T) {
	frame := animation.NewFrame(2, 1)
	frame.Set(0, 0, animation.Cell{Ch: 'x', Color: animation.ColorRed, Bold: true})
	frame.Set(1, 0, animation.Cell{Ch: 'y'})

	var buf bytes.Buffer
	if err := RenderFrame(&buf, frame); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{Home, Reset, "\x1b[1m", "\x1b[31m", "x", "y"} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered frame missing %q in %q", want, out)
		}
	}
}

func TestClear(t *testing.T) {
	var buf bytes.Buffer
	if err := Clear(&buf); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != Reset+ClearScreen+Home {
		t.Fatalf("clear = %q", got)
	}
}
