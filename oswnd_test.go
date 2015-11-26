package oswnd

import (
	"testing"
	"fmt"
)

func TestWindow(*testing.T) {
	Main(func() {
		w := New()
		w.SetTitle("Hello Window!")
		w.On(EventKeyDown, func(param ...int){
			fmt.Println(param)
		})
	})
}
