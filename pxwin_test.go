package pxwin

import (
	"testing"
)

func TestWindow(*testing.T) {
	Init()
	win := New()
	win.EventListener = func(e int, a int, b int) {
		println(e)
	}
	EventDrive()
}
