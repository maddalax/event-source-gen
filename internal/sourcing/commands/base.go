package commands

type Handler[T any] struct {
	handler *T
}
