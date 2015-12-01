package oswnd

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		w.SetTitle("Hello Window!")
		w.MoveToScreenCenter()
		w.OnKeyDown = func(keyCode, count int){
			fmt.Println(keyCode, count)
		}
	})
}
