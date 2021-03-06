package oswnd

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	w := New()
	w.SetTitle("Hello Window!")
	w.SetClientSzie(Size{500, 500})
	w.MoveToScreenCenter()
	w.OnKeyDown = func(keyCode, count int){
		fmt.Println(keyCode, count)
	}
	w.Show()
	BlockAndHandleEvents()
}
