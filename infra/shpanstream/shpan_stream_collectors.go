package shpanstream

import "context"

func CollectCountGroupedBy[T any](ctx context.Context, s Stream[T], grouper func(*T) string) (map[string]uint64, error) {
	result := make(map[string]uint64)
	err := s.Consume(ctx, func(v *T) {
		key := grouper(v)
		result[key]++
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
