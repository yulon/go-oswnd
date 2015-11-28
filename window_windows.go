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

	defWindowProc = user32.NewProc("DefWindowProcW").Call
	postQuitMessage = user32.NewProc("PostQuitMessage").Call

	loadIcon = user32.NewProc("LoadIconW").Call
	hcDefault, _, _ = loadIcon(hProcess, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("DEFAULT_ICON"))))

	loadCursor = user32.NewProc("LoadCursorW").Call
	hcArrow, _, _ = loadCursor(0, idc_arrow)

	wc *wndclassex

	registerClassEx = user32.NewProc("RegisterClassExW").Call
	createWindowEx = user32.NewProc("CreateWindowExW").Call
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
			wnd, ok := wndMap[hWnd]
			if ok {
				mh, ok := wnd.msgHandlers[uMsg]
				if ok {
					ret := mh(wParam, lParam)
					if !ret {
						return 0
					}
				}
			}
			ret, _, _ := defWindowProc(hWnd, uMsg, wParam, lParam)
			return ret
		}),
	}
	wc.cbSize = uint32(unsafe.Sizeof(*wc))
	registerClassEx(uintptr(unsafe.Pointer(wc)))

	f()
	if len(wndMap) == 0 {
		return
	}

	GetMessage := user32.NewProc("GetMessageW").Call
	DispatchMessage := user32.NewProc("DispatchMessageW").Call
	TranslateMessage := user32.NewProc("TranslateMessage").Call
	msg := &msg{}

	for {
		ret, _, _ := GetMessage(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
		if ret == 0 {
			return
		}

		TranslateMessage(uintptr(unsafe.Pointer(msg)))
		DispatchMessage(uintptr(unsafe.Pointer(msg)))
	}
}

type Window struct{
	hWnd uintptr
	msgHandlers map[uintptr]msgHandler
	keyCounts map[uintptr]int
	*EventHandlers
}

type msgHandler func(wParam, lParam uintptr) bool

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
	hWnd, _, err := createWindowEx(
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
	wnd := &Window{
		hWnd: hWnd,
		keyCounts: map[uintptr]int{},
		EventHandlers: &EventHandlers{},
	}
	wnd.msgHandlers = map[uintptr]msgHandler{
		wm_keydown: func(wParam, lParam uintptr) bool {
			wnd.keyCounts[wParam]++
			if wnd.OnKeyDown != nil {
				wnd.OnKeyDown(int(wParam), wnd.keyCounts[wParam])
			}
			return true
		},
		wm_keyup: func(wParam, lParam uintptr) bool {
			wnd.keyCounts[wParam] = 0
			if wnd.OnKeyUp != nil {
				wnd.OnKeyUp(int(wParam))
			}
			return true
		},
		wm_destroy: func(wParam, lParam uintptr) bool {
			delete(wndMap, wnd.hWnd)
			if len(wndMap) == 0 {
				postQuitMessage(0)
				return false
			}
			return true
		},
	}
	wndMap[hWnd] = wnd
	return wnd
}

func (w *Window) GetUnderlyingObject() uintptr {
	return w.hWnd
}

var (
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW").Call
	getWindowText = user32.NewProc("GetWindowTextW").Call
)

func (w *Window) GetTitle() string {
	leng, _, _ := getWindowTextLength(w.hWnd)
	str := syscall.StringToUTF16(strings.Repeat(" ", int(leng)))
	getWindowText(w.hWnd, uintptr(unsafe.Pointer(&str[0])), leng)
	return syscall.UTF16ToString(str)
}

var setWindowText = user32.NewProc("SetWindowTextW").Call

func (w *Window) SetTitle(title string) {
	setWindowText(w.hWnd, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
}

var getWindowRect = user32.NewProc("GetWindowRect").Call

func (w *Window) GetRect() *Rect {
	r := &rect{}
	getWindowRect(w.hWnd, uintptr(unsafe.Pointer(r)))
	return &Rect{
		int(r.Left),
		int(r.Top),
		int(r.Right - r.Left),
		int(r.Bottom - r.Top),
	}
}

var getClientRect = user32.NewProc("GetClientRect").Call
var clientToScreen = user32.NewProc("ClientToScreen").Call

func (w *Window) GetClientRect() *Rect {
	r := &rect{}
	getClientRect(w.hWnd, uintptr(unsafe.Pointer(r)))
	clientToScreen(w.hWnd, uintptr(unsafe.Pointer(r)))
	return &Rect{
		int(r.Left),
		int(r.Top),
		int(r.Right),
		int(r.Bottom),
	}
}

var moveWindow = user32.NewProc("MoveWindow").Call

func (w *Window) SetRect(r *Rect) {
	moveWindow(w.hWnd, uintptr(r.Left), uintptr(r.Top), uintptr(r.Width), uintptr(r.Height), 1)
}
