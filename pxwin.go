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
	ClientRect() *Rect
	Move(*Rect)
	On(event int, eh EventHandler)
	Paint(pixels []byte)
}

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}

type EventHandler func(param ...int)
