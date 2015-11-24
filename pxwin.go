package pxwin

const (
	EventPaint = iota
	EventKeyDown
	EventKeyUp
)

type Window interface{
	GetTitle() string
	SetTitle(title string)
	GetRect() *Rect
	SetEventListener(eventListener func(event int, param ...int))
}

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}
