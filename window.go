package oswnd

type EventListeners struct{
	OnKeyDown func(keyCode, count int)
	OnKeyUp func(keyCode int)
	OnPaint func()
	OnSize func()
}

const (
	LayoutHidden = iota
	LayoutVisible
	LayoutDefault
	LayoutMaximize
	LayoutMinimize
	LayoutRestore
)

var working bool
var wndMap = map[uintptr]*Window{}

func (w *Window) GetId() uintptr {
	return w.id
}

func (w *Window) GetDisplayId() uintptr {
	return w.did
}

func (w *Window) GetPadding() Bounds {
	return w.padding
}

func (w *Window) GetBorder() Bounds {
	return w.border
}

func (w *Window) GetClientSzie() Size {
	r := w.GetRect()
	return Size{
		r.Width - w.sizeDiff.Width,
		r.Height - w.sizeDiff.Height,
	}
}

func (w *Window) SetClientSzie(set Size) {
	r := w.GetRect()
	r.Width = set.Width + w.sizeDiff.Width
	r.Height = set.Height + w.sizeDiff.Height
	w.SetRect(r)
}
