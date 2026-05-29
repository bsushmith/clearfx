package animation

import "math"

type pageBurnStyle struct{}

func init() { Register(pageBurnStyle{}) }

func (pageBurnStyle) Name() string { return "page-burn" }
func (pageBurnStyle) Description() string {
	return "burning edges consume the screen inward"
}
func (pageBurnStyle) New(width, height int, opts Options) Animator {
	return pageBurnAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type pageBurnAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a pageBurnAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	cx := float64(a.width-1) / 2
	cy := float64(a.height-1) / 2
	maxD := math.Hypot(cx, cy)
	burn := t * maxD * 1.25
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			dx := math.Abs(float64(x) - cx)
			dy := math.Abs(float64(y) - cy)
			edgeDist := math.Min(math.Min(float64(x), float64(a.width-1-x)), math.Min(float64(y), float64(a.height-1-y)))
			front := burn - edgeDist*1.8 + math.Sin(float64(x)*0.33+float64(y)*0.29+t*9)*2.5*a.intensity
			if front < 0 {
				continue
			}
			centerDist := math.Hypot(dx, dy)
			if centerDist < maxD-burn*0.82 {
				continue
			}
			cell := Cell{Ch: '.', Color: ColorRed}
			if front > 7 {
				cell = Cell{Ch: ' ', Color: ColorDefault}
			} else if front > 4 {
				cell = Cell{Ch: '#', Color: ColorBrightRed, Bold: true}
			} else if front > 2 {
				cell = Cell{Ch: '*', Color: ColorBrightYellow, Bold: true}
			} else {
				cell = Cell{Ch: '@', Color: ColorBrightWhite, Bold: true}
			}
			f.Set(x, y, cell)
		}
	}
	return f
}
