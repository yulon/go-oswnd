package oswnd

import (
	"syscall"
	"unsafe"
	"strings"
	"fmt"
)

type wndclassex struct{
	cbSize uint32
	style uint32
	lpfnWndProc uintptr
	cbClsExtra int32
	cbWndExtra int32
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
	message uint32
	wParam uint32
	lParam uint32
	time uint32
	x uint32
	y uint32
}

const (
	idc_arrow = 32512

	wm_paint = 0x000F
	wm_keydown = 0x0100
	wm_keyup = 0x0101
	wm_destroy = 0x0002
	wm_size = 0x0005
	wm_ncpaint = 0x0085
	wm_nccalcsize = 0x0083
	wm_move = 0x0003
	wm_moving = 0x0216
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	hProcess, _, _ = kernel32.NewProc("GetModuleHandleW").Call(0)

	defWindowProc = user32.NewProc("DefWindowProcW")
	postQuitMessage = user32.NewProc("PostQuitMessage")

	loadIcon = user32.NewProc("LoadIconW")
	hcDefault, _, _ = loadIcon.Call(hProcess, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("DEFAULT_ICON"))))

	loadCursor = user32.NewProc("LoadCursorW")
	hcArrow, _, _ = loadCursor.Call(0, idc_arrow)

	winMap = map[uintptr]*Window{}

	wc *wndclassex

	registerClassEx = user32.NewProc("RegisterClassExW")
	createWindowEx = user32.NewProc("CreateWindowExW")
)

type rect struct{
	Left int32
	Top int32
	Right int32
	Bottom int32
}

type size struct{
	Width uint16
	Height uint16
}

type nccalcsize_params struct{
	rgrc *[3]rect
	lppos uintptr
}

func Main(f func()) {
	wc = &wndclassex{
		style: 2 | 1 | 8,
		hInstance: hProcess,
		hIcon: hcDefault,
		hCursor: hcArrow,
		hbrBackground: 15 + 1,
		lpszClassName: uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("oswnd"))),
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
	wc.cbSize = uint32(unsafe.Sizeof(*wc))
	registerClassEx.Call(uintptr(unsafe.Pointer(wc)))

	f()
	if len(winMap) == 0 {
		return
	}

	GetMessage := user32.NewProc("GetMessageW")
	DispatchMessage := user32.NewProc("DispatchMessageW")
	TranslateMessage := user32.NewProc("TranslateMessage")
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
	rect *Rect
	cRect *Rect
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
	hWnd, _, err := createWindowEx.Call(
		1,
		wc.lpszClassName,
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
	if hWnd == 0 {
		fmt.Println("oswnd:", err)
		return nil
	}
	win := &Window{
		hWnd: hWnd,
		msgHandlers: map[uintptr]msgHandler{},
		rect: &Rect{},
		cRect: &Rect{},
	}
	winMap[hWnd] = win
	return win
}

func (w *Window) GetUnderlyingObject() uintptr {
	return w.hWnd
}

type msgHandler func(hWnd, wParam, lParam uintptr) bool

var (
	beginPaint = user32.NewProc("BeginPaint")
	endPaint = user32.NewProc("EndPaint")
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
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	getWindowText = user32.NewProc("GetWindowTextW")
)

func (w *Window) GetTitle() string {
	leng, _, _ := getWindowTextLength.Call(w.hWnd)
	str := syscall.StringToUTF16(strings.Repeat(" ", int(leng)))
	getWindowText.Call(w.hWnd, uintptr(unsafe.Pointer(&str[0])), leng)
	return syscall.UTF16ToString(str)
}

var setWindowText = user32.NewProc("SetWindowTextW")

func (w *Window) SetTitle(title string) {
	setWindowText.Call(w.hWnd, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
}

var getWindowRect = user32.NewProc("GetWindowRect")

func (w *Window) GetRect() *Rect {
	r := &rect{}
	getWindowRect.Call(w.hWnd, uintptr(unsafe.Pointer(r)))
	return &Rect{
		int(r.Left),
		int(r.Top),
		int(r.Right - r.Left),
		int(r.Bottom - r.Top),
	}
}

var getClientRect = user32.NewProc("GetClientRect")
var clientToScreen = user32.NewProc("ClientToScreen")

func (w *Window) GetClientRect() *Rect {
	r := &rect{}
	getClientRect.Call(w.hWnd, uintptr(unsafe.Pointer(r)))
	clientToScreen.Call(w.hWnd, uintptr(unsafe.Pointer(r)))
	return &Rect{
		int(r.Left),
		int(r.Top),
		int(r.Right),
		int(r.Bottom),
	}
}

var moveWindow = user32.NewProc("MoveWindow")

func (w *Window) Move(left int, top int) {
	moveWindow.Call(w.hWnd, uintptr(left), uintptr(top), uintptr(w.rect.Width), uintptr(w.rect.Height), 1)
}
