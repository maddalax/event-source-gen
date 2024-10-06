package util

type RecursiveStruct[T any] interface {
	Value() T
	GetChildren() []T
}

func RecursiveEach[T RecursiveStruct[T]](structure RecursiveStruct[T], callback func(T)) {
	callback(structure.Value())
	for _, child := range structure.GetChildren() {
		callback(child)
		RecursiveEach(child, callback)
	}
}
