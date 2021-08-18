package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

//func drawText(s tcell.Screen, x1, y1 int, style tcell.Style, text string) {
//	row := y1
//	col := x1
//	for _, r := range []rune(text) {
//		s.SetContent(col, row, r, nil, style)
//		col++
//	}
//}

func main() {
	// TODO: Replace this with file selector
	dat, e := os.Open("demo.txt")
	bytes := make([]byte, 500)
	numOfBytes, err := dat.Read(bytes)

	if e != nil {
		panic(e)
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.Clear()

	currentRow := 0
	bytesPerLine := 20
	var maxRowIndex int = numOfBytes / bytesPerLine

	redraw := func() {
		s.Clear()

		for row := 0; row < 2; row++ {
			rowIndex := currentRow + row
			for col := 0; col < bytesPerLine; col++ {
				byteIndex := (rowIndex * bytesPerLine) + col
				if byteIndex >= numOfBytes {
					break
				}
				b := bytes[byteIndex]
				s.SetContent(col, row, rune(b), nil, textStyle)
			}
		}
		s.Show()
	}

	redraw()

	// Event loop
	quit := func() {
		s.Fini()
		os.Exit(0)
	}
	for {
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				quit()
			} else if ev.Key() == tcell.KeyUp {
				currentRow -= 1
				if currentRow < 0 {
					currentRow = 0
				}
				redraw()
			} else if ev.Key() == tcell.KeyDown {
				currentRow += 1
				if currentRow >= maxRowIndex {
					currentRow = maxRowIndex
				}
				redraw()
			}
		}
	}
}