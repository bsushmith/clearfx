package animation

import "math"

type oceanWaveStyle struct{}
type greatWaveStyle struct{}

func init() {
	Register(oceanWaveStyle{})
	Register(greatWaveStyle{})
}

func (oceanWaveStyle) Name() string { return "ocean-wave" }
func (oceanWaveStyle) Description() string {
	return "moving ocean swells travel left to right across the terminal"
}
func (oceanWaveStyle) New(width, height int, opts Options) Animator {
	return oceanWaveAnimator{
		width:     width,
		height:    height,
		intensity: intensityScale(opts.Intensity),
		palette:   PaletteFor(opts.Palette),
	}
}

func (greatWaveStyle) Name() string { return "great-wave" }
func (greatWaveStyle) Description() string {
	return "a Great Wave-inspired curl crashes across the terminal"
}
func (greatWaveStyle) New(width, height int, opts Options) Animator {
	return &greatWaveAnimator{
		width:     width,
		height:    height,
		intensity: intensityScale(opts.Intensity),
		surfaces:  make([]int, width),
		palette:   PaletteFor(opts.Palette),
	}
}

type oceanWaveAnimator struct {
	width     int
	height    int
	intensity float64
	palette   Palette
}

func (a oceanWaveAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)

	base := int(float64(a.height) * 0.42)
	amp := math.Max(2, float64(a.height)*0.13*a.intensity)
	for x := 0; x < a.width; x++ {
		phaseX := float64(x)
		surface := float64(base) +
			math.Sin(phaseX*0.22-t*math.Pi*5.2)*amp +
			math.Sin(phaseX*0.09-t*math.Pi*2.4)*amp*0.45
		top := int(surface)

		for y := top; y < a.height; y++ {
			depth := y - top
			f.Set(x, y, a.movingWaterCell(depth, x, y, t))
		}

		crest := math.Sin(phaseX*0.22 - t*math.Pi*5.2)
		if crest > 0.72 {
			f.Set(x, top-1, Cell{Ch: '^', Color: a.palette.Primary, Bold: true})
			if crest > 0.9 {
				f.Set(x, top-2, Cell{Ch: '\'', Color: a.palette.Primary, Bold: true})
			}
		}
	}

	for band := 0; band < 4; band++ {
		yBase := a.height - 2 - band*3
		speed := t * math.Pi * (5.5 + float64(band))
		for x := 0; x < a.width; x++ {
			phase := float64(x)*0.34 - speed + float64(band)*0.8
			if math.Sin(phase) > 0.4 {
				y := yBase + int(math.Sin(float64(x)*0.12-t*math.Pi*2)*1.5)
				f.Set(x, y, Cell{Ch: '~', Color: a.palette.Secondary, Bold: band == 0})
			}
			if math.Sin(phase) > 0.86 {
				y := yBase - 1
				f.Set(x, y, Cell{Ch: '.', Color: a.palette.Primary, Bold: true})
			}
		}
	}

	return f
}

type greatWaveAnimator struct {
	width     int
	height    int
	intensity float64
	surfaces  []int
	palette   Palette
}

func (a *greatWaveAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)

	base := float64(a.height) * 0.64
	scale := math.Max(10, float64(a.width)*0.24)
	height := math.Max(5, float64(a.height)*0.48*a.intensity)
	crestX := float64(a.width)*(-0.18+1.18*t) + math.Sin(t*math.Pi*2)*2
	surfaces := a.surfaces

	for x := 0; x < a.width; x++ {
		u := (float64(x) - crestX) / scale
		shoulder := math.Exp(-u * u * 1.25)
		backSwell := math.Exp(-(u+0.95)*(u+0.95)*0.85) * 0.55
		frontDrop := smoothstep(0.0, 1.05, u) * height * 0.48
		travelRipple := math.Sin(float64(x)*0.18-t*math.Pi*4.4) * 1.4
		surface := base - height*shoulder - height*backSwell + frontDrop + travelRipple
		top := int(surface)
		surfaces[x] = top
		for y := top; y < a.height; y++ {
			depth := y - top
			f.Set(x, y, a.greatWaveWaterCell(depth, x, y, t))
		}
	}

	// Draw the breaker surface as connected slope characters, so the wave reads
	// as one rolling body instead of a geometric arc.
	for x := 1; x < a.width-1; x++ {
		y := surfaces[x]
		slope := surfaces[x+1] - surfaces[x-1]
		ch := '~'
		if slope <= -2 {
			ch = '/'
		} else if slope >= 2 {
			ch = '\\'
		} else if math.Sin(float64(x)*0.45-t*math.Pi*8) > 0.72 {
			ch = '^'
		}
		f.Set(x, y, Cell{Ch: ch, Color: a.palette.Primary, Bold: true})
		f.Set(x, y+1, Cell{Ch: '~', Color: a.palette.Secondary, Bold: true})
	}

	crestY := base - height*1.24
	lipLength := scale * 0.78
	for i := 0; i < int(lipLength); i++ {
		s := float64(i) / lipLength
		x := int(crestX + s*lipLength)
		y := int(crestY + s*s*height*0.86 + math.Sin(s*math.Pi*2+t*math.Pi*3)*1.3)
		f.Set(x, y, Cell{Ch: '~', Color: a.palette.Primary, Bold: true})
		f.Set(x-1, y, Cell{Ch: '^', Color: a.palette.Primary, Bold: true})
		f.Set(x, y+1, Cell{Ch: '~', Color: a.palette.Secondary, Bold: true})

		if i%3 == 0 {
			drop := 2 + int(s*height*0.22)
			for j := 1; j <= drop; j++ {
				if (i+j)%2 == 0 {
					f.Set(x+j/2, y+j, Cell{Ch: '.', Color: a.palette.Primary, Bold: true})
				}
			}
		}
	}

	// Wind-blown foam fingers trail from the lip and move with the breaker.
	for finger := 0; finger < 9; finger++ {
		start := float64(finger) / 8
		startX := crestX + start*lipLength
		startY := crestY + start*start*height*0.78
		length := int(5 + float64(finger%4)*2*a.intensity)
		for j := 0; j < length; j++ {
			x := int(startX + float64(j)*1.45 + math.Sin(t*math.Pi*4+float64(finger))*1.5)
			y := int(startY + float64(j)*0.55 + math.Sin(float64(j)*0.9+float64(finger))*1.2)
			ch := '\''
			if j%4 == 0 {
				ch = '.'
			}
			f.Set(x, y, Cell{Ch: ch, Color: a.palette.Primary, Bold: true})
		}
	}

	for band := 0; band < 4; band++ {
		y := a.height - 2 - band*3
		for x := 0; x < a.width; x++ {
			phase := float64(x)*0.34 - t*math.Pi*(4.5+float64(band)*0.45) + float64(band)*0.9
			if math.Sin(phase) > 0.3 {
				f.Set(x, y, Cell{Ch: '~', Color: a.palette.Secondary, Bold: band == 0})
			}
			if math.Sin(phase) > 0.88 {
				f.Set(x, y-1, Cell{Ch: '^', Color: a.palette.Primary, Bold: true})
			}
		}
	}

	return f
}

func smoothstep(edge0, edge1, x float64) float64 {
	if x <= edge0 {
		return 0
	}
	if x >= edge1 {
		return 1
	}
	v := (x - edge0) / (edge1 - edge0)
	return v * v * (3 - 2*v)
}

func (a oceanWaveAnimator) movingWaterCell(depth, x, y int, t float64) Cell {
	flow := math.Sin(float64(x)*0.5 - t*math.Pi*8 + float64(y)*0.23)
	if depth == 0 {
		return Cell{Ch: '~', Color: a.palette.Secondary, Bold: true}
	}
	if depth < 3 {
		if flow > 0.55 {
			return Cell{Ch: '.', Color: a.palette.Primary, Bold: true}
		}
		return Cell{Ch: '~', Color: a.palette.Secondary, Bold: true}
	}
	if depth < 8 {
		if flow > 0.65 {
			return Cell{Ch: '-', Color: a.palette.Cool}
		}
		return Cell{Ch: '=', Color: a.palette.Cool}
	}
	if flow > 0.72 {
		return Cell{Ch: '~', Color: a.palette.Cool}
	}
	return Cell{Ch: '~', Color: a.palette.Accent}
}

func waterCell(depth, x, y int, palette Palette) Cell {
	if depth == 0 {
		return Cell{Ch: '^', Color: palette.Primary, Bold: true}
	}
	if depth < 3 {
		return Cell{Ch: '~', Color: palette.Secondary, Bold: true}
	}
	if (x+y)%9 == 0 {
		return Cell{Ch: '.', Color: palette.Cool}
	}
	if depth < 8 {
		return Cell{Ch: '=', Color: palette.Cool}
	}
	return Cell{Ch: '~', Color: palette.Accent}
}

func (a greatWaveAnimator) greatWaveWaterCell(depth, x, y int, t float64) Cell {
	flow := math.Sin(float64(x)*0.42 - t*math.Pi*7 + float64(y)*0.18)
	if depth == 0 {
		return Cell{Ch: '~', Color: a.palette.Primary, Bold: true}
	}
	if depth < 3 {
		if flow > 0.48 {
			return Cell{Ch: '.', Color: a.palette.Primary, Bold: true}
		}
		return Cell{Ch: '~', Color: a.palette.Secondary, Bold: true}
	}
	if depth < 8 {
		if flow > 0.6 {
			return Cell{Ch: '-', Color: a.palette.Cool}
		}
		return Cell{Ch: '=', Color: a.palette.Cool}
	}
	return Cell{Ch: '~', Color: a.palette.Accent}
}
