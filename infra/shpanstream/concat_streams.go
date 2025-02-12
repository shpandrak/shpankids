package shpanstream

import (
	"context"
	"errors"
	"io"
	"shpankids/infra/util/functional"
)

type concatenatedStream[T any] struct {
	streams []*stream[T]
}

func ConcatenatedStream[T any](streams ...Stream[T]) Stream[T] {
	if len(streams) == 0 {
		return EmptyStream[T]()
	}
	return NewStream(&concatenatedStream[T]{
		streams: functional.MapSliceWhileFilteringNoErr(streams, func(s Stream[T]) **stream[T] {
			if internal, ok := s.(*stream[T]); ok {
				return &internal
			}
			return nil
		}),
	})

}

func (ms *concatenatedStream[T]) Open(ctx context.Context) error {
	// Only open the first stream
	return openSubStream(ctx, ms.streams[0])
}

func openSubStream[T any](ctx context.Context, s *stream[T]) error {
	var allErrors []error
	for _, l := range s.allLifecycleElement {
		if err := l.Open(ctx); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return errors.Join(allErrors...)
}

func (ms *concatenatedStream[T]) Close() {
	if len(ms.streams) == 0 {
		return
	}
	// Close only the current stream
	closeSubStream(ms.streams[0])
}

func closeSubStream[T any](s *stream[T]) {
	for _, l := range s.allLifecycleElement {
		l.Close()
	}
}

func (ms *concatenatedStream[T]) Emit(ctx context.Context) (*T, error) {

	currStreamNextItem, err := ms.streams[0].provider(ctx)
	if err != nil {
		if err == io.EOF {
			// If current stream is done, close it and continue with the next one
			if len(ms.streams) > 1 {
				closeSubStream(ms.streams[0])
				// Remove the current stream from the list
				ms.streams = ms.streams[1:]
				// Open the next stream
				err = openSubStream(ctx, ms.streams[0])
				// If the next stream is not opened, return the error
				if err != nil {
					// Cleanup since we are done, no need to keep the streams no need to close
					ms.streams = make([]*stream[T], 0)
					return nil, err
				}

				return ms.Emit(ctx)
			} else {
				// This is the last stream, return EOF
				return nil, err
			}
		} else {
			// This is an error, not EOF
			return nil, err
		}
	}
	return currStreamNextItem, nil

}
