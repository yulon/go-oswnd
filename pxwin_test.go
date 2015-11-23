package pxwin

import (
	"testing"
)

func TestMessageLoop(*testing.T) {
	Init()
	New("test")
	MessageLoop()
}
