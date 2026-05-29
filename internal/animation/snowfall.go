//go:build experimental

package animation

import "math"

type snowfallStyle struct{}

func init() { Register(snowfallStyle{}) }

func (snowfallStyle) Name() string { return "snowfall" }
func (snowfallStyle) Description() string {
	return "snowflakes drift down and blanket the terminal"
}
func (snowfallStyle) New(width, height int, opts Options) Animator {
	return snowfallAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type snowfallAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a snowfallAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	count := int(float64(a.width*a.height) * 0.08 * a.intensity)
	if count < 30 {
		count = 30
	}
	for i := 0; i < count; i++ {
		x0 := (i*17 + 11) % a.width
		y0 := (i*23 + 5) % a.height
		fall := int(t * float64(a.height*2+(i%9)))
		x := x0 + int(math.Sin(t*math.Pi*4+float64(i))*3)
		y := (y0 + fall) % a.height
		ch := '.'
		if i%7 == 0 {
			ch = '*'
		} else if i%5 == 0 {
			ch = '+'
		}
		f.Set(x, y, Cell{Ch: ch, Color: ColorBrightWhite, Bold: ch != '.'})
	}
	snowLine := a.height - int(t*float64(a.height)*0.45)
	for y := snowLine; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			if (x+y)%3 != 0 {
				f.Set(x, y, Cell{Ch: '.', Color: ColorWhite})
			}
		}
	}
	return f
}
