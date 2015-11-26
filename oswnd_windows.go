package oswnd

import (
	"syscall"
	"unsafe"
	"strings"
)

type wndclassex struct{
	cbSize uintptr
	style uintptr
	lpfnWndProc uintptr
	cbClsExtra uintptr
	cbWndExtra uintptr
	hInstance uintptr
	hIcon uintptr
	hCursor uintptr
	hbrBackground uintptr
	lpszMenuName uintptr
	lpszClassName uintptr
	hIconSm uintptr
}

type msg struct{
	hwnd uintptr
	message uintptr
	wParam uintptr
	lParam uintptr
	time uintptr
	x uintptr
	y uintptr
}

const (
	wm_paint = 0x000F
	wm_keydown = 0x0100
	wm_keyup = 0x0101
	wm_destroy = 0x0002
	wm_size = 0x0005
)

var (
	user32, _ = syscall.LoadDLL("user32.dll")
	kernel32, _ = syscall.LoadDLL("kernel32.dll")

	hProcess, _, _ = kernel32.MustFindProc("GetModuleHandleW").Call(0)

	defWindowProc = user32.MustFindProc("DefWindowProcW")
	postQuitMessage = user32.MustFindProc("PostQuitMessage")

	loadCursor = user32.MustFindProc("LoadCursorW")
	cArrow, _, _ = loadCursor.Call(0, 32512)

	winMap = map[uintptr]*Window{}

	wClass = &wndclassex{
		cbSize: 48,
		style: 2 | 1 | 8,
		hInstance: hProcess,
		hCursor: cArrow,
		hbrBackground: 15 + 1,
		lpszClassName: uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("pxwin"))),
		lpfnWndProc: syscall.NewCallback(func(hWnd, uMsg, wParam, lParam uintptr) uintptr {
			win, ok := winMap[hWnd]
			if ok {
				h, ok := win.msgHandlers[uMsg]
				if ok {
					ret := h(hWnd, wParam, lParam)
					if ret {
						return 0
					}
				}
			}
			if uMsg == wm_destroy {
				delete(winMap, hWnd)
				if len(winMap) == 0 {
					postQuitMessage.Call(0)
					return 0
				}
			}
			ret, _, _ := defWindowProc.Call(hWnd, uMsg, wParam, lParam)
			return ret
		}),
	}

	registerClassEx = user32.MustFindProc("RegisterClassExW")
	createWindowEx = user32.MustFindProc("CreateWindowExW")
)

func Main(f func()) {
	registerClassEx.Call(uintptr(unsafe.Pointer(wClass)))

	f()

	GetMessage := user32.MustFindProc("GetMessageW")
	DispatchMessage := user32.MustFindProc("DispatchMessageW")
	TranslateMessage := user32.MustFindProc("TranslateMessage")
	msg := &msg{}

	for {
		ret, _, _ := GetMessage.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
		if ret == 0 {
			return
		}

		TranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
		DispatchMessage.Call(uintptr(unsafe.Pointer(msg)))
	}
}

type Window struct{
	hWnd uintptr
	msgHandlers map[uintptr]msgHandler
}

const (
	ws_visible = 0x10000000
	ws_caption = 0x00C00000
	ws_sysmenu = 0x00080000
	ws_overlapped = 0x00000000
	ws_thickframe = 0x00040000
	ws_maximizebox = 0x00010000
	ws_minimizebox = 0x00020000
)

func New() *Window {
	hWnd, _, _ := createWindowEx.Call(
		1,
		wClass.lpszClassName,
		0,
		ws_visible | ws_caption | ws_sysmenu | ws_overlapped | ws_thickframe | ws_maximizebox | ws_minimizebox,
		0,
		0,
		500,
		500,
		0,
		0,
		hProcess,
		0,
	)
	win := &Window{
		hWnd,
		map[uintptr]msgHandler{},
	}
	winMap[hWnd] = win
	return win
}

func (w *Window) Pointer() uintptr {
	return w.hWnd
}

type msgHandler func(hWnd, wParam, lParam uintptr) bool

var (
	beginPaint = user32.MustFindProc("BeginPaint")
	endPaint = user32.MustFindProc("EndPaint")
)

var ehConv = map[int]func(eh EventHandler) (msg uintptr, mh msgHandler) {
	EventPaint: func(eh EventHandler) (msg uintptr, mh msgHandler) {
		return wm_paint, func(hWnd, wParam, lParam uintptr) bool {
			p := make([]byte, 64)
			beginPaint.Call(hWnd, uintptr(unsafe.Pointer(&p[0])))
			eh()
			endPaint.Call(hWnd, uintptr(unsafe.Pointer(&p[0])))
			return true
		}
	},
	EventKeyDown: func(eh EventHandler) (msg uintptr, mh msgHandler) {
		return wm_keydown, func(hWnd, wParam, lParam uintptr) bool {
			eh(int(wParam))
			return false
		}
	},
	EventKeyUp: func(eh EventHandler) (msg uintptr, mh msgHandler) {
		return wm_keyup, func(hWnd, wParam, lParam uintptr) bool {
			eh(int(wParam))
			return false
		}
	},
}

func (w *Window) On(eventType int, eh EventHandler) {
	msg, mh := ehConv[eventType](eh)
	w.msgHandlers[msg] = mh
}

var (
	getWindowTextLength = user32.MustFindProc("GetWindowTextLengthW")
	getWindowText = user32.MustFindProc("GetWindowTextW")
)

func (w *Window) GetTitle() string {
	leng, _, _ := getWindowTextLength.Call(w.hWnd)
	str := syscall.StringToUTF16(strings.Repeat(" ", int(leng)))
	getWindowText.Call(w.hWnd, uintptr(unsafe.Pointer(&str[0])), leng)
	return syscall.UTF16ToString(str)
}

var setWindowText = user32.MustFindProc("SetWindowTextW")

func (w *Window) SetTitle(title string) {
	setWindowText.Call(w.hWnd, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
}

var getWindowRect = user32.MustFindProc("GetWindowRect")

type ltrb struct{
	Left int
	Top int
	Right int
	Bottom int
}

func (w *Window) Rect() *Rect {
	wr := &ltrb{}
	getWindowRect.Call(w.hWnd, uintptr(unsafe.Pointer(wr)))
	return &Rect{
		wr.Left,
		wr.Top,
		wr.Right - wr.Left,
		wr.Bottom - wr.Top,
	}
}

var getClientRect = user32.MustFindProc("GetClientRect")
var clientToScreen = user32.MustFindProc("ClientToScreen")

func (w *Window) ClientRect() *Rect {
	wr := &Rect{}
	getClientRect.Call(w.hWnd, uintptr(unsafe.Pointer(wr)))
	clientToScreen.Call(w.hWnd, uintptr(unsafe.Pointer(wr)))
	return wr
}

var moveWindow = user32.MustFindProc("MoveWindow")

func (w *Window) Move(r *Rect) {
	moveWindow.Call(w.hWnd, uintptr(r.Left), uintptr(r.Top), uintptr(r.Width), uintptr(r.Height), 1)
}
