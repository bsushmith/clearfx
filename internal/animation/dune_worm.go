//go:build experimental

package animation

import "math"

type duneWormStyle struct{}

func init() { Register(duneWormStyle{}) }

func (duneWormStyle) Name() string { return "dune-worm" }
func (duneWormStyle) Description() string {
	return "a sandworm breaches through dunes with a trailing wake"
}
func (duneWormStyle) New(width, height int, opts Options) Animator {
	return duneWormAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity)}
}

type duneWormAnimator struct {
	width     int
	height    int
	intensity float64
}

func (a duneWormAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)

	horizon := int(float64(a.height) * 0.38)
	for y := horizon; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			if (x+y)%13 == 0 {
				f.Set(x, y, Cell{Ch: '.', Color: ColorYellow})
			}
			if y > horizon && math.Sin(float64(x)*0.16+float64(y)*0.45) > 0.82 {
				f.Set(x, y, Cell{Ch: '-', Color: ColorYellow})
			}
		}
	}

	headX := int(-float64(a.width)*0.18 + t*float64(a.width)*1.45)
	centerY := int(float64(a.height)*0.58 + math.Sin(t*math.Pi*2.2)*float64(a.height)*0.12)
	segments := 28
	spacing := math.Max(2.4, float64(a.width)/42)

	for i := segments - 1; i >= 0; i-- {
		progress := float64(i) / float64(segments-1)
		x := float64(headX) - float64(i)*spacing
		bodyWave := math.Sin(progress*math.Pi*3.4 + t*math.Pi*4)
		y := float64(centerY) + bodyWave*float64(a.height)*0.11
		breach := math.Sin(progress*math.Pi*2.8 + t*math.Pi*5)
		if breach < -0.12 && i > 4 {
			drawSubsurfaceWake(f, int(x), int(y), i)
			continue
		}

		rx := int(math.Max(2, (4.8-progress*2.3)*a.intensity))
		ry := int(math.Max(1, (2.2-progress*1.0)*a.intensity))
		if i == 0 {
			rx += 2
			ry += 1
		}
		drawWormSegment(f, int(x), int(y), rx, ry, i == 0, i)
		drawSandWake(f, int(x), int(y)+ry, i, t)
	}

	drawWormHead(f, headX, centerY, t)
	for i := 0; i < 18; i++ {
		x := headX - i*3
		y := centerY + int(math.Sin(float64(i)*0.55+t*math.Pi*5)*3)
		drawWakeRidges(f, x, y+3, i)
	}
	return f
}

func drawWormSegment(f Frame, cx, cy, rx, ry int, head bool, idx int) {
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			nx := float64(dx) / math.Max(1, float64(rx))
			ny := float64(dy) / math.Max(1, float64(ry))
			if nx*nx+ny*ny > 1 {
				continue
			}
			ch := '#'
			color := ColorYellow
			bold := false
			if dy < 0 {
				color = ColorBrightYellow
				bold = true
			}
			if dy == ry || (idx+dx+dy)%5 == 0 {
				ch = '~'
				color = ColorBrightBlack
			}
			if head {
				ch = '@'
				color = ColorBrightYellow
				bold = true
			}
			f.Set(cx+dx, cy+dy, Cell{Ch: ch, Color: color, Bold: bold})
		}
	}
}

func drawWormHead(f Frame, x, y int, t float64) {
	mouthOpen := 2 + int(math.Sin(t*math.Pi*8)*1.5)
	head := []struct {
		dx int
		dy int
		ch rune
	}{
		{0, -3, '^'}, {-1, -2, '/'}, {0, -2, '@'}, {1, -2, '\\'},
		{-2, -1, '<'}, {-1, -1, '#'}, {0, -1, '#'}, {1, -1, '#'}, {2, -1, '>'},
		{-2, 0, '<'}, {-1, 0, '#'}, {0, 0, 'O'}, {1, 0, '#'}, {2, 0, '>'},
		{-1, 1, '\\'}, {0, 1 + mouthOpen/3, 'v'}, {1, 1, '/'},
	}
	for _, p := range head {
		color := ColorBrightYellow
		if p.ch == 'O' {
			color = ColorBrightBlack
		}
		f.Set(x+p.dx, y+p.dy, Cell{Ch: p.ch, Color: color, Bold: true})
	}
}

func drawSandWake(f Frame, x, y, idx int, t float64) {
	for i := 0; i < 5; i++ {
		dx := i + 1
		lift := int(math.Sin(float64(idx+i)+t*math.Pi*7) * 1.5)
		cell := Cell{Ch: '.', Color: ColorBrightYellow}
		if i%2 == 0 {
			cell.Ch = '*'
		}
		f.Set(x-dx, y+lift, cell)
		f.Set(x+dx/2, y+lift+1, Cell{Ch: '.', Color: ColorYellow})
	}
}

func drawSubsurfaceWake(f Frame, x, y, idx int) {
	for i := -3; i <= 3; i++ {
		if (idx+i)%2 == 0 {
			f.Set(x+i, y, Cell{Ch: '~', Color: ColorYellow})
		}
	}
}

func drawWakeRidges(f Frame, x, y, idx int) {
	for j := 0; j < 7; j++ {
		ch := '-'
		if j%3 == 0 {
			ch = '.'
		}
		f.Set(x-j, y+j/3, Cell{Ch: ch, Color: ColorYellow})
		f.Set(x-j, y-j/3, Cell{Ch: ch, Color: ColorYellow})
	}
	if idx%4 == 0 {
		f.Set(x, y-2, Cell{Ch: '*', Color: ColorBrightYellow, Bold: true})
	}
}
