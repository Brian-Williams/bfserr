# bfserr

[![Go Reference](https://pkg.go.dev/badge/github.com/Brian-Williams/bfserr.svg)](https://pkg.go.dev/github.com/Brian-Williams/bfserr)
[![Go Tests](https://github.com/Brian-Williams/bfserr/actions/workflows/test.yml/badge.svg)](https://github.com/Brian-Williams/bfserr/actions)

A Go package that provides **breadth-first error search** utilities.  
It extends the standard `errors.As` semantics to traverse
error trees *breadth-first* instead of *depth-first*.

## Features

- **Breadth-first traversal** of wrapped error trees
- Works with both `Unwrap() error` and `Unwrap() []error`
- Compatible with Go’s `errors.As` and custom `As(any) bool` methods
- Ideal for finding “higher-level” causes before deep leaves

---

## Example
The following example will set e as `customErr{"target"}`
```go
func ExampleAsType() {
	err := fmt.Errorf("wrap 1: %w",
		fmt.Errorf("wrap 2: %w",
			customErr{"target too deep"},
		),
	)
	err2 := fmt.Errorf("wrap 1: %w", customErr{"target"})

	if e, ok := bfserr.AsType[customErr](errors.Join(err, err2)); ok {
		fmt.Println("found:", e)
	}
}
```

---

## Installation

```bash
go get github.com/Brian-Williams/bfserr
```

---

## Performance

If you can use [`errors.As`](https://pkg.go.dev/errors#As) you should. In most common error trees it will outperform `bfserr.AsType`.

In the case where you need a bfs of an error tree this will have comparable or better performance to [`errors`](https://pkg.go.dev/errors).
```
BenchmarkAs-10                   8333600               121.9 ns/op
BenchmarkAsType-10               7887480               149.2 ns/op
BenchmarkAsDeep-10               5703589               203.5 ns/op            40 B/op          2 allocs/op
BenchmarkAsTypeDeep-10          12341869                96.14 ns/op          104 B/op          4 allocs/op
```