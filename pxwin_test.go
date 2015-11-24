package pxwin

import (
	"testing"
)

func TestMessageLoop(*testing.T) {
	Init()
	win := New("test")
	win.EventListener = func(e int, a int, b int) {
		println(e)
	}
	EventDrive()
}
