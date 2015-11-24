package pxwin

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		w.ListenEvent(EventKeyDown, func(param ...int){
			fmt.Println(param)
		})
	})
}
