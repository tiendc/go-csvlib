package csvlib

type Int interface {
	int | int8 | int16 | int32 | int64
}

type IntEx interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type UInt interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type UIntEx interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	float32 | float64
}

type FloatEx interface {
	~float32 | ~float64
}

type Number interface {
	Int | UInt | Float
}

type NumberEx interface {
	IntEx | UIntEx | FloatEx
}

type String interface {
	string
}

type StringEx interface {
	~string
}

type LTComparable interface {
	Int | IntEx | UInt | UIntEx | Float | FloatEx | String | StringEx
}
