# Go iterate!

The package is a library to work with iterators.
It contains primitive implementations to iterate over the embedded language types.
Also, it allows to generate an iterator kit for your custom types.

### Usage

```go
package main

import (
    "fmt"
	
    iter "github.com/soven/go-iterate"
)

// FibonacciIter is an implementation of iter.IntIterator
// iterating over fibonacci numbers.
type FibonacciIter struct {
    cur, next int
    i, n int
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

func main() {
    fib10 := FibonacciN(10)
    fib10Doubled := DoubleInts(fib10)
    fmt.Println(fib10Doubled) // 0, 2, 2, 4, 6, 10, 16, 26, 42, 68
}

func DoubleInts(vv iter.IntIterator) []int {
    // Set doubling converter.
    vv = iter.IntConverting(vv, iter.IntConvert(func(v int) (int, error) {
        return v*2, nil
    }))
    
    // Unroll the iterator to a slice of int. 
    return iter.IntUnroll(vv)
}
```