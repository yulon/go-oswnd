package oswnd

const (
	EventPaint = iota
	EventKeyDown
	EventKeyUp
)

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}

type EventHandler func(param ...int)
