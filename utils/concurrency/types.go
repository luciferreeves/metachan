package concurrency

type ParallelResult[T any] struct {
	Value T
	Error error
}

type ParallelMapResult[T any] struct {
	Value T
	Error error
}
