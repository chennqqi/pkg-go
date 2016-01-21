package pkgmemoize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFibonacciEMemoFunc(t *testing.T) {
	calls := 0
	fibonacci := NewEMemoFunc(
		func(i int, eMemoFunc EMemoFunc) (interface{}, error) {
			calls++
			if i == 0 {
				return uint64(0), nil
			}
			if i == 1 {
				return uint64(1), nil
			}
			n1, err := eMemoFunc.Do(i - 1)
			if err != nil {
				return 0, err
			}
			n2, err := eMemoFunc.Do(i - 2)
			if err != nil {
				return 0, err
			}
			return n1.(uint64) + n2.(uint64), nil
		},
	)
	result, err := fibonacci.Do(93)
	require.NoError(t, err)
	require.Equal(t, uint64(12200160415121876738), result.(uint64))
	require.Equal(t, 94, calls)
}

func TestFibonacciMemoFunc(t *testing.T) {
	calls := 0
	fibonacci := NewMemoFunc(
		func(i int, memoFunc MemoFunc) interface{} {
			calls++
			if i == 0 {
				return uint64(0)
			}
			if i == 1 {
				return uint64(1)
			}
			return memoFunc.Do(i-1).(uint64) + memoFunc.Do(i-2).(uint64)
		},
	)
	require.Equal(t, uint64(12200160415121876738), fibonacci.Do(93))
	require.Equal(t, 94, calls)
}
