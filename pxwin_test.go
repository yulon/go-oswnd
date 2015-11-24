package pxwin

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Init()
	w := New()
	w.ListenEvent(EventKeyDown, func(param ...int){
		fmt.Println(param)
	})
	EventDrive()
}
