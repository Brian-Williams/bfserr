package bfserr_test

import (
	"errors"
	"fmt"

	"github.com/Brian-Williams/bfserr"
)

type customErr struct{ msg string }

func (e customErr) Error() string { return e.msg }

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
	// Output:
	// found: target
}
