package shpanstream

import (
	"context"
	"fmt"
	"io"
)

type clusterSortedStream[T any, O any, C comparable] struct {
	nextItem              *T
	currClassifier        C
	src                   *stream[T]
	clusterClassifierFunc func(a *T) C
	merger                func(ctx context.Context, clusterClassifier C, clusterStream Stream[T]) (*O, error)
}

func ClusterSortedStream[T any, O any, C comparable](
	clusterFactory func(ctx context.Context, clusterClassifier C, clusterStream Stream[T]) (*O, error),
	clusterClassifierFunc func(a *T) C,
	src Stream[T]) Stream[O] {

	return NewStream[O](&clusterSortedStream[T, O, C]{
		src:                   src.(*stream[T]),
		clusterClassifierFunc: clusterClassifierFunc,
		merger:                clusterFactory,
	})

}

func (fs *clusterSortedStream[T, O, C]) Open(ctx context.Context) error {
	openErr := openSubStream(ctx, fs.src)
	if openErr != nil {
		return openErr
	}
	nextItem, firstErr := fs.src.provider(ctx)
	if firstErr != nil {
		if firstErr == io.EOF {
			fs.nextItem = nil
			return nil
		}
		return firstErr
	}

	fs.nextItem = nextItem
	fs.currClassifier = fs.clusterClassifierFunc(nextItem)
	return nil
}

func (fs *clusterSortedStream[T, O, C]) Close() {
	closeSubStream(fs.src)
}

func (fs *clusterSortedStream[T, O, C]) Emit(ctx context.Context) (*O, error) {
	if fs.nextItem == nil {
		return nil, io.EOF
	}

	currClusterClassifier := fs.currClassifier

	// Create a cluster stream that yields items belonging to the current cluster
	clusterStream := newStream(

		func(ctx context.Context) (*T, error) {
			if fs.nextItem == nil {
				return nil, io.EOF
			}

			nextClassifier := fs.clusterClassifierFunc(fs.nextItem)
			if nextClassifier != currClusterClassifier {
				// Next item belongs to a new cluster
				return nil, io.EOF
			}

			// Yield fs.nextItem
			item := fs.nextItem

			// Advance fs.nextItem
			next, err := fs.src.provider(ctx)
			if err != nil {
				if err == io.EOF {
					// No more items
					fs.nextItem = nil
				} else {
					// Error occurred
					return nil, err
				}
			} else {
				fs.nextItem = next
			}

			return item, nil
		},

		// Avoid closing the underlying stream!
		nil,
	)

	// Call the merger function with the current cluster classifier and the cluster stream
	result, mergeErr := fs.merger(ctx, currClusterClassifier, clusterStream)
	if mergeErr != nil {
		// Make sure we wrap the error so e.g. even if it is io.EOF, it is not mistaken for end of stream (because go is stupid)
		return nil, fmt.Errorf("failed merging: %w", mergeErr)
	}

	// Update fs.currClassifier if fs.nextItem is not nil
	if fs.nextItem != nil {
		fs.currClassifier = fs.clusterClassifierFunc(fs.nextItem)
	}

	return result, nil
}
