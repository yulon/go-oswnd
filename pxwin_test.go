package pxwin

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Init()
	win := New()
	win.SetEventListener(func(event int, param ...int) {
		fmt.Println(event, param)
	})
	EventDrive()
}
