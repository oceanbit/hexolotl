package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

func drawText(s tcell.Screen, x1, y1 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
	}
}

func main() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)

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

	startLine := 0
	text := [6]string{"Testing line 1", "Testing line 2", "Testing line 3", "Testing line 4", "Testing line 5", "Testing line 6"}

	redraw := func() {
		s.Clear()
		for i := 0; i < 2 && (startLine + i) < len(text); i++ {
			fmt.Println(i)
			drawText(s, 0, i, boxStyle, text[startLine + i])
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
				startLine -= 1
				if startLine < 0 {
					startLine = 0
				}
				redraw()
			} else if ev.Key() == tcell.KeyDown {
				startLine += 1
				if startLine >= len(text) {
					startLine = len(text) - 1
				}
				redraw()
			}
		}
	}
}