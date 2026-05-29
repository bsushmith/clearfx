//go:build experimental

package animation

import "math"

type sandstormStyle struct{}

func init() { Register(sandstormStyle{}) }

func (sandstormStyle) Name() string { return "sandstorm" }
func (sandstormStyle) Description() string {
	return "windblown sand sweeps across the terminal"
}
func (sandstormStyle) New(width, height int, opts Options) Animator {
	return sandstormAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type sandstormAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a sandstormAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	density := int(math.Max(2, 5/a.intensity))
	for y := 0; y < a.height; y++ {
		gust := int(math.Sin(float64(y)*0.23+t*math.Pi*5)*10*a.intensity) + int(t*float64(a.width)*1.7)
		for x := 0; x < a.width; x++ {
			n := pseudoNoise(x-gust, y*3, t*0.25)
			if n > 0.82 || (x+y+gust)%density == 0 {
				ch := '.'
				color := ColorYellow
				if n > 0.94 {
					ch = '*'
					color = ColorBrightYellow
				} else if (x+y)%4 == 0 {
					ch = '-'
				}
				f.Set(x, y, Cell{Ch: ch, Color: color, Bold: n > 0.94})
			}
		}
	}
	return f
}
