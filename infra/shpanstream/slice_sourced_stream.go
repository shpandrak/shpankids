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

func JustPtrFilterNils[T any](slice ...*T) Stream[T] {
	slicePtr := make([]*T, 0, len(slice))
	for _, v := range slice {
		if v != nil {
			slicePtr = append(slicePtr, v)
		}
	}
	return newStream[T](func(ctx context.Context) (*T, error) {
		if len(slicePtr) == 0 {
			return nil, io.EOF
		}
		v := slicePtr[0]
		slicePtr = slicePtr[1:]
		return v, nil
	}, nil)
}
