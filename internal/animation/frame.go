package animation

type ANSIColor int16

const (
	ColorDefault ANSIColor = 0
	ColorBlack   ANSIColor = 30
	ColorRed     ANSIColor = 31
	ColorGreen   ANSIColor = 32
	ColorYellow  ANSIColor = 33
	ColorBlue    ANSIColor = 34
	ColorMagenta ANSIColor = 35
	ColorCyan    ANSIColor = 36
	ColorWhite   ANSIColor = 37

	ColorBrightBlack   ANSIColor = 90
	ColorBrightRed     ANSIColor = 91
	ColorBrightGreen   ANSIColor = 92
	ColorBrightYellow  ANSIColor = 93
	ColorBrightBlue    ANSIColor = 94
	ColorBrightMagenta ANSIColor = 95
	ColorBrightCyan    ANSIColor = 96
	ColorBrightWhite   ANSIColor = 97
)

type Cell struct {
	Ch    rune
	Color ANSIColor
	Bold  bool
}

type Frame struct {
	Width  int
	Height int
	Cells  []Cell
}

func NewFrame(width, height int) Frame {
	return Frame{Width: width, Height: height, Cells: make([]Cell, width*height)}
}

func (f Frame) Set(x, y int, c Cell) {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return
	}
	f.Cells[y*f.Width+x] = c
}
