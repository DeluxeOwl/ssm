package ssm

var End Fn = nil

var _endPtr = ptrOf(End)

func IsEnd(f Fn) bool {
	return ptrOf(f) == _endPtr
}
