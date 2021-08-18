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
	fileName string,
	bytesPerRow int64,
	numberOfRows int,
	onChange OnChangeFn,
) (chan rxgo.Item, *os.File) {
	fileSeekChannel := make(chan rxgo.Item)

	fileSeekObservable := rxgo.FromChannel(fileSeekChannel)

	var currentFileChunkIndex int64 = 0

	// TODO: Replace this with file selector
	activeFile, openErr := os.Open(fileName)
	check(openErr)

	byteChunk := make([]byte, bytesPerRow * int64(numberOfRows))
	byteChunkSize := 0

	readFileUpdateChunk := func(offset int64) {
		var err error
		byteChunkSize, err = activeFile.ReadAt(byteChunk, offset)
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

func simpleFileSelector(s tcell.Screen, style tcell.Style, quit func()) string {
	fileName := ""

	//File input validation
	for fileName == "" {
		attemptedFileName := textPrompt("Enter a file to analyze: ", s, style, quit)
		_, statErr := os.Stat(attemptedFileName)
		if os.IsNotExist(statErr) {
			writeText(
				"Invalid filepath, please input a valid filepath. Hit enter to try again",
				0,
				s,
				style,
			)
			s.Sync()
			waitForEnter(
				s,
				quit,
				func(ev tcell.Event) {},
			)
			continue
		}
		check(statErr)
		fileName = attemptedFileName
		break
	}

	return fileName
}
func writeText(text string, y int, s tcell.Screen, style tcell.Style) {
	for i, promptChar := range text {
		s.SetContent(i, y, promptChar, nil, style)
	}
}

func waitForEnter(s tcell.Screen, quit func(), evFn func(ev tcell.Event)) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				quit()
			} else if ev.Key() == tcell.KeyEnter {
				s.Clear()
				return
			}
		}
		evFn(ev)
	}
}

func textPrompt(promptText string, s tcell.Screen, style tcell.Style, quit func()) string {
	writeText(promptText, 0, s, style)
	s.Sync()

	var output string = ""
	colIndex := 0
	waitForEnter(s, quit, func(ev tcell.Event) {
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyBackspace {
				s.SetContent(colIndex, 1, rune(" "[0]), nil, style)
				s.Sync()
				if len(output) > 0 && colIndex > 0 {
					output = output[:len(output) - 1]
					colIndex--
				}
			} else {
				char := ev.Rune()
				s.SetContent(colIndex, 1, char, nil, style)
				output += string(char)
				s.Sync()
				colIndex++
			}
		}
	})
	s.Clear()
	return output
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

	quit := func() {
		s.Fini()
		//closeErr := activeFile.Close()
		//check(closeErr)
		os.Exit(0)
	}

	fileName := simpleFileSelector(s, textStyle, quit)

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

	fileSeekChannel, _ := readAndObserveFile(
		fileName,
		bytesPerRow,
		numberOfRows,
		redraw,
	)

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