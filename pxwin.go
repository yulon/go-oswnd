package pxwin

const (
	EventPaint = iota
	EventKeyDown
	EventKeyUp
)

type Window interface{
	SetEventListener(eventHandler func(event int, param1 int, param2 int))
}
