package bfserr_test

import (
	"errors"
	"fmt"

	"github.com/Brian-Williams/bfserr"
)

type customErr struct{ msg string }

func (e customErr) Error() string { return e.msg }

func ExampleAsType() {
	// Create a tree of errors
	err := fmt.Errorf("wrap 1: %w",
		fmt.Errorf("wrap 2: %w",
			customErr{"target too deep"},
		),
	)
	err2 := fmt.Errorf("wrap 1: %w", customErr{"target"})
	joinedErr := errors.Join(err, err2)

	// Use bfserr.AsType to search for a specific type of error in the tree
	// finding the shallowest matching error first
	if e, ok := bfserr.AsType[customErr](joinedErr); ok {
		fmt.Println("bfserr found:", e)
	}

	// Use errors.As to search for a specific type of error in the tree
	// finding the deepest matching error first
	var errAs customErr
	if errors.As(errors.Join(err, err2), &errAs) {
		fmt.Println("errors.As found:", errAs)
	}

	// Output:
	// bfserr found: target
	// errors.As found: target too deep
}
