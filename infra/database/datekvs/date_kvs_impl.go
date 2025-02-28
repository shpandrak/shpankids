package datekvs

import (
	"context"
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

func (d *dateKvsImpl[T]) StreamAllForDate(ctx context.Context, forDate Date) shpanstream.Stream[functional.Entry[string, T]] {
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

func (d *dateKvsImpl[T]) StreamAllEntriesForKey(ctx context.Context, key string) shpanstream.Stream[DatedRecord[T]] {
	return shpanstream.FlatMapStream(
		d.streamAllDateRecords(ctx),
		func(currDate *Date) shpanstream.Stream[DatedRecord[T]] {
			find, err := d.Find(ctx, *currDate, key)
			if err != nil {
				return shpanstream.NewErrorStream[DatedRecord[T]](err)
			}
			if find == nil {
				return shpanstream.EmptyStream[DatedRecord[T]]()
			} else {
				return shpanstream.Just(DatedRecord[T]{
					Date:  *currDate,
					Value: *find,
				})
			}
		},
	)
}

func (d *dateKvsImpl[T]) StreamRangeForKey(ctx context.Context, from Date, to Date, key string) shpanstream.Stream[DatedRecord[T]] {

	return shpanstream.FlatMapStream(
		d.streamAllDatesForDateRage(ctx, from, to),
		func(currDate *Date) shpanstream.Stream[DatedRecord[T]] {
			find, err := d.Find(ctx, *currDate, key)
			if err != nil {
				return shpanstream.NewErrorStream[DatedRecord[T]](err)
			}
			if find == nil {
				return shpanstream.EmptyStream[DatedRecord[T]]()
			} else {
				return shpanstream.Just(DatedRecord[T]{
					Date:  *currDate,
					Value: *find,
				})
			}
		},
	)
}

func (d *dateKvsImpl[T]) Stream(ctx context.Context) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
	return shpanstream.FlatMapStream[Date, DatedRecord[functional.Entry[string, T]]](
		d.streamAllDateRecords(ctx),
		func(ns *Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
			return shpanstream.MapStream[functional.Entry[string, T], DatedRecord[functional.Entry[string, T]]](
				d.getStoreForDate(ctx, *ns).Stream(ctx),
				func(entry *functional.Entry[string, T]) *DatedRecord[functional.Entry[string, T]] {
					return &DatedRecord[functional.Entry[string, T]]{*ns, *entry}
				})

		},
	)

}

func (d *dateKvsImpl[T]) streamAllDatesForDateRage(ctx context.Context, from Date, to Date) shpanstream.Stream[Date] {
	return d.streamAllDateRecords(ctx).Filter(func(tm *Date) bool {
		return !tm.Before(from.Time) && tm.Before(to.Time)
	})
}

func (d *dateKvsImpl[T]) streamAllDateRecords(ctx context.Context) shpanstream.Stream[Date] {
	return shpanstream.MapStreamWhileFiltering(
		d.rawKvs.StreamAllNamespaces(ctx),
		func(ns *string) *Date {
			tm, err := time.Parse(time.DateOnly, *ns)
			if err != nil {
				return nil
			}
			return NewDateFromTime(tm)
		},
	)
}

func (d *dateKvsImpl[T]) StreamRange(ctx context.Context, from Date, to Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
	return shpanstream.FlatMapStream[Date, DatedRecord[functional.Entry[string, T]]](
		d.streamAllDatesForDateRage(ctx, from, to),
		func(ns *Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]] {
			return shpanstream.MapStream[functional.Entry[string, T], DatedRecord[functional.Entry[string, T]]](
				d.getStoreForDate(ctx, *ns).Stream(ctx),
				func(entry *functional.Entry[string, T]) *DatedRecord[functional.Entry[string, T]] {
					return &DatedRecord[functional.Entry[string, T]]{*ns, *entry}
				})
		},
	)
}

func (d *dateKvsImpl[T]) ManipulateOrCreate(
	ctx context.Context,
	forDate Date,
	key string,
	manipulator func(*T) (T, error),
) error {
	return d.getStoreForDate(ctx, forDate).ManipulateOrCreate(ctx, key, manipulator)
}
