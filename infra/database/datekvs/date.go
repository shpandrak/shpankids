package datekvs

import (
	"context"
	"io"
	"shpankids/infra/shpanstream"
	"time"
)

type Date struct {
	time.Time
}

//goland:noinspection GoMixedReceiverTypes
func (d Date) MarshalJSON() ([]byte, error) {
	return d.Time.MarshalJSON()
}

//goland:noinspection GoMixedReceiverTypes
func (d *Date) UnmarshalJSON(data []byte) error {
	return d.Time.UnmarshalJSON(data)
}

type DatedRecord[T any] struct {
	Date  Date
	Value T
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func TodayDate(location *time.Location) Date {
	return *NewDateFromTime(time.Now().In(location))
}
func NewDateFromTime(t time.Time) *Date {
	return &Date{time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

//goland:noinspection GoMixedReceiverTypes
func (d Date) String() string {
	return d.Format(time.DateOnly)
}

//goland:noinspection GoMixedReceiverTypes
func (d Date) DateEndTime() time.Time {
	return d.AddDay().Time
}

//goland:noinspection GoMixedReceiverTypes
func (d Date) AddDay() Date {
	return Date{d.AddDate(0, 0, 1)}
}

func NewDateRangeStream(from Date, to Date) shpanstream.Stream[Date] {
	curr := from
	return shpanstream.NewSimpleStream(func(ctx context.Context) (*Date, error) {
		if curr.After(to.Time) {
			return nil, io.EOF
		}
		ret := curr
		curr = curr.AddDay()
		return &ret, nil
	})
}
