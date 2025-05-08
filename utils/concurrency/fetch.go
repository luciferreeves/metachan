package concurrency

import (
	"sync"
)

// ParallelResult represents a result or error from a parallel computation
type ParallelResult[T any] struct {
	Value T
	Error error
}

// Parallel executes multiple functions concurrently and returns their results
// This is a powerful utility that allows us to fetch data from multiple APIs in parallel
func Parallel[T any](funcs ...func() (T, error)) []ParallelResult[T] {
	results := make([]ParallelResult[T], len(funcs))
	var wg sync.WaitGroup

	for i, f := range funcs {
		wg.Add(1)
		go func(index int, function func() (T, error)) {
			defer wg.Done()
			value, err := function()
			results[index] = ParallelResult[T]{
				Value: value,
				Error: err,
			}
		}(i, f)
	}

	wg.Wait()
	return results
}

// ParallelMapResult represents a result from a map operation that may contain an error
type ParallelMapResult[T any] struct {
	Value T
	Error error
}

// ParallelMap applies a function to each item in a slice concurrently
// This is useful for operations like fetching skip times for multiple episodes at once
func ParallelMap[T any, R any](items []T, f func(T) (R, error)) []ParallelMapResult[R] {
	results := make([]ParallelMapResult[R], len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(index int, element T) {
			defer wg.Done()
			value, err := f(element)
			results[index] = ParallelMapResult[R]{
				Value: value,
				Error: err,
			}
		}(i, item)
	}

	wg.Wait()
	return results
}

// ParallelMapWithLimit applies a function to each item in a slice concurrently
// with a maximum number of concurrent operations
// This is crucial for rate-limited APIs like AniSkip
func ParallelMapWithLimit[T any, R any](items []T, limit int, f func(T) (R, error)) []ParallelMapResult[R] {
	results := make([]ParallelMapResult[R], len(items))
	var wg sync.WaitGroup

	// Create a semaphore channel with the specified limit
	semaphore := make(chan struct{}, limit)

	for i, item := range items {
		wg.Add(1)
		go func(index int, element T) {
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() {
				// Release semaphore when done
				<-semaphore
				wg.Done()
			}()

			value, err := f(element)
			results[index] = ParallelMapResult[R]{
				Value: value,
				Error: err,
			}
		}(i, item)
	}

	wg.Wait()
	return results
}
