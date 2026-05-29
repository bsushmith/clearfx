package animation

import "math"

type lightningStyle struct{}

func init() { Register(lightningStyle{}) }

func (lightningStyle) Name() string { return "lightning" }
func (lightningStyle) Description() string {
	return "branching lightning flashes across the terminal"
}
func (lightningStyle) New(width, height int, opts Options) Animator {
	return lightningAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type lightningAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a lightningAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	flash := math.Sin(t * math.Pi * 8)
	if flash < -0.25 {
		return f
	}
	startX := a.width/2 + int(math.Sin(t*math.Pi*3)*float64(a.width)/4)
	x := startX
	for y := 0; y < a.height; y++ {
		x += int(math.Sin(float64(y)*1.7+t*math.Pi*9) * 1.6 * a.intensity)
		ch := '|'
		if y%3 == 0 {
			ch = '/'
		} else if y%3 == 1 {
			ch = '\\'
		}
		f.Set(x, y, Cell{Ch: ch, Color: ColorBrightWhite, Bold: true})
		if y%5 == 2 {
			drawLightningBranch(f, x, y, -1, 4+y%5)
		}
		if y%7 == 3 {
			drawLightningBranch(f, x, y, 1, 3+y%4)
		}
	}
	return f
}

func drawLightningBranch(f Frame, x, y, dir, length int) {
	for i := 1; i <= length; i++ {
		ch := '/'
		if dir > 0 {
			ch = '\\'
		}
		f.Set(x+dir*i, y+i/2, Cell{Ch: ch, Color: ColorBrightCyan, Bold: true})
	}
}
