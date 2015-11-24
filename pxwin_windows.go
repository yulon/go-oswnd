package pxwin

import (
	"syscall"
	"unsafe"
)

const (
	ws_visible = 0x10000000
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
)

var (
	user32, _ = syscall.LoadLibrary("user32.dll")
	kernel32, _ = syscall.LoadLibrary("kernel32.dll")

	getModuleHandle, _ = syscall.GetProcAddress(kernel32, "GetModuleHandleW")
	hProcess, _, _ = syscall.Syscall(getModuleHandle, 1, 0, 0, 0)

	defWindowProc, _ = syscall.GetProcAddress(user32, "DefWindowProcW")
	postQuitMessage, _ = syscall.GetProcAddress(user32, "PostQuitMessage")

	loadCursor, _ = syscall.GetProcAddress(user32, "LoadCursorW")
	cArrow, _, _ = syscall.Syscall(loadCursor, 2, 0, 32512, 0)

	winMap = map[uintptr]*windowsWindow{}

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
				if win.eventListener != nil {
					switch uMsg {
						case wm_paint:
							win.eventListener(EventPaint, 0, 0)
							return 0
						case wm_keydown:
							win.eventListener(EventKeyDown, 0, 0)
						case wm_keyup:
							win.eventListener(EventKeyUp, 0, 0)
						case wm_destroy:
							delete(winMap, hWnd)
							if len(winMap) == 0 {
								syscall.Syscall(postQuitMessage, 1, 0, 0, 0)
								return 0
							}
					}
				}
				if uMsg == wm_destroy {
					delete(winMap, hWnd)
					if len(winMap) == 0 {
						syscall.Syscall(postQuitMessage, 1, 0, 0, 0)
						return 0
					}
				}
			}
			ret, _, _ := syscall.Syscall6(defWindowProc, 4, hWnd, uMsg, wParam, lParam, 0, 0)
			return ret
		}),
	}

	createWindowEx, _ = syscall.GetProcAddress(user32, "CreateWindowExW")
)

func Init() {
	registerClassEx, _ := syscall.GetProcAddress(user32, "RegisterClassExW")
	syscall.Syscall(registerClassEx, 1, uintptr(unsafe.Pointer(wClass)), 0, 0)
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

func EventDrive() {
	GetMessage, _ := syscall.GetProcAddress(user32, "GetMessageW")
	DispatchMessage, _ := syscall.GetProcAddress(user32, "DispatchMessageW")
	TranslateMessage, _ := syscall.GetProcAddress(user32, "TranslateMessage")
	msg := &msg{}

	for {
		ret, _, _ := syscall.Syscall6(GetMessage, 4, uintptr(unsafe.Pointer(msg)), 0, 0, 0, 0, 0)
		if ret == 0 {
			return
		}

		syscall.Syscall(TranslateMessage, 1, uintptr(unsafe.Pointer(msg)), 0, 0)
		syscall.Syscall(DispatchMessage, 1, uintptr(unsafe.Pointer(msg)), 0, 0)
	}
}

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

type windowsWindow struct{
	hWnd uintptr
	eventListener func(int, int, int)
}

func (w *windowsWindow) SetEventListener(h func(int, int, int)) {
	w.eventListener = h
}

func New() Window {
	hWnd, _, _ := syscall.Syscall12(createWindowEx, 12,
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
	win := &windowsWindow{}
	winMap[hWnd] = win
	return win
}
