package pxwin

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		var i int
		w.On(EventKeyDown, func(param ...int){
			i += 10
			fmt.Println(param)
			w.Move(&Rect{i, i, 300, 300})
		})
	})
}
