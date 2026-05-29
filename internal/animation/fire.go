package animation

import "math"

type fireStyle struct{}

func init() { Register(fireStyle{}) }

func (fireStyle) Name() string        { return "fire" }
func (fireStyle) Description() string { return "flames rise from the bottom of the terminal" }
func (fireStyle) New(width, height int, opts Options) Animator {
	return fireAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type fireAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a fireAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	rise := int((0.15 + t*1.15) * float64(a.height))
	for y := 0; y < a.height; y++ {
		fromBottom := a.height - 1 - y
		if fromBottom > rise {
			continue
		}
		for x := 0; x < a.width; x++ {
			wave := math.Sin(float64(x)*0.35+t*16) + math.Sin(float64(x)*0.11+float64(y)*0.22+t*10)
			heat := float64(rise-fromBottom)/math.Max(1, float64(rise)) + wave*0.16*a.intensity
			if heat < 0.08 {
				continue
			}
			f.Set(x, y, fireCell(heat))
		}
	}
	return f
}

func fireCell(heat float64) Cell {
	switch {
	case heat > 0.9:
		return Cell{Ch: '@', Color: ColorBrightWhite, Bold: true}
	case heat > 0.72:
		return Cell{Ch: '#', Color: ColorBrightYellow, Bold: true}
	case heat > 0.52:
		return Cell{Ch: '*', Color: ColorYellow, Bold: true}
	case heat > 0.32:
		return Cell{Ch: ':', Color: ColorBrightRed}
	default:
		return Cell{Ch: '.', Color: ColorRed}
	}
}

func intensityScale(v string) float64 {
	switch v {
	case "low":
		return 0.7
	case "high":
		return 1.35
	default:
		return 1
	}
}
