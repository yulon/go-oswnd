package pxwin

import (
	"testing"
	//"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		p := make([]byte, 500 * 500 * 4)
		for i := 0; i < len(p); i++ {
			p[i] = 255
		}
		w.On(EventPaint, func(param ...int){
			w.Paint(p)
		})
	})
}
