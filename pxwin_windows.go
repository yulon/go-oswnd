package pxwin

import (
	"syscall"
	"unsafe"
)

var (
	user32, _ = syscall.LoadLibrary("user32.dll")
	kernel32, _ = syscall.LoadLibrary("kernel32.dll")

	getModuleHandle, _ = syscall.GetProcAddress(kernel32, "GetModuleHandleW")
	hProcess, _, _ = syscall.Syscall(getModuleHandle, 1, 0, 0, 0)

	defWindowProc, _ = syscall.GetProcAddress(user32, "DefWindowProcW")
	postQuitMessage, _ = syscall.GetProcAddress(user32, "PostQuitMessage")

	wClass = &wndclassex{
		cbSize: 48,
		style: 2 | 1 | 8,
		hInstance: hProcess,
		hbrBackground: 15 + 1,
		lpszClassName: uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("pxwin"))),
		lpfnWndProc: syscall.NewCallback(func(hWnd, uMsg, wParam, lParam uintptr) uintptr {
			println(uMsg)
			if uMsg == 2 {
				syscall.Syscall(postQuitMessage, 1, 0, 0, 0)
				return 0
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

func MessageLoop() {
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

const (
	ws_visible = 0x10000000
	ws_caption = 0x00C00000
	ws_sysmenu = 0x00080000
	ws_overlapped = 0x00000000
	ws_thickframe = 0x00040000
	ws_maximizebox = 0x00010000
	ws_minimizebox = 0x00020000
)

func New(title string) {
	syscall.Syscall12(createWindowEx, 12,
		1,
		wClass.lpszClassName,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
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
}
