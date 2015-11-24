package pxwin

const (
	EventPaint = iota
	EventKeyDown
	EventKeyUp
)

type Window interface{
	GetTitle() string
	SetTitle(title string)
	Rect() *Rect
	Move(*Rect)
	ListenEvent(event int, eh EventHandler)
}

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}

type EventHandler func(param ...int)
