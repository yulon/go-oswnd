package pxwin

import (
	"testing"
	"bytes"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		var p []byte
		w.On(EventPaint, func(param ...int){
			r := w.Rect()
			pn := r.Width * r.Height
			if len(p) / 4 != pn {
				p = bytes.Repeat([]byte{0, 102, 178, 255}, pn)
			}
			w.Paint(p)
		})
	})
}
