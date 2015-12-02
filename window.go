package oswnd

type EventListeners struct{
	OnKeyDown func(keyCode, count int)
	OnKeyUp func(keyCode int)
}

var wndMap = map[uintptr]*Window{}

const (
	DisplayVisible = iota
	DisplayHidden
	DisplayMaximize
	DisplayMinimize
)

func (w *Window) SetClientSzie(s Size) {
	r := w.GetRect()
	p := w.GetPadding()
	r.Width = p.Right + s.Width + p.Right
	r.Height = p.Top + s.Height + p.Bottom
	w.SetRect(r)
}
