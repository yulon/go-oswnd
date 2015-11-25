package pxwin

import (
	"testing"
	"bytes"
	//"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		w.On(EventPaint, func(param ...int){
			r := w.Rect()
			w.Paint(bytes.Repeat([]byte{155}, r.Width * r.Height * 4))
		})
	})
}
