package ansi

import (
	"fmt"
	"os"
)

func ClearScreen() {
	print("\033[2J")
}

// SetCursor moves the cursor to the given absolute coordinates
func SetCursor(x, y int) {
	printf("\033[%v;%vH", x+1, y+1)
}

func SetCursorX(a int) {
	printf("\033[%vG", a+1)
}

func MoveCursorUp(a int) {
	printf("\033[%vA", a)
}

func MoveCursorDown(a int) {
	printf("\033[%vB", a)
}

func MoveCursorDownLinear(a int) {
	printf("\033[%vE", a)
}

func print(s string) {
	fmt.Fprint(os.Stderr, s)
}

func printf(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}
