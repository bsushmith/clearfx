package animation

import "math"

type matrixRainStyle struct{}

func init() { Register(matrixRainStyle{}) }

func (matrixRainStyle) Name() string { return "matrix-rain" }
func (matrixRainStyle) Description() string {
	return "green code rain falls down and wipes the terminal"
}
func (matrixRainStyle) New(width, height int, opts Options) Animator {
	return matrixRainAnimator{width: width, height: height, intensity: intensityScale(opts.Intensity), palette: PaletteFor(opts.Palette)}
}

type matrixRainAnimator struct {
	width     int
	height    int
	intensity float64
	palette   Palette
}

func (a matrixRainAnimator) Frame(t float64) Frame {
	t = clamp01(t)
	f := NewFrame(a.width, a.height)
	chars := []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ$#@%&")
	density := int(math.Max(2, 5/a.intensity))
	for x := 0; x < a.width; x++ {
		if x%density != 0 && (x+int(t*10))%(density+1) != 0 {
			continue
		}
		speed := 0.65 + float64((x*17)%9)/10
		head := int((t*speed*float64(a.height*2) + float64((x*7)%a.height))) % (a.height + 12)
		trail := 6 + int(6*a.intensity)
		for i := 0; i < trail; i++ {
			y := head - i
			if y < 0 || y >= a.height {
				continue
			}
			idx := (x*31 + y*13 + int(t*40)) % len(chars)
			cell := Cell{Ch: chars[idx], Color: a.palette.Primary}
			if i == 0 {
				cell = Cell{Ch: chars[idx], Color: a.palette.Highlight, Bold: true}
			} else if i < 3 {
				cell.Color = a.palette.Secondary
				cell.Bold = true
			}
			f.Set(x, y, cell)
		}
	}
	return f
}
