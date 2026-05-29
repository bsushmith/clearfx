//go:build experimental

package animation

import "math"

type inkDropStyle struct{}

func init() { Register(inkDropStyle{}) }

func (inkDropStyle) Name() string { return "ink-drop" }
func (inkDropStyle) Description() string {
	return "expanding ink blots consume the terminal"
}
func (inkDropStyle) New(width, height int, opts Options) Animator {
	return inkDropAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type inkDropAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a inkDropAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	for _, drop := range inkDrops {
		cx := drop.x * float64(a.width)
		cy := drop.y * float64(a.height)
		radius := t * drop.r * math.Hypot(float64(a.width), float64(a.height)) * a.intensity
		margin := radius + 2
		minX := maxInt(0, int(cx-margin))
		maxX := minInt(a.width-1, int(cx+margin)+1)
		minY := maxInt(0, int(cy-margin/1.8))
		maxY := minInt(a.height-1, int(cy+margin/1.8)+1)
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				dx := float64(x) - cx
				dy := (float64(y) - cy) * 1.8
				edge := radius + math.Sin(float64(x)*0.45+float64(y)*0.31+t*8)*2
				dist := math.Hypot(dx, dy)
				if dist > edge {
					continue
				}
				ch := '#'
				color := ColorBrightBlack
				if math.Abs(dist-edge) < 2.2 {
					ch = '@'
					color = ColorWhite
				} else if pseudoNoise(x, y, t) > 0.72 {
					ch = '%'
				}
				f.Set(x, y, Cell{Ch: ch, Color: color, Bold: ch == '@'})
			}
		}
	}
	return f
}

var inkDrops = [...]struct {
	x float64
	y float64
	r float64
}{
	{0.5, 0.5, 0.42},
	{0.2, 0.35, 0.28},
	{0.78, 0.62, 0.32},
	{0.35, 0.78, 0.24},
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
