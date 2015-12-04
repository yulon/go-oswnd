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
	pt point
}

type point struct{
	x int16
	y int16
}

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
	adjustWindowRectEx = user32.NewProc("AdjustWindowRectEx").Call
)

const (
	cs_hredraw = 0x0002
	cs_vredraw = 0x0001
	cs_dblclks = 0x0008

	idc_arrow = 32512
)

func Factory(f func()) {
	wc = &wndclassex{
		style: cs_hredraw | cs_vredraw | cs_dblclks,
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
	EventListeners
	padding Bounds
}

type msgHandler func(wParam, lParam uintptr) bool

const (
	ws_ex_dlgmodalframe = 0x00000001

	ws_caption = 0x00C00000
	ws_sysmenu = 0x00080000
	ws_overlapped = 0x00000000
	ws_thickframe = 0x00040000
	ws_maximizebox = 0x00010000
	ws_minimizebox = 0x00020000

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

func New() *Window {
	var dwExStyle uintptr = ws_ex_dlgmodalframe
	var dwStyle uintptr = ws_caption | ws_sysmenu | ws_overlapped | ws_thickframe | ws_maximizebox | ws_minimizebox

	hWnd, _, err := createWindowEx(
		dwExStyle,
		wc.lpszClassName,
		0,
		dwStyle,
		0,
		0,
		0,
		0,
		0,
		0,
		hProcess,
		0,
	)
	if hWnd == 0 {
		fmt.Println("oswnd:", err)
		return nil
	}

	b32 := bounds32{500, 500, 1000, 1000}
	adjustWindowRectEx(uintptr(unsafe.Pointer(&b32)), dwStyle, 0, dwExStyle)

	wnd := &Window{
		hWnd: hWnd,
		keyCounts: map[uintptr]int{},
		EventListeners: EventListeners{},
		padding: Bounds{
			500 - int(b32.Left),
			500 - int(b32.Top),
			int(b32.Right) - 1000,
			int(b32.Bottom) - 1000,
		},
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

func (w *Window) GetId() uintptr {
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

func (w *Window) GetRect() Rect {
	b32 := bounds32{}
	getWindowRect(w.hWnd, uintptr(unsafe.Pointer(&b32)))
	return Rect{
		int(b32.Left),
		int(b32.Top),
		int(b32.Right - b32.Left),
		int(b32.Bottom - b32.Top),
	}
}

var moveWindow = user32.NewProc("MoveWindow").Call

func (w *Window) SetRect(r Rect) {
	moveWindow(w.hWnd, uintptr(r.Left), uintptr(r.Top), uintptr(r.Width), uintptr(r.Height), 1)
}

const (
	sm_cxscreen = 0x000
	sm_cyscreen = 0x001
)

var getSystemMetrics = user32.NewProc("GetSystemMetrics").Call

func (w *Window) MoveToScreenCenter() {
	r := w.GetRect()
	sw, _, _ := getSystemMetrics(sm_cxscreen)
	sh, _, _ := getSystemMetrics(sm_cyscreen)
	r.Left = (int(sw) - r.Width) / 2
	r.Top = (int(sh) - r.Height) / 2
	w.SetRect(r)
}

var showWindow = user32.NewProc("ShowWindow").Call

const (
	sw_show = 0x005
	sw_hide = 0x000
	sw_maximize = 0x003
	sw_minimize = 0x006
	sw_restore = 0x009
	sw_showdefault = 0x00A
)

var lf2swf = map[int]uintptr{
	LayoutVisible: sw_show,
	LayoutHidden: sw_hide,
	LayoutMaximize: sw_maximize,
	LayoutMinimize: sw_minimize,
	LayoutRestore: sw_restore,
	LayoutDefault: sw_showdefault,
}

func (w *Window) SetLayout(flag int) {
	showWindow(w.hWnd, lf2swf[flag])
}

const (
	gwl_style = 0x00FFFFFFF0

	ws_visible = 0x10000000
	ws_maximize = 0x01000000
	ws_minimize = 0x20000000
)

var getWindowLong = user32.NewProc("GetWindowLongW").Call

func (w *Window) GetLayout() int {
	dwStyle, _, _ := getWindowLong(w.hWnd, gwl_style)
	if dwStyle & ws_visible == 0 {
		return LayoutHidden
	} else {
		switch {
			case dwStyle & ws_maximize > 0:
				return LayoutMaximize
			case dwStyle & ws_minimize > 0:
				return LayoutMinimize
			default:
				return LayoutDefault
		}
	}
}
