package shpanstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"shpankids/infra/util/functional"
)

type StreamLifecycle interface {
	Open(ctx context.Context) error
	Close()
}

type streamLifecycleWrapper struct {
	openFunc  func(ctx context.Context) error
	closeFunc func()
}

func NewStreamLifecycle(openFunc func(ctx context.Context) error, closeFunc func()) StreamLifecycle {
	return &streamLifecycleWrapper{openFunc: openFunc, closeFunc: closeFunc}
}

func (s *streamLifecycleWrapper) Open(ctx context.Context) error {
	if s.openFunc != nil {
		return s.openFunc(ctx)
	}
	return nil
}

func (s *streamLifecycleWrapper) Close() {
	if s.closeFunc != nil {
		s.closeFunc()
	}
}

type StreamProvider[T any] interface {
	StreamLifecycle
	Emit(ctx context.Context) (*T, error)
}

type stream[T any] struct {
	provider            StreamProviderFunc[T]
	allLifecycleElement []StreamLifecycle
}

type Stream[T any] interface {
	Consume(ctx context.Context, f func(*T)) error
	ConsumeWithErr(ctx context.Context, f func(*T) error) error
	FilterWithError(predicate func(context.Context, *T) (bool, error)) Stream[T]
	Filter(predicate func(*T) bool) Stream[T]
	Limit(limit int) Stream[T]
	Skip(limit int) Stream[T]

	FindFirst(ctx context.Context) (Optional[T], error)
	GetFirst(ctx context.Context) (*T, error)
	FindLast(ctx context.Context) (Optional[T], error)
	Collect(ctx context.Context) ([]*T, error)
	SubscribeOnStreamLifecycle(lch StreamLifecycle) Stream[T]
}

func NewStream[T any](provider StreamProvider[T]) Stream[T] {
	return newStream(provider.Emit, []StreamLifecycle{provider})
}

func newStream[T any](streamProviderFunc StreamProviderFunc[T], allLifecycleElement []StreamLifecycle) Stream[T] {
	return &stream[T]{provider: streamProviderFunc, allLifecycleElement: allLifecycleElement}
}

func NewSimpleStream[T any](streamProviderFunc StreamProviderFunc[T]) Stream[T] {
	return &stream[T]{provider: streamProviderFunc, allLifecycleElement: nil}
}

type StreamProviderFunc[T any] func(ctx context.Context) (*T, error)

func (s *stream[T]) Consume(ctx context.Context, f func(*T)) error {
	return s.ConsumeWithErr(ctx, func(v *T) error {
		f(v)
		return nil
	})
}

func (s *stream[T]) ConsumeWithErr(ctx context.Context, f func(*T) error) error {

	ctxWithCancel, cancelFunc := context.WithCancel(ctx)
	// Running all lifecycle elements
	err := errors.Join(functional.MapSliceNoErr(s.allLifecycleElement, func(l StreamLifecycle) error {
		return l.Open(ctxWithCancel)
	})...)

	defer func() {
		cancelFunc()
		for _, l := range s.allLifecycleElement {
			l.Close()
		}
	}()

	if err != nil {
		return err
	}

	for {
		v, err := s.provider(ctxWithCancel)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = f(v)
		if err != nil {
			return err
		}
	}
}

func MapStream[SRC any, TGT any](src Stream[SRC], mapper func(*SRC) *TGT) Stream[TGT] {
	return MapStreamWithError(src, func(ctx context.Context, src *SRC) (*TGT, error) {
		return mapper(src), nil
	})
}
func FlatMapStream[SRC any, TGT any](src Stream[SRC], mapper func(*SRC) Stream[TGT]) Stream[TGT] {
	streamOfStreams := MapStreamWithError[SRC, Stream[TGT]](src, func(ctx context.Context, src *SRC) (*Stream[TGT], error) {
		s := mapper(src)
		return &s, nil
	})

	collect, _ := streamOfStreams.Collect(context.Background())
	return ConcatenatedStream[TGT](functional.MapSliceUnPtr(collect)...)
}

func MapStreamWithError[SRC any, TGT any](srcS Stream[SRC], mapper func(context.Context, *SRC) (*TGT, error)) Stream[TGT] {
	src, ok := srcS.(*stream[SRC])
	if !ok {
		slog.Error("Failed to cast Stream to stream")
		return nil
	}
	return newStream[TGT](
		func(ctx context.Context) (*TGT, error) {
			v, err := src.provider(ctx)
			if err != nil {
				return nil, err
			}
			return mapper(ctx, v)
		}, src.allLifecycleElement,
	)
}

func (s *stream[T]) GetFirst(ctx context.Context) (*T, error) {
	first, err := s.FindFirst(ctx)
	if err != nil {
		return nil, err
	}
	get, err := first.Get()
	if err != nil {
		return nil, err
	}
	return &get, nil

}
func (s *stream[T]) FindFirst(ctx context.Context) (Optional[T], error) {
	ctxWitCancel, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	var result *T
	var found bool
	err := s.Consume(ctxWitCancel, func(v *T) {
		result = v
		found = true
		cancelFunc()
	})
	// Checkin found first, because if found err will be "context canceled"
	if found {
		return NewOptional(*result), nil
	}
	if err != nil {
		return nil, err
	}

	return EmptyOptional[T](), nil

}

func (s *stream[T]) FindLast(ctx context.Context) (Optional[T], error) {
	var result *T
	var found bool
	err := s.Consume(ctx, func(v *T) {
		result = v
		found = true
	})
	// Checkin found first, because if found err will be "context canceled"
	if found {
		return NewOptional(*result), nil
	}
	if err != nil {
		return nil, err
	}
	return EmptyOptional[T](), nil

}

func (s *stream[T]) Filter(predicate func(*T) bool) Stream[T] {
	return s.FilterWithError(func(ctx context.Context, v *T) (bool, error) {
		return predicate(v), nil
	})
}

func (s *stream[T]) Collect(ctx context.Context) ([]*T, error) {
	var result []*T
	err := s.Consume(ctx, func(v *T) {
		result = append(result, v)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *stream[T]) FilterWithError(predicate func(context.Context, *T) (bool, error)) Stream[T] {
	return newStream[T](func(ctx context.Context) (*T, error) {
		for {
			v, err := s.provider(ctx)
			if err != nil {
				return nil, err
			}
			shouldKeep, err := predicate(ctx, v)
			if err != nil {
				// Wrapping errors, e.g. we don't want EOF accidentally returned from here
				return nil, fmt.Errorf("filter failed for stream: %w", err)
			}
			if shouldKeep {
				return v, nil
			}
		}
	}, s.allLifecycleElement)
}

func (s *stream[T]) Limit(limit int) Stream[T] {
	if limit <= 0 {
		return EmptyStream[T]()
	}
	alreadyConsumed := 1
	return newStream[T](func(ctx context.Context) (*T, error) {
		for {
			if alreadyConsumed > limit {
				return nil, io.EOF
			}

			v, err := s.provider(ctx)
			if err != nil {
				return nil, err
			}
			alreadyConsumed++
			return v, nil
		}
	}, s.allLifecycleElement)
}

func (s *stream[T]) Skip(skip int) Stream[T] {
	alreadySkipped := false
	return newStream[T](func(ctx context.Context) (*T, error) {
		if !alreadySkipped {
			alreadySkipped = true
			for i := 0; i < skip; i++ {
				_, err := s.provider(ctx)
				if err != nil {
					return nil, err
				}
			}
		}
		return s.provider(ctx)

	}, s.allLifecycleElement)
}

func (s *stream[T]) SubscribeOnStreamLifecycle(lch StreamLifecycle) Stream[T] {
	s.allLifecycleElement = append(s.allLifecycleElement, lch)
	return s
}
