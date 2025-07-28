package utils

import (
	"context"
	"runtime"
	"sync"

	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"golang.org/x/sync/errgroup"
)

// WrapProcessor adapts a context-less processor to be used with ProcessSliceParallel
func WrapProcessor[T any, R any](processor func(T) R) func(context.Context, T) R {
	return func(_ context.Context, item T) R {
		return processor(item)
	}
}

// ProcessSliceAutoParallel automatically chooses number of workers (CPU count)
func ProcessSliceAutoParallel[T any, R any](ctx context.Context, items []T, processor func(context.Context, T) R) []R {
	return ProcessSliceParallel(ctx, items, processor, runtime.NumCPU())
}

// ProcessSliceParallel processes a slice in parallel using goroutines
// This is useful for CPU-intensive operations like data transformation
func ProcessSliceParallel[T any, R any](ctx context.Context, items []T, processor func(context.Context, T) R, maxWorkers int) []R {
	if len(items) == 0 {
		return make([]R, 0)
	}

	// For small datasets, use sequential processing
	if len(items) < constants.SmallDatasetThreshold {
		result := make([]R, len(items))
		for i, item := range items {
			result[i] = processor(ctx, item)
		}
		return result
	}

	// Limit number of workers to prevent excessive goroutine creation
	if maxWorkers <= 0 || maxWorkers > len(items) {
		maxWorkers = len(items)
	}

	result := make([]R, len(items))
	var wg sync.WaitGroup

	// Create a channel to distribute work
	workChan := make(chan int, len(items))

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for { // Loop indefinitely
				select {
				case <-ctx.Done(): // Check for cancellation
					return // Exit the worker
				case idx, ok := <-workChan:
					if !ok {
						return // Channel closed, exit the worker
					}
					// Pass the context down to the processor
					result[idx] = processor(ctx, items[idx])
				}
			}
		}()
	}

	// Send work to workers
	for i := 0; i < len(items); i++ {
		workChan <- i
	}
	close(workChan)

	// Wait for all workers to complete
	wg.Wait()

	return result
}

// ProcessSliceSequential processes a slice sequentially
// Use this for I/O bound operations or when order matters
func ProcessSliceSequential[T any, R any](items []T, processor func(T) R) []R {
	if len(items) == 0 {
		return make([]R, 0)
	}

	result := make([]R, len(items))
	for i, item := range items {
		result[i] = processor(item)
	}
	return result
}

// RunQueriesParallel runs multiple database queries in parallel
// This is useful when you have independent queries that can run simultaneously
func RunQueriesParallel[T any](ctx context.Context, queries []func(context.Context) (T, error)) ([]T, error) {
	if len(queries) == 0 {
		return make([]T, 0), nil
	}

	// For single query, run sequentially
	if len(queries) == 1 {
		result, err := queries[0](ctx)
		if err != nil {
			return nil, err
		}
		return []T{result}, nil
	}

	type queryResult struct {
		index  int
		result T
		err    error
	}

	resultChan := make(chan queryResult, len(queries))

	// Run queries in parallel
	for i, query := range queries {
		go func(idx int, q func(context.Context) (T, error)) {
			result, err := q(ctx)
			resultChan <- queryResult{
				index:  idx,
				result: result,
				err:    err,
			}
		}(i, query)
	}

	// Collect results in order
	results := make([]T, len(queries))
	for i := 0; i < len(queries); i++ {
		select {
		case <-ctx.Done(): // Check for cancellation on every iteration
			return nil, ctx.Err() // Exit immediately if context is cancelled
		case res := <-resultChan:
			if res.err != nil {
				return nil, res.err // Exit on the first error
			}
			results[res.index] = res.result
		}
	}

	return results, nil
}

// RunConcurrent runs a list of functions concurrently and returns the first error (if any).
func RunConcurrent(ctx context.Context, tasks ...func(context.Context) error) error {
	// g is an errgroup.Group, created with a context.
	// It automatically handles cancellation: if one task fails, it cancels the context for all others.
	g, gCtx := errgroup.WithContext(ctx)

	for _, task := range tasks {
		// Capture the task variable in the closure
		currentTask := task
		g.Go(func() error {
			// Run the task with the group's context
			return currentTask(gCtx)
		})
	}

	// Wait() returns the first non-nil error from any of the tasks.
	// If the context is cancelled, it will also return the context's error.
	return g.Wait()
}
