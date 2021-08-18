package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/reactivex/rxgo/v2"
	"io"
	"log"
	//"math/bits"
	"os"
)

func check(e error) {
	if e == io.EOF {
		return
	}
	if e != nil {
		panic(e)
	}
}
type OnChangeFn func([]byte, int64, int64)

func readAndObserveFile(
	bytesPerRow int64,
	numberOfRows int,
	onChange OnChangeFn,
) (chan rxgo.Item, *os.File) {
	fileSeekChannel := make(chan rxgo.Item)

	fileSeekObservable := rxgo.FromChannel(fileSeekChannel)

	var currentFileChunkIndex int64 = 0

	// TODO: Replace this with file selector
	activeFile, openErr := os.Open("demo.txt")
	check(openErr)

	byteChunk := make([]byte, bytesPerRow * int64(numberOfRows))
	byteChunkSize := 0

	readFileUpdateChunk := func(offset int64) {
		var err error
		byteChunkSize, err = activeFile.ReadAt(byteChunk, 0)
		check(err)
	}

	readFileUpdateChunk(0)

	// Get file size
	fileStat, statE := activeFile.Stat()
	fileSize := fileStat.Size()
	check(statE)

	// Eager initialize UI
	onChange(
		byteChunk,
		currentFileChunkIndex,
		fileSize,
	)

	// Update UI on observable update
	fileSeekObservable.DoOnNext(func (i interface {}) {
		if i == "up" {
			currentFileChunkIndex -= 1
			if currentFileChunkIndex < 0 {
				currentFileChunkIndex = 0
			}
		} else if i == "down" {
			currentFileChunkIndex += 1
			maxRowIndex := fileSize / bytesPerRow
			if currentFileChunkIndex >= maxRowIndex {
				currentFileChunkIndex = maxRowIndex
			}
		}
		readFileUpdateChunk(currentFileChunkIndex * bytesPerRow)
		onChange(
			byteChunk,
			currentFileChunkIndex,
			fileSize,
		)
	})

	return fileSeekChannel, activeFile
}

func main() {
	var bytesPerRow int64 = 120
	numberOfRows := 3

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.Clear()

	var redraw OnChangeFn = func(byteChunk []byte, currentFileChunkIndex int64, fileSize int64) {
		s.Clear()
		for row := 0; row < numberOfRows; row++ {
			rowIndex := currentFileChunkIndex + int64(row)
			for col := 0; int64(col) < bytesPerRow; col++ {
				byteTotalIndex := (rowIndex * bytesPerRow) + int64(col)
				byteLocalIndex := (int64(row) * bytesPerRow) + int64(col)
				if byteTotalIndex >= fileSize {
					break
				}
				b := byteChunk[byteLocalIndex]
				s.SetContent(col, row, rune(b), nil, textStyle)
			}
		}
		s.Show()
	}

	fileSeekChannel, activeFile := readAndObserveFile(
		bytesPerRow,
		numberOfRows,
		redraw,
	)

	quit := func() {
		s.Fini()
		closeErr := activeFile.Close()
		check(closeErr)
		os.Exit(0)
	}
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				quit()
			} else if ev.Key() == tcell.KeyUp {
				fileSeekChannel <- rxgo.Of("up")
			} else if ev.Key() == tcell.KeyDown {
				fileSeekChannel <- rxgo.Of("down")
			}
		}
	}
}