package datekvs

import (
	"context"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"time"
)

type Date struct {
	time.Time
}

type DatedRecord[T any] struct {
	Date  Date
	Value T
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}
func NewDateFromTime(t time.Time) *Date {
	return &Date{time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

func (d Date) String() string {
	return d.Format(time.DateOnly)
}
func (d Date) AddDay() Date {
	return Date{d.AddDate(0, 0, 1)}
}

type DateKvStore[T any] interface {
	Set(ctx context.Context, forDate Date, key string, value T) error
	Unset(ctx context.Context, forDate Date, key string) error
	Get(ctx context.Context, forDate Date, key string) (T, error)
	Find(ctx context.Context, forDate Date, key string) (*T, error)

	GetAllForDate(ctx context.Context, forDate Date) shpanstream.Stream[functional.Entry[string, T]]
	GetRangeForKey(ctx context.Context, from Date, to Date, key string) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]]
	GetRange(ctx context.Context, from Date, to Date) shpanstream.Stream[DatedRecord[functional.Entry[string, T]]]
}
