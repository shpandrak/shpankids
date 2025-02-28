package datekvs

import (
	"context"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
)

type DateKvStore[T any] interface {
	Set(ctx context.Context, forDate Date, key string, value T) error
	Unset(ctx context.Context, forDate Date, key string) error
	Get(ctx context.Context, forDate Date, key string) (T, error)
	Find(ctx context.Context, forDate Date, key string) (*T, error)
	ManipulateOrCreate(ctx context.Context, forDate Date, key string, manipulator func(*T) (T, error)) error

	Stream(ctx context.Context) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]]
	StreamAllForDate(ctx context.Context, forDate Date) shpanstream.Stream[functional.Entry[string, T]]
	StreamRangeForKey(ctx context.Context, from Date, to Date, key string) shpanstream.Stream[DatedRecord[T]]
	StreamAllEntriesForKey(ctx context.Context, key string) shpanstream.Stream[DatedRecord[T]]
	StreamRange(ctx context.Context, from Date, to Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]]
}
