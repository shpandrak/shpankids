package shpanstream

import (
	"context"
	"errors"
	"io"
	"shpankids/infra/util/functional"
)

type mergeSortingStream[T any] struct {
	streams    []*stream[T]
	comparator func(a, b *T) int
	nextBuffer []*T
}

func MergedSortedStream[T any](comparator func(a, b *T) int, streams ...Stream[T]) Stream[T] {
	if len(streams) == 0 {
		return EmptyStream[T]()
	}

	return NewStream(&mergeSortingStream[T]{
		streams: functional.MapSliceWhileFilteringNoErr(streams, func(s Stream[T]) **stream[T] {
			if internal, ok := s.(*stream[T]); ok {
				return &internal
			}
			return nil
		}),
		comparator: comparator,
		nextBuffer: make([]*T, len(streams)),
	})
}

func (ms *mergeSortingStream[T]) Open(ctx context.Context) error {
	var allErrors []error

	for _, s := range ms.streams {
		for _, l := range s.allLifecycleElement {
			if err := l.Open(ctx); err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}
	return errors.Join(allErrors...)
}

func (ms *mergeSortingStream[T]) Close() {
	for _, s := range ms.streams {
		for _, l := range s.allLifecycleElement {
			l.Close()
		}
	}
}

func (ms *mergeSortingStream[T]) Emit(ctx context.Context) (*T, error) {
	if ms.nextBuffer == nil {
		ms.nextBuffer = make([]*T, len(ms.streams))
		for i, s := range ms.streams {
			v, err := s.provider(ctx)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if err == nil {
				ms.nextBuffer[i] = v
			}
		}
	} else {
		for i, s := range ms.streams {
			if ms.nextBuffer[i] == nil {
				v, err := s.provider(ctx)
				if err != nil && err != io.EOF {
					return nil, err
				}
				if err == nil {
					ms.nextBuffer[i] = v
				}
			}
		}
	}

	minIndex := -1
	var min *T

	for i, v := range ms.nextBuffer {
		if v != nil {
			if min == nil || ms.comparator(v, min) < 0 {
				min = v
				minIndex = i
			}
		}
	}

	if min == nil {
		return nil, io.EOF
	}

	ms.nextBuffer[minIndex] = nil
	return min, nil
}
