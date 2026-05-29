package animation

import "math"

type glitchStyle struct{}

func init() { Register(glitchStyle{}) }

func (glitchStyle) Name() string { return "glitch" }
func (glitchStyle) Description() string {
	return "scrambled blocks and scanlines snap the terminal clear"
}
func (glitchStyle) New(width, height int, opts Options) Animator {
	return glitchAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type glitchAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a glitchAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	chars := []rune("@#%&$+=-:.;")
	coverage := 1 - math.Abs(t*2-1)*0.35
	for y := 0; y < a.height; y++ {
		shift := int(math.Sin(float64(y)*0.9+t*math.Pi*10) * 8 * a.intensity)
		scanline := (y+int(t*40))%5 == 0
		for x := 0; x < a.width; x++ {
			n := pseudoNoise(x+shift, y, t)
			if n > coverage {
				continue
			}
			ch := chars[(x*11+y*7+int(t*50))%len(chars)]
			color := ColorBrightCyan
			if scanline {
				color = ColorBrightWhite
			} else if (x+y+int(t*20))%3 == 0 {
				color = ColorBrightMagenta
			} else if (x+y)%4 == 0 {
				color = ColorBrightGreen
			}
			f.Set(x, y, Cell{Ch: ch, Color: color, Bold: scanline})
		}
	}
	return f
}

func pseudoNoise(x, y int, t float64) float64 {
	n := uint32(x)*374761393 + uint32(y)*668265263 + uint32(t*1024)*2246822519
	n = (n ^ (n >> 13)) * 1274126177
	n ^= n >> 16
	return float64(n&0xffff) / 65535
}
