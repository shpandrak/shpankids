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

	GetAllForDate(ctx context.Context, forDate Date) shpanstream.Stream[functional.Entry[string, T]]
	GetRangeForKey(ctx context.Context, from Date, to Date, key string) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]]
	GetRange(ctx context.Context, from Date, to Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]]
}
