package animation

import "math"

type meteorShowerStyle struct{}

func init() { Register(meteorShowerStyle{}) }

func (meteorShowerStyle) Name() string { return "meteor-shower" }
func (meteorShowerStyle) Description() string {
	return "diagonal meteors streak across the terminal"
}
func (meteorShowerStyle) New(width, height int, opts Options) Animator {
	return meteorShowerAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity), palette: PaletteFor(opts.Palette)}
}

type meteorShowerAnimator struct {
	width     int
	height    int
	intensity float64
	palette   Palette
}

func (a meteorShowerAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	count := int(8 + 10*a.intensity)
	for i := 0; i < count; i++ {
		startX := (i*19 + 7) % (a.width + a.height)
		startY := -((i * 5) % a.height)
		speed := float64(a.width+a.height) * (0.45 + float64(i%5)*0.08)
		x := startX + int(t*speed) - a.height/3
		y := startY + int(t*speed*0.38) + (i % 4)
		length := 5 + i%7
		for j := 0; j < length; j++ {
			tx := x - j
			ty := y - j/2
			cell := Cell{Ch: '.', Color: a.palette.Warm}
			if j == 0 {
				cell = Cell{Ch: '*', Color: a.palette.Primary, Bold: true}
			} else if j < 3 {
				cell = Cell{Ch: '+', Color: a.palette.Accent, Bold: true}
			} else if math.Sin(float64(j)+t*10) > 0 {
				cell.Color = a.palette.Highlight
			}
			f.Set(tx, ty, cell)
		}
	}
	return f
}
