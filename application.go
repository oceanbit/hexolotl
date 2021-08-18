package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var bytesPerLine int64 = 20
	numberOfRows := 3

	// TODO: Replace this with file selector
	dat, e := os.Open("demo.txt")
	check(e)
	bytes := make([]byte, bytesPerLine * int64(numberOfRows))
	numOfBytes, err := dat.ReadAt(bytes, 0)
	check(err)
	fileStat, statE := dat.Stat()
	fileSize := fileStat.Size()
	check(statE)

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

	var currentRow int64 = 0

	redraw := func() {
		numOfOffsetBytes, err := dat.ReadAt(bytes, currentRow * bytesPerLine)
		numOfBytes = numOfOffsetBytes
		check(err)

		s.Clear()
		for row := 0; row < numberOfRows; row++ {
			rowIndex := currentRow + int64(row)
			for col := 0; int64(col) < bytesPerLine; col++ {
				byteTotalIndex := (rowIndex * bytesPerLine) + int64(col)
				byteLocalIndex := (int64(row) * bytesPerLine) + int64(col)
				if byteTotalIndex >= fileSize {
					break
				}
				b := bytes[byteLocalIndex]
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
				maxRowIndex := fileSize / bytesPerLine
				if currentRow >= maxRowIndex {
					currentRow = maxRowIndex
				}
				redraw()
			}
		}
	}
}