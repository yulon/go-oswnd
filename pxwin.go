package pxwin

const (
	EventPaint = iota
	EventKeyDown
	EventKeyUp
)

type Window struct{
	EventListener func(int, int, int)
}
