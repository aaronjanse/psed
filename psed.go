package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pkg/term"
)

const (
	WORD int = 1
)

var mode = WORD

func getch() []byte {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)
	numRead, err := t.Read(bytes)
	t.Restore()
	t.Close()
	if err != nil {
		return nil
	}
	return bytes[0:numRead]
}

type Cursor struct {
	x int
	y int
}

var lines = []string{""}
var cursor = Cursor{0, 0}

func init() {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 { // data is being piped from stdin
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		lines = strings.Split(string(bytes), "\n")
		for _, line := range lines {
			fmt.Fprintln(os.Stderr, line)
		}
		fmt.Fprintf(os.Stderr, "\033[%vF\033[0G", len(lines))
	}
}

func main() {
	defer func() {
		fmt.Println(strings.Join(lines, "\n"))
	}()
	for {
		printMaskLine()
		c := getch()
		switch {
		case bytes.Equal(c, []byte{3}):
			return
		case bytes.Equal(c, []byte{4}):
			return
		case bytes.Equal(c, []byte{13}): // newline
			maskEntireLine()
			lines = append(lines, "")
			cursor.x = 0
			cursor.y++
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprint(os.Stderr, " \033[1D")
		case bytes.Equal(c, []byte{127}): // backspace
			if cursor.x > 0 {
				lines[cursor.y] = lines[cursor.y][:cursor.x-1] + lines[cursor.y][cursor.x:]
				fmt.Fprintf(os.Stderr, "\r%s", lines[cursor.y]+" ")
				fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x)
				cursor.x--
			} else if cursor.y > 0 {
				oldLen := len(lines[cursor.y-1])
				lines[cursor.y-1] += lines[cursor.y]
				lines = append(lines[:cursor.y], lines[cursor.y+1:]...) // remove line
				fmt.Fprintf(os.Stderr, "\033[%vF\r", len(lines))
				for _, line := range lines {
					fmt.Fprintln(os.Stderr, line)
				}
				fmt.Fprint(os.Stderr, "\033[36m~\033[39m"+strings.Repeat(" ", len(lines[len(lines)-1])))
				fmt.Fprintf(os.Stderr, "\033[%vF\033[%vG", len(lines)-cursor.y+1, oldLen+1)
				cursor.x = oldLen
				cursor.y--
			}
		case bytes.Equal(c, []byte{27, 91, 68}): // left
			if cursor.x > 0 {
				cursor.x--
				fmt.Fprint(os.Stderr, "\033[1D")
			} else {
				if cursor.y > 0 {
					maskEntireLine()
					cursor.y--
					cursor.x = len(lines[cursor.y])
					fmt.Fprintf(os.Stderr, "\033[1A\033[%vG", cursor.x+1)
				}
			}
		case bytes.Equal(c, []byte{27, 91, 67}): // right
			if !(cursor.y >= len(lines)-1 && cursor.x > len(lines[cursor.y])-1) {
				cursor.x++
				if cursor.x > len(lines[cursor.y]) {
					if cursor.y < len(lines)-1 {
						maskEntireLine()
						cursor.x = 0
						cursor.y++
						fmt.Fprint(os.Stderr, "\033[1E")
					}
				} else {
					fmt.Fprint(os.Stderr, "\033[1C")
				}
			}
		case bytes.Equal(c, []byte{27, 91, 65}): // up
			if cursor.y > 0 {
				maskEntireLine()
				cursor.y--
				fmt.Fprint(os.Stderr, "\033[1A")
			}
			if cursor.x > len(lines[cursor.y]) {
				cursor.x = len(lines[cursor.y])
				fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
			}
		case bytes.Equal(c, []byte{27, 91, 66}): // down
			if cursor.y < len(lines)-1 {
				maskEntireLine()
				cursor.y++
				fmt.Fprint(os.Stderr, "\033[1B")
			}
			if cursor.x >= len(lines[cursor.y]) {
				cursor.x = len(lines[cursor.y])
				fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
			}
		case bytes.Compare(c, []byte{32}) > 0 && bytes.Compare(c, []byte{127}) <= 0: // printable chars
			cursor.x++
			lines[cursor.y] += string(c)
			fmt.Fprint(os.Stderr, string(c))
		case bytes.Equal(c, []byte{32}):
			if mode == WORD {
				line := lines[cursor.y]
				words := strings.Split(line, " ")
				lastWord := words[len(words)-1]
				runeCount := utf8.RuneCountInString(lastWord)
				fmt.Fprintf(os.Stderr, "\033[%vD", runeCount)          // move cursor back
				fmt.Fprintf(os.Stderr, strings.Repeat("*", runeCount)) // replace word with astricks
			}

			lines[cursor.y] += " "
			cursor.x++
			fmt.Fprint(os.Stderr, " ")
		case bytes.Equal(c, []byte{9}):
			fmt.Fprint(os.Stderr, "\t")
		default:
			// fmt.Fprintln(os.Stderr, "")
			// fmt.Fprintln(os.Stderr, c)
		}
	}
}

func printMaskLine() {
	line := lines[cursor.y]
	if mode == WORD {
		maskedLine := ""
		currentX := 0
		words := strings.Split(line, " ")
		for _, word := range words {
			runeCount := utf8.RuneCountInString(word)
			if currentX <= cursor.x && cursor.x <= currentX+runeCount {
				maskedLine += word
			} else {
				maskedLine += strings.Repeat("*", runeCount)
			}
			maskedLine += " "
			currentX += runeCount + 1
		}
		line = maskedLine
	}
	fmt.Fprintf(os.Stderr, "\r%s", line+" ")
	fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
}

func maskEntireLine() {
	if mode == WORD {
		maskedLine := ""
		words := strings.Split(lines[cursor.y], " ")
		for _, word := range words {
			runeCount := utf8.RuneCountInString(word)
			maskedLine += strings.Repeat("*", runeCount) + " "
		}
		fmt.Fprintf(os.Stderr, "\r%s", maskedLine+" ")
	}
}
