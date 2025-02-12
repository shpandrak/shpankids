package shpanstream

import "context"

type DownStreamProviderFunc[S any, T any] func(ctx context.Context, srcProviderFunc StreamProviderFunc[S]) (*T, error)

func NewDownStream[S any, T any](
	src Stream[S],
	downStreamProviderFunc DownStreamProviderFunc[S, T],
	openFunc func(ctx context.Context, srcProviderFunc StreamProviderFunc[S]) error,
) Stream[T] {
	cs := src.(*stream[S])

	dsLifecycle := NewStreamLifecycle(
		func(ctx context.Context) error {
			err := openSubStream(ctx, cs)
			if err != nil {
				return err
			}
			return openFunc(ctx, cs.provider)
		}, func() {
			closeSubStream(cs)
		})
	return newStream[T](
		func(ctx context.Context) (*T, error) {
			return downStreamProviderFunc(ctx, cs.provider)
		},
		[]StreamLifecycle{
			dsLifecycle,
		},
	)
}
