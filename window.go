package oswnd

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}

type Size struct{
	Width int
	Height int
}

type Padding struct{
	Left int
	Top int
	Right int
	Bottom int
}

type EventListeners struct{
	OnKeyDown func(keyCode, count int)
	OnKeyUp func(keyCode int)
}

var wndMap = map[uintptr]*Window{}

func (w *Window) SetClientSzie(s Size) {
	r := w.GetRect()
	p := w.GetPadding()
	r.Width = p.Right + s.Width + p.Right
	r.Height = p.Top + s.Height + p.Bottom
	w.SetRect(r)
}
