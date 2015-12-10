package oswnd

import (
	"runtime"
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
	pt point16
}

var (
	working bool

	user32, _ = syscall.LoadLibrary("user32.dll")
	kernel32, _ = syscall.LoadLibrary("kernel32.dll")

	getModuleHandleW, _ = syscall.GetProcAddress(kernel32, "GetModuleHandleW")
	hProcess, _, _ = syscall.Syscall(getModuleHandleW, 1, 0, 0, 0)

	defWindowProc, _ = syscall.GetProcAddress(user32, "DefWindowProcW")
	postQuitMessage, _ = syscall.GetProcAddress(user32, "PostQuitMessage")

	loadIcon, _ = syscall.GetProcAddress(user32, "LoadIconW")
	hiApp, _, _ = syscall.Syscall(loadIcon, 2, 0, idi_application, 0)
	hiLogo, _, _ = syscall.Syscall(loadIcon, 2, 0, idi_winlogo, 0)

	loadCursor, _ = syscall.GetProcAddress(user32, "LoadCursorW")
	hcArrow, _, _ = syscall.Syscall(loadCursor, 2, 0, idc_arrow, 0)

	wc *wndclassex

	registerClassEx, _ = syscall.GetProcAddress(user32, "RegisterClassExW")
	createWindowEx, _ = syscall.GetProcAddress(user32, "CreateWindowExW")
	adjustWindowRectEx, _ = syscall.GetProcAddress(user32, "AdjustWindowRectEx")

	validateRect, _ = syscall.GetProcAddress(user32, "ValidateRect")
)

const (
	cs_dblclks = 0x0008

	idi_application = 0x007F00
	idi_winlogo = 0x007F05
	idc_arrow = 0x007F00
)

func Factory(f func()) {
	if working {
		return
	}
	working = true

	runtime.LockOSThread()

	wc = &wndclassex{
		style: cs_dblclks,
		hInstance: hProcess,
		hIcon: hiApp,
		hIconSm: hiLogo,
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
			ret, _, _ := syscall.Syscall6(defWindowProc, 4, hWnd, uMsg, wParam, lParam, 0, 0)
			return ret
		}),
	}
	wc.cbSize = uint32(unsafe.Sizeof(*wc))
	syscall.Syscall(registerClassEx, 1, uintptr(unsafe.Pointer(wc)), 0, 0)

	f()
	if len(wndMap) == 0 {
		return
	}

	GetMessage, _ := syscall.GetProcAddress(user32, "GetMessageW")
	TranslateMessage, _ := syscall.GetProcAddress(user32, "TranslateMessage")
	DispatchMessage, _ := syscall.GetProcAddress(user32, "DispatchMessageW")
	msg := uintptr(unsafe.Pointer(&msg{}))
	var ret uintptr
	for {
		ret, _, _ = syscall.Syscall6(GetMessage, 4, msg, 0, 0, 0, 0, 0)
		if ret == 0 {
			return
		}
		syscall.Syscall(TranslateMessage, 1, msg, 0, 0)
		syscall.Syscall(DispatchMessage, 1, msg, 0, 0)
	}
}

type Window struct{
	hWnd uintptr
	msgHandlers map[uintptr]msgHandler
	keyCounts map[uintptr]int
	EventListeners
	padding Bounds
	customPadding bool
	border Bounds
	customBorder bool
	sizeDiff Size
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

	hWnd, _, err := syscall.Syscall12(
		createWindowEx,
		12,
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

	boundDiffs := bounds32{500, 500, 1000, 1000}
	syscall.Syscall6(adjustWindowRectEx, 4, uintptr(unsafe.Pointer(&boundDiffs)), dwStyle, 0, dwExStyle, 0, 0)
	boundDiffs.Left = 500 - boundDiffs.Left
	boundDiffs.Top = 500 - boundDiffs.Top
	boundDiffs.Right -= 1000
	boundDiffs.Bottom -= 1000

	wnd := &Window{
		hWnd: hWnd,
		keyCounts: map[uintptr]int{},
		EventListeners: EventListeners{},
		padding: Bounds{
			int(boundDiffs.Left),
			int(boundDiffs.Bottom),
			int(boundDiffs.Right),
			int(boundDiffs.Bottom),
		},
		border: Bounds{
			0,
			int(boundDiffs.Top - boundDiffs.Bottom),
			0,
			0,
		},
		sizeDiff: Size{
			int(boundDiffs.Left + boundDiffs.Right),
			int(boundDiffs.Top + boundDiffs.Bottom),
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
				syscall.Syscall(postQuitMessage, 1, 0, 0, 0)
				return false
			}
			return true
		},
		wm_paint: func(wParam, lParam uintptr) bool {
			if wnd.OnPaint != nil {
				wnd.OnPaint()
				syscall.Syscall(validateRect, 2, wnd.hWnd, 0, 0)
				return false
			}
			return true
		},
		wm_size: func(wParam, lParam uintptr) bool {
			if wnd.OnSize != nil {
				wnd.OnSize()
			}
			if wnd.OnPaint != nil {
				wnd.OnPaint()
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
	getWindowTextLength, _ = syscall.GetProcAddress(user32, "GetWindowTextLengthW")
	getWindowText, _ = syscall.GetProcAddress(user32, "GetWindowTextW")
)

func (w *Window) GetTitle() string {
	leng, _, _ := syscall.Syscall(getWindowTextLength, 1, w.hWnd, 0, 0)
	str := syscall.StringToUTF16(strings.Repeat(" ", int(leng)))
	syscall.Syscall(getWindowText, 3, w.hWnd, uintptr(unsafe.Pointer(&str[0])), leng)
	return syscall.UTF16ToString(str)
}

var setWindowText, _ = syscall.GetProcAddress(user32, "SetWindowTextW")

func (w *Window) SetTitle(title string) {
	syscall.Syscall(setWindowText, 2, w.hWnd, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0)
}

var getWindowRect, _ = syscall.GetProcAddress(user32, "GetWindowRect")

func (w *Window) GetRect() Rect {
	b32 := bounds32{}
	syscall.Syscall(getWindowRect, 2, w.hWnd, uintptr(unsafe.Pointer(&b32)), 0)
	return Rect{
		int(b32.Left),
		int(b32.Top),
		int(b32.Right - b32.Left),
		int(b32.Bottom - b32.Top),
	}
}

var moveWindow, _ = syscall.GetProcAddress(user32, "MoveWindow")

func (w *Window) SetRect(r Rect) {
	syscall.Syscall6(moveWindow, 6, w.hWnd, uintptr(r.Left), uintptr(r.Top), uintptr(r.Width), uintptr(r.Height), 1)
}

const (
	sm_cxscreen = 0x000
	sm_cyscreen = 0x001
)

var getSystemMetrics, _ = syscall.GetProcAddress(user32, "GetSystemMetrics")

func (w *Window) MoveToScreenCenter() {
	r := w.GetRect()
	sw, _, _ := syscall.Syscall(getSystemMetrics, 1, sm_cxscreen, 0, 0)
	sh, _, _ := syscall.Syscall(getSystemMetrics, 1, sm_cyscreen, 0, 0)
	r.Left = (int(sw) - r.Width) / 2
	r.Top = (int(sh) - r.Height) / 2
	w.SetRect(r)
}

var showWindow, _ = syscall.GetProcAddress(user32, "ShowWindow")

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
	syscall.Syscall(showWindow, 2, w.hWnd, lf2swf[flag], 0)
}

const (
	gwl_style = 0x00FFFFFFF0

	ws_visible = 0x10000000
	ws_maximize = 0x01000000
	ws_minimize = 0x20000000
)

var getWindowLong, _ = syscall.GetProcAddress(user32, "GetWindowLongW")

func (w *Window) GetLayout() int {
	dwStyle, _, _ := syscall.Syscall(getWindowLong, 2, w.hWnd, gwl_style, 0)
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
