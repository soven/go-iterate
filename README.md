# Go iterate!

[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/soven/go-iterate#section-documentation)


The package is a library to work with iterators.
It contains primitive implementations to iterate over the language embedded types.

Also, it allows to generate an iterator kit for your custom types.

### Install 

```shell
go get "github.com/soven/go-iterate"
cd $GOPATH/src/github.com/soven/go-iterate
make install

# Make sure $GOPATH/bin in your $PATH
gen-go-iter-kit version
# Gen Go Iter KIT v0.0.3
```

### The kit generator usage
```shell
# Put your custom type configuration in place of the next macros (angle bracket inclusive):
# <type_name> - name of your type as is.
# <prefix_name> - prefix for your type in generated identifier names. 
#    Usually that is the same as <type_name> but capitalized.
# <zero_type_value> - zero value of your type. 
#    For example 0 is zero value for int, "" is zero value for string.
#    You should pre define a zero value for your type as well.
# <package_name> - package name which should be like the package of your type.
# <path_to_package> - path to the package dir.
gen-go-iter-kit -target="<type_name>" -prefix="<prefix_name>" \
  -zero="<zero_type_value>" -package="<package_name>" -path="<path_to_package>"
```

### The library usage

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