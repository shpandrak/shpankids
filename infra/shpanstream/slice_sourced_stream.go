package shpanstream

import (
	"context"
	"io"
)

func Just[T any](slice ...T) Stream[T] {
	return newStream[T](func(ctx context.Context) (*T, error) {
		if len(slice) == 0 {
			return nil, io.EOF
		}
		v := slice[0]
		slice = slice[1:]
		return &v, nil
	}, nil)
}

func JustPtr[T any](slice []*T) Stream[T] {
	return newStream[T](func(ctx context.Context) (*T, error) {
		if len(slice) == 0 {
			return nil, io.EOF
		}
		v := slice[0]
		slice = slice[1:]
		return v, nil
	}, nil)
}
