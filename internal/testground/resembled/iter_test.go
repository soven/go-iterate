package resembled_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	iter "github.com/soven/go-iterate"
)

// FibonacciIter is an implementation of iter.IntIterator
// iterating over fibonacci numbers.
type FibonacciIter struct {
	cur, next int
	i, n      int
}

// FibonacciN returns an iterator for n first fibonacci numbers.
func FibonacciN(n int) *FibonacciIter {
	return &FibonacciIter{cur: 0, next: 1, n: n}
}

func (fib FibonacciIter) HasNext() bool { return fib.i < fib.n }

func (fib *FibonacciIter) Next() int {
	ret := fib.cur
	fib.cur, fib.next = fib.next, fib.cur+fib.next
	fib.i++
	return ret
}

func (fib FibonacciIter) Err() error { return nil }

func Test_Example_Fibonacci(t *testing.T) {
	fib10 := FibonacciN(10)
	fib10Doubled := DoubleInts(fib10)
	assert.EqualValues(t, []int{0, 2, 2, 4, 6, 10, 16, 26, 42, 68}, fib10Doubled)
}

func DoubleInts(vv iter.IntIterator) []int {
	// Set doubling converter.
	vv = iter.IntConverting(vv, iter.IntConvert(func(v int) (int, error) {
		return v * 2, nil
	}))

	// Unroll the iterator to a slice of int.
	return iter.IntUnroll(vv)
}
