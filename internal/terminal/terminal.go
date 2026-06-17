package terminal

import (
	"io"
	"os"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/bsushmith/clearfx/internal/animation"
)

const (
	ClearScrollback = "\x1b[3J"
	ClearScreen = "\x1b[2J"
	Home        = "\x1b[H"
	HideCursor  = "\x1b[?25l"
	ShowCursor  = "\x1b[?25h"
	Reset       = "\x1b[0m"
)

var renderBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 16*1024)
		return &buf
	},
}

func IsTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func RenderFrame(w io.Writer, frame animation.Frame) error {
	bufPtr := renderBufferPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]
	defer func() {
		if cap(buf) <= 1<<20 {
			*bufPtr = buf
			renderBufferPool.Put(bufPtr)
		}
	}()

	buf = append(buf, Home...)
	var current animation.ANSIColor = -1
	currentBold := false
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			cell := frame.Cells[y*frame.Width+x]
			if cell.Ch == 0 {
				cell.Ch = ' '
			}
			if cell.Color != current || cell.Bold != currentBold {
				buf = append(buf, Reset...)
				if cell.Bold {
					buf = append(buf, "\x1b[1m"...)
				}
				if cell.Color != animation.ColorDefault {
					buf = append(buf, "\x1b["...)
					buf = strconv.AppendInt(buf, int64(cell.Color), 10)
					buf = append(buf, 'm')
				}
				current = cell.Color
				currentBold = cell.Bold
			}
			if cell.Ch < utf8.RuneSelf {
				buf = append(buf, byte(cell.Ch))
			} else {
				buf = utf8.AppendRune(buf, cell.Ch)
			}
		}
		if y < frame.Height-1 {
			buf = append(buf, "\r\n"...)
		}
	}
	buf = append(buf, Reset...)
	_, err := w.Write(buf)
	return err
}

func Clear(w io.Writer) error {
	_, err := io.WriteString(w, Reset+ClearScrollback+ClearScreen+Home)
	return err
}
