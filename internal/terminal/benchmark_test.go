package terminal

import (
	"io"
	"testing"

	"github.com/bsushmith/clearfx/internal/animation"
)

func BenchmarkRenderFrame(b *testing.B) {
	frame := animation.NewFrame(120, 40)
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			color := animation.ColorCyan
			bold := false
			if (x+y)%7 == 0 {
				color = animation.ColorBrightWhite
				bold = true
			}
			frame.Set(x, y, animation.Cell{Ch: '~', Color: color, Bold: bold})
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := RenderFrame(io.Discard, frame); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStylesRenderFrame(b *testing.B) {
	const (
		width  = 120
		height = 40
	)
	opts := animation.Options{Intensity: "medium"}
	for _, style := range animation.List() {
		b.Run(style.Name(), func(b *testing.B) {
			anim := style.New(width, height, opts)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				frame := anim.Frame(float64(i%60) / 59)
				if err := RenderFrame(io.Discard, frame); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
