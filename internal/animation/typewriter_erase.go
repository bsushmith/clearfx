package animation

import "math"

type typewriterEraseStyle struct{}

func init() { Register(typewriterEraseStyle{}) }

func (typewriterEraseStyle) Name() string { return "typewriter-erase" }
func (typewriterEraseStyle) Description() string {
	return "rows erase with a typewriter cursor trail"
}
func (typewriterEraseStyle) New(width, height int, opts Options) Animator {
	return typewriterEraseAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity), palette: PaletteFor(opts.Palette)}
}

type typewriterEraseAnimator struct {
	width     int
	height    int
	intensity float64
	palette   Palette
}

func (a typewriterEraseAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	total := a.width * a.height
	cursor := int(t * float64(total+a.width))
	glyphs := []rune("abcdefghijklmnopqrstuvwxyz0123456789{}[]();:$#")
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			pos := y*a.width + x
			if pos < cursor-a.width/2 {
				continue
			}
			if pos <= cursor {
				f.Set(x, y, Cell{Ch: '_', Color: a.palette.Primary, Bold: true})
				continue
			}
			idx := (x*7 + y*13 + int(math.Sin(t*8)*10)) % len(glyphs)
			color := a.palette.Dim
			if pos-cursor < a.width {
				color = a.palette.Neutral
			}
			f.Set(x, y, Cell{Ch: glyphs[idx], Color: color})
		}
	}
	return f
}
