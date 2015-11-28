package oswnd

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		w.SetTitle("Hello Window!")
		w.OnKeyDown = func(keyCode, count int){
			fmt.Println(keyCode, count)
		}
	})
}
