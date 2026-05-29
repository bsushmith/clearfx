package animation

import "math"

type starfieldStyle struct{}

func init() { Register(starfieldStyle{}) }

func (starfieldStyle) Name() string { return "starfield" }
func (starfieldStyle) Description() string {
	return "stars rush toward the viewer at warp speed"
}
func (starfieldStyle) New(width, height int, opts Options) Animator {
	return starfieldAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type starfieldAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a starfieldAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	cx := float64(a.width-1) / 2
	cy := float64(a.height-1) / 2
	count := int(float64(a.width*a.height) * 0.08 * a.intensity)
	if count < 24 {
		count = 24
	}
	for i := 0; i < count; i++ {
		angle := float64((i*137)%360) * math.Pi / 180
		base := 0.08 + float64((i*29)%100)/100
		dist := math.Mod(base+t*(1.6+float64(i%5)*0.08), 1)
		r := dist * math.Max(float64(a.width), float64(a.height)) * 0.72
		x := int(cx + math.Cos(angle)*r)
		y := int(cy + math.Sin(angle)*r*0.55)
		if x < 0 || x >= a.width || y < 0 || y >= a.height {
			continue
		}
		ch := '.'
		color := ColorWhite
		if dist > 0.62 {
			ch = '*'
			color = ColorBrightWhite
		}
		if dist > 0.82 {
			ch = '+'
			color = ColorBrightCyan
			drawStarTrail(f, x, y, cx, cy, dist)
		}
		f.Set(x, y, Cell{Ch: ch, Color: color, Bold: dist > 0.62})
	}
	return f
}

func drawStarTrail(f Frame, x, y int, cx, cy, dist float64) {
	dx := float64(x) - cx
	dy := float64(y) - cy
	length := int(1 + dist*4)
	for i := 1; i <= length; i++ {
		tx := x - int(dx*0.04*float64(i))
		ty := y - int(dy*0.05*float64(i))
		f.Set(tx, ty, Cell{Ch: '-', Color: ColorCyan})
	}
}
