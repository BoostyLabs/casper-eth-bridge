package math_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"tricorn/internal/math"
)

func TestMath(t *testing.T) {
	preparedValues := []struct {
		from           uint64
		to             uint64
		splitByNumber  uint64
		expectedResult []math.Range
	}{
		{
			from:          5,
			to:            10,
			splitByNumber: 2,
			expectedResult: []math.Range{
				{5, 7},
				{7, 9},
				{9, 10},
			},
		},
		{
			from:          10,
			to:            5,
			splitByNumber: 2,
		},
		{
			from:          5,
			to:            10,
			splitByNumber: 0,
		},
		{
			from:          5,
			to:            10,
			splitByNumber: 20,
			expectedResult: []math.Range{
				{5, 10},
			},
		},
	}

	t.Run("split range", func(t *testing.T) {
		result, err := math.SplitRange(math.Range{preparedValues[0].from, preparedValues[0].to}, preparedValues[0].splitByNumber)
		require.NoError(t, err)
		require.Equal(t, preparedValues[0].expectedResult, result)
	})

	t.Run("negative split range", func(t *testing.T) {
		result, err := math.SplitRange(math.Range{preparedValues[1].from, preparedValues[1].to}, preparedValues[1].splitByNumber)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "failed, to less than from"))
		require.Empty(t, result)
	})

	t.Run("negative split range", func(t *testing.T) {
		result, err := math.SplitRange(math.Range{preparedValues[2].from, preparedValues[2].to}, preparedValues[2].splitByNumber)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "failed, split by number equal to 0"))
		require.Empty(t, result)
	})

	t.Run("split range", func(t *testing.T) {
		result, err := math.SplitRange(math.Range{preparedValues[3].from, preparedValues[3].to}, preparedValues[3].splitByNumber)
		require.NoError(t, err)
		require.Equal(t, preparedValues[3].expectedResult, result)
	})
}
