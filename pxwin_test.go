package pxwin

import (
	"testing"
)

func TestWindow(*testing.T) {
	Init()
	win := New()
	win.SetEventListener(func(event int, param1 int, param2 int) {
		println(event)
	})
	EventDrive()
}
