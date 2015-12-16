package oswnd

/*
#cgo CFLAGS: -O2
#cgo LDFLAGS: -lX11
#include <stdlib.h>
#include <X11/Xlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
	"fmt"
)

var (
	dpy *C.Display
	attributes C.XSetWindowAttributes
	root C.Window
)

func Factory(f func()) {
	if working {
		return
	}
	working = true

	runtime.LockOSThread()

	dpy = C.XOpenDisplay(nil)
	attributes.background_pixel = C.XWhitePixel(dpy, 0)
	root = C.XRootWindow(dpy, 0)

	f()
	if len(wndMap) == 0 {
		return
	}

	var event C.XClientMessageEvent
	for {
		C.XNextEvent(dpy, (*C.XEvent)(unsafe.Pointer(&event)))
		fmt.Println(event.message_type)
		if event._type == C.ClientMessage {
			C.XCloseDisplay(dpy)
			C.exit(0)
		}
	}
}

type Window struct{
	id uintptr
	did uintptr
	EventListeners
	padding Bounds
	border Bounds
	sizeDiff Size
}

func New() *Window {
	xWnd := C.XCreateWindow(
		dpy,
		root,
		0,
		0,
		500,
		500,
		0,
		C.XDefaultDepth(dpy, 0),
		C.InputOutput,
		C.XDefaultVisual(dpy, 0),
		C.CWBackPixel,
		&attributes,
	)
	wdw := C.CString("WM_DELETE_WINDOW")
	wmDelete := C.XInternAtom(dpy, wdw, 1)
	C.free(unsafe.Pointer(wdw))
	fmt.Println(wmDelete)
	C.XSetWMProtocols(dpy, xWnd, &wmDelete, 1)
	C.XSelectInput(dpy, xWnd, C.ExposureMask | C.KeyPressMask)
	C.XMapWindow(dpy, xWnd)

	wnd := &Window{
		id: uintptr(xWnd),
	}

	wndMap[wnd.id] = wnd
	return wnd
}

func (w *Window) SetTitle(title string) {
	
}

func (w *Window) GetRect() Rect {
	return Rect{}
}

func (w *Window) SetRect(r Rect) {

}

func (w *Window) MoveToScreenCenter() {
	
}

func (w *Window) SetLayout(flag int) {
	
}
