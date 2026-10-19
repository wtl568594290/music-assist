package services

import (
	"context"
	"sync"
)

type Result[T any] struct {
	Item T
	Err  error
}

type Options struct {
	MaxConcurrent int

	// 每个任务完成后调用
	OnComplete func(index int, err error)
}

func RunConcurrent[T any](
	ctx context.Context,
	items []T,
	options Options,
	fn func(T) error,
) []Result[T] {

	if len(items) == 0 {
		return nil
	}

	maxConcurrent := options.MaxConcurrent

	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	if maxConcurrent > len(items) {
		maxConcurrent = len(items)
	}

	results := make([]Result[T], len(items))
	tasks := make(chan int)

	var wg sync.WaitGroup

	// Worker
	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return

				case index, ok := <-tasks:
					if !ok {
						return
					}

					err := fn(items[index])

					results[index] = Result[T]{
						Item: items[index],
						Err:  err,
					}

					if options.OnComplete != nil {
						options.OnComplete(index, err)
					}
				}
			}
		}()
	}

	// Producer
	for i := range items {
		select {
		case <-ctx.Done():
			close(tasks)
			wg.Wait()
			return results

		case tasks <- i:
		}
	}

	close(tasks)
	wg.Wait()

	return results
}
