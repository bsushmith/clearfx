//go:build experimental

package animation

import "math"

type auroraStyle struct{}

func init() { Register(auroraStyle{}) }

func (auroraStyle) Name() string { return "aurora" }
func (auroraStyle) Description() string {
	return "soft colored aurora bands drift across the terminal"
}
func (auroraStyle) New(width, height int, opts Options) Animator {
	return auroraAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type auroraAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a auroraAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	bands := 4
	for band := 0; band < bands; band++ {
		base := float64(a.height) * (0.18 + float64(band)*0.12)
		amp := float64(a.height) * (0.08 + float64(band)*0.015) * a.intensity
		for x := 0; x < a.width; x++ {
			phase := float64(x)*0.11 + t*math.Pi*(1.5+float64(band)*0.35) + float64(band)
			y := int(base + math.Sin(phase)*amp + math.Sin(phase*0.37)*amp*0.55)
			thick := 2 + band%2
			for dy := -thick; dy <= thick; dy++ {
				ch := '~'
				color := ColorBrightGreen
				if band%3 == 1 {
					color = ColorBrightCyan
				} else if band%3 == 2 {
					color = ColorBrightMagenta
				}
				if dy == 0 {
					ch = '='
				} else if math.Abs(float64(dy)) == float64(thick) {
					ch = '.'
				}
				f.Set(x, y+dy, Cell{Ch: ch, Color: color, Bold: dy == 0})
			}
		}
	}
	return f
}
