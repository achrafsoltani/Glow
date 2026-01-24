package x11

// X11 Request Opcodes
const (
	OpCreateWindow    = 1
	OpChangeWindowAttributes = 2
	OpGetWindowAttributes = 3
	OpDestroyWindow   = 4
	OpMapWindow       = 8
	OpUnmapWindow     = 10
	OpConfigureWindow = 12
	OpInternAtom      = 16
	OpChangeProperty  = 18
	OpDeleteProperty  = 19
	OpGetProperty     = 20
	OpCreateGC        = 55
	OpFreeGC          = 60
	OpPolyFillRect    = 70
	OpPutImage        = 72
)

// Window classes
const (
	WindowClassCopyFromParent = 0
	WindowClassInputOutput    = 1
	WindowClassInputOnly      = 2
)

// Window attributes mask
const (
	CWBackPixmap       = 1 << 0
	CWBackPixel        = 1 << 1
	CWBorderPixmap     = 1 << 2
	CWBorderPixel      = 1 << 3
	CWBitGravity       = 1 << 4
	CWWinGravity       = 1 << 5
	CWBackingStore     = 1 << 6
	CWBackingPlanes    = 1 << 7
	CWBackingPixel     = 1 << 8
	CWOverrideRedirect = 1 << 9
	CWSaveUnder        = 1 << 10
	CWEventMask        = 1 << 11
	CWDontPropagate    = 1 << 12
	CWColormap         = 1 << 13
	CWCursor           = 1 << 14
)

// Event masks - these determine which events we receive
const (
	KeyPressMask        = 1 << 0
	KeyReleaseMask      = 1 << 1
	ButtonPressMask     = 1 << 2
	ButtonReleaseMask   = 1 << 3
	EnterWindowMask     = 1 << 4
	LeaveWindowMask     = 1 << 5
	PointerMotionMask   = 1 << 6
	ExposureMask        = 1 << 15
	StructureNotifyMask = 1 << 17
	FocusChangeMask     = 1 << 21
)

// Event types - the type field in event packets
const (
	EventKeyPress        = 2
	EventKeyRelease      = 3
	EventButtonPress     = 4
	EventButtonRelease   = 5
	EventMotionNotify    = 6
	EventEnterNotify     = 7
	EventLeaveNotify     = 8
	EventFocusIn         = 9
	EventFocusOut        = 10
	EventExpose          = 12
	EventDestroyNotify   = 17
	EventUnmapNotify     = 18
	EventMapNotify       = 19
	EventConfigureNotify = 22
	EventClientMessage   = 33
)

// Image formats for PutImage
const (
	ImageFormatBitmap  = 0
	ImageFormatXYPixmap = 1
	ImageFormatZPixmap = 2
)
