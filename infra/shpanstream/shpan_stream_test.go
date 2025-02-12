package shpanstream

import (
	"context"
	"github.com/stretchr/testify/require"
	"shpankids/infra/util/functional"
	"testing"
)

func TestBasicStreamsYey(t *testing.T) {
	collect, err :=
		Just(1, 2, 3, 4, 5).
			Filter(func(i *int) bool {
				return i != nil && *i > 2
			}).
			Collect(context.Background())

	require.NoError(t, err)
	require.Equal(t, 3, len(collect))

	collect, err =
		MapStream(Just(1, 2, 3, 4, 5), func(i *int) *int {
			if i != nil {
				return functional.ValueToPointer(*i + 1)
			} else {
				return nil
			}
		}).
			Filter(func(i *int) bool {
				return i != nil && *i > 2
			}).
			Collect(context.Background())

	require.NoError(t, err)
	require.Equal(t, 4, len(collect))

}
