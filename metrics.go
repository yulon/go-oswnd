package oswnd

type Rect struct{
	Left int
	Top int
	Width int
	Height int
}

type rect32 struct{
	Left int32
	Top int32
	Width int32
	Height int32
}

type Size struct{
	Width int
	Height int
}

type size16 struct{
	Width uint16
	Height uint16
}

type Bounds struct{
	Left int
	Top int
	Right int
	Bottom int
}

type bounds32 struct{
	Left int32
	Top int32
	Right int32
	Bottom int32
}
