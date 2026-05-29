package animation

import "math"

type blackHoleStyle struct{}

func init() { Register(blackHoleStyle{}) }

func (blackHoleStyle) Name() string { return "black-hole" }
func (blackHoleStyle) Description() string {
	return "particles spiral inward and collapse into the center"
}
func (blackHoleStyle) New(width, height int, opts Options) Animator {
	return blackHoleAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type blackHoleAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a blackHoleAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	cx := float64(a.width-1) / 2
	cy := float64(a.height-1) / 2
	maxR := math.Hypot(cx, cy)
	count := int(float64(a.width*a.height) * 0.11 * a.intensity)
	if count < 36 {
		count = 36
	}
	collapse := math.Pow(1-t, 1.4)
	for i := 0; i < count; i++ {
		seed := float64(i)
		startR := maxR * (0.28 + math.Mod(seed*0.618, 0.78))
		r := startR * collapse
		angle := seed*2.399 + t*math.Pi*(4.2+math.Mod(seed, 7)*0.18)
		x := int(cx + math.Cos(angle)*r)
		y := int(cy + math.Sin(angle)*r*0.52)
		if x < 0 || x >= a.width || y < 0 || y >= a.height {
			continue
		}
		ch := '.'
		color := ColorBrightBlue
		if r < maxR*0.16 {
			ch = '*'
			color = ColorBrightMagenta
		} else if i%5 == 0 {
			ch = '+'
			color = ColorBrightCyan
		}
		f.Set(x, y, Cell{Ch: ch, Color: color, Bold: r < maxR*0.25})
	}

	ringR := math.Max(2, maxR*0.18*(1-t*0.5))
	for i := 0; i < 96; i++ {
		angle := float64(i)/96*math.Pi*2 + t*math.Pi*5
		x := int(cx + math.Cos(angle)*ringR)
		y := int(cy + math.Sin(angle)*ringR*0.45)
		f.Set(x, y, Cell{Ch: '@', Color: ColorBrightWhite, Bold: true})
	}
	f.Set(int(cx), int(cy), Cell{Ch: ' ', Color: ColorDefault})
	return f
}
