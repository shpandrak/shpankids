package shpanstream

import (
	"context"
	"io"
)

func EmptyStream[T any]() Stream[T] {
	return newStream(func(ctx context.Context) (*T, error) {
		return nil, io.EOF
	}, nil)
}
