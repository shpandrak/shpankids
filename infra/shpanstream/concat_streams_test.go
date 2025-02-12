package shpanstream

import (
	"context"
	"github.com/stretchr/testify/require"
	"shpankids/infra/util/functional"
	"testing"
)

func TestConcatenatedStream(t *testing.T) {
	ctx := context.Background()

	// Create constant streams using Just
	stream1 := Just(1, 2, 3)
	stream2 := Just(4, 5)
	stream3 := EmptyStream[int]()

	concStream := ConcatenatedStream(stream1, stream2, stream3)

	// Assert the results directly
	expected := []int{1, 2, 3, 4, 5}

	// Collect results from the concatenated stream
	results, err := concStream.Collect(ctx)
	require.NoError(t, err)
	require.EqualValues(t, expected, functional.MapSliceUnPtr(results))
}
