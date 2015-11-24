package pxwin

const (
	EventPaint = iota
	EventKeyDown
	EventKeyUp
)

type Window interface{
	SetEventListener(eventListener func(event int, param1 int, param2 int))
}
