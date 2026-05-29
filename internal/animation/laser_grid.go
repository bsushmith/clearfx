//go:build experimental

package animation

import "math"

type laserGridStyle struct{}

func init() { Register(laserGridStyle{}) }

func (laserGridStyle) Name() string { return "laser-grid" }
func (laserGridStyle) Description() string {
	return "scanning laser lines erase the screen in a grid"
}
func (laserGridStyle) New(width, height int, opts Options) Animator {
	return laserGridAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type laserGridAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a laserGridAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	vx := int(t * float64(a.width+a.height))
	hy := int(t * float64(a.height+a.width/3))
	spacing := int(math.Max(4, 9/a.intensity))
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			grid := x%spacing == 0 || y%(spacing/2+1) == 0
			scan := absInt(x-vx) < 2 || absInt(y-hy) < 1 || absInt((x+y)-vx) < 2
			if !grid && !scan {
				continue
			}
			cell := Cell{Ch: '.', Color: ColorBlue}
			if grid {
				cell = Cell{Ch: '+', Color: ColorCyan}
			}
			if scan {
				cell = Cell{Ch: '#', Color: ColorBrightWhite, Bold: true}
			}
			f.Set(x, y, cell)
		}
	}
	return f
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
