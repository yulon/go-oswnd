package oswnd

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}

type EventListeners struct{
	OnKeyDown func(keyCode, count int)
	OnKeyUp func(keyCode int)
}

var wndMap = map[uintptr]*Window{}
