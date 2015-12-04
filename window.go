package oswnd

type EventListeners struct{
	OnKeyDown func(keyCode, count int)
	OnKeyUp func(keyCode int)
}

var wndMap = map[uintptr]*Window{}

const (
	VisibilityVisible = iota
	VisibilityHidden
	VisibilityMaximize
	VisibilityMinimize
	VisibilityRestore
)

func (w *Window) GetPadding() Bounds {
	return w.padding
}

func (w *Window) GetClientSzie() Size {
	r := w.GetRect()
	p := w.GetPadding()
	return Size{
		r.Width - p.Left - p.Right,
		r.Height - p.Top - p.Bottom,
	}
}

func (w *Window) SetClientSzie(set Size) {
	r := w.GetRect()
	p := w.GetPadding()
	r.Width = p.Right + set.Width + p.Right
	r.Height = p.Top + set.Height + p.Bottom
	w.SetRect(r)
}
