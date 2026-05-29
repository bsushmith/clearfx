//go:build experimental

package animation

import "math"

type rainstormStyle struct{}

func init() { Register(rainstormStyle{}) }

func (rainstormStyle) Name() string { return "rainstorm" }
func (rainstormStyle) Description() string {
	return "slanted rain falls with occasional lightning flashes"
}
func (rainstormStyle) New(width, height int, opts Options) Animator {
	return rainstormAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type rainstormAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a rainstormAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	count := int(float64(a.width*a.height) * 0.16 * a.intensity)
	if count < 40 {
		count = 40
	}
	for i := 0; i < count; i++ {
		x0 := (i*23 + 5) % (a.width + a.height)
		y0 := (i*31 + 11) % a.height
		fall := int(t * float64(a.height*3+(i%17)))
		x := x0 - fall/3
		y := (y0 + fall) % a.height
		cell := Cell{Ch: '/', Color: ColorBrightBlue}
		if i%4 == 0 {
			cell = Cell{Ch: '|', Color: ColorCyan}
		}
		f.Set(x, y, cell)
		f.Set(x+1, y-1, Cell{Ch: '/', Color: ColorBlue})
	}

	if math.Sin(t*math.Pi*10) > 0.82 {
		x := a.width/3 + int(math.Sin(t*math.Pi*3)*float64(a.width)/5)
		for y := 0; y < a.height/2; y++ {
			x += int(math.Sin(float64(y)*1.4+t*12) * 1.2)
			f.Set(x, y, Cell{Ch: '/', Color: ColorBrightWhite, Bold: true})
			if y%4 == 0 {
				f.Set(x+1, y+1, Cell{Ch: '\\', Color: ColorBrightCyan, Bold: true})
			}
		}
	}
	return f
}
