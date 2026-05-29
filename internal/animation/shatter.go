//go:build experimental

package animation

import "math"

type shatterStyle struct{}

func init() { Register(shatterStyle{}) }

func (shatterStyle) Name() string { return "shatter" }
func (shatterStyle) Description() string {
	return "screen fragments crack apart and fall away"
}
func (shatterStyle) New(width, height int, opts Options) Animator {
	return shatterAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type shatterAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a shatterAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	blockW := 6
	blockH := 3
	for by := 0; by < a.height; by += blockH {
		for bx := 0; bx < a.width; bx += blockW {
			id := bx/blockW + by/blockH*31
			delay := pseudoNoise(id, by, 0) * 0.35
			local := clamp01((t - delay) / 0.75)
			centerX := float64(bx+blockW/2) - float64(a.width)/2
			dir := 1.0
			if centerX < 0 {
				dir = -1
			}
			offX := int(dir * local * local * float64(a.width) * (0.16 + pseudoNoise(id, bx, 0)*0.22) * a.intensity)
			offY := int(local * local * float64(a.height) * (0.45 + pseudoNoise(by, id, 0)*0.7))
			rot := int(math.Sin(local*math.Pi*2+float64(id)) * 2)
			for y := 0; y < blockH; y++ {
				for x := 0; x < blockW; x++ {
					if pseudoNoise(bx+x, by+y, 0) < 0.18+local*0.38 {
						continue
					}
					ch := '#'
					if local > 0.2 {
						ch = '%'
					}
					if x == 0 || y == 0 || x == blockW-1 || y == blockH-1 {
						ch = '+'
					}
					color := ColorBrightWhite
					if local > 0.35 {
						color = ColorCyan
					}
					f.Set(bx+x+offX+rot, by+y+offY-rot, Cell{Ch: ch, Color: color, Bold: local < 0.35})
				}
			}
		}
	}
	for i := 0; i < a.width; i += 3 {
		y := int(float64(a.height)/2 + math.Sin(float64(i)*0.3+t*8)*float64(a.height)*0.12)
		f.Set(i, y, Cell{Ch: '/', Color: ColorBrightWhite, Bold: true})
	}
	return f
}
