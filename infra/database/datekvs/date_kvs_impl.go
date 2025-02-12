package datekvs

import (
	"context"
	"fmt"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"time"
)

type dateKvsImpl[T any] struct {
	rawKvs kvstore.RawJsonStore
}

func NewDateKvsImpl[T any](rawKvs kvstore.RawJsonStore) DateKvStore[T] {
	return &dateKvsImpl[T]{rawKvs: rawKvs}
}

func (d *dateKvsImpl[T]) GetAllForDate(ctx context.Context, forDate Date) shpanstream.Stream[functional.Entry[string, T]] {
	return d.getStoreForDate(ctx, forDate).Stream(ctx)
}

func (d *dateKvsImpl[T]) getStoreForDate(_ context.Context, forDate Date) kvstore.JsonKvStore[string, T] {
	return kvstore.NewJsonKvStoreImpl[string, T](d.rawKvs, forDate.String(), kvstore.StringKeyToString, kvstore.StringToKey)
}

func (d *dateKvsImpl[T]) Set(ctx context.Context, forDate Date, key string, value T) error {
	return d.getStoreForDate(ctx, forDate).Set(ctx, key, value)
}

func (d *dateKvsImpl[T]) Unset(ctx context.Context, forDate Date, key string) error {
	return d.getStoreForDate(ctx, forDate).Unset(ctx, key)
}

func (d *dateKvsImpl[T]) Get(ctx context.Context, forDate Date, key string) (T, error) {
	return d.getStoreForDate(ctx, forDate).Get(ctx, key)
}

func (d *dateKvsImpl[T]) Find(ctx context.Context, forDate Date, key string) (*T, error) {
	return d.getStoreForDate(ctx, forDate).Find(ctx, key)
}

func (d *dateKvsImpl[T]) GetRangeForKey(ctx context.Context, from Date, to Date, key string) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
	return shpanstream.NewErrorStream[DatedRecord[functional.Entry[string, T]]](fmt.Errorf("not implemented"))
}

func (d *dateKvsImpl[T]) GetRange(ctx context.Context, from Date, to Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
	return shpanstream.FlatMapStream[Date, DatedRecord[functional.Entry[string, T]]](
		shpanstream.MapStreamWithError(
			d.rawKvs.StreamAllNamespaces(ctx),
			func(_ context.Context, ns *string) (*Date, error) {
				tm, err := time.Parse(time.DateOnly, *ns)
				if err != nil {
					return nil, err
				}
				return NewDateFromTime(tm), nil

			},
		).FilterWithError(func(ctx context.Context, tm *Date) (bool, error) {
			return !tm.Before(from.Time) && tm.Before(to.Time), nil
		}),
		func(ns *Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
			return shpanstream.MapStream[functional.Entry[string, T], DatedRecord[functional.Entry[string, T]]](
				d.getStoreForDate(ctx, *ns).Stream(ctx),
				func(entry *functional.Entry[string, T]) *DatedRecord[functional.Entry[string, T]] {
					return &DatedRecord[functional.Entry[string, T]]{*ns, *entry}
				})

		},
	)

}
