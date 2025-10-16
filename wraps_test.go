package bfserr_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"testing"

	"github.com/Brian-Williams/bfserr"
)

type poser struct {
	msg string
	f   func(error) bool
}

var poserPathErr = &fs.PathError{Op: "poser"}

func (p *poser) Error() string     { return p.msg }
func (p *poser) Is(err error) bool { return p.f(err) }
func (p *poser) As(err any) bool {
	switch x := err.(type) {
	case **poser:
		*x = p
	case *errorT:
		*x = errorT{"poser"}
	case **fs.PathError:
		*x = poserPathErr
	default:
		return false
	}
	return true
}

func TestAs(t *testing.T) {
	_, errF := os.Open("non-existing")
	poserErr := &poser{"oh no", nil}

	testCases := []struct {
		err    error
		asType func(error) (any, bool)
		match  bool
		want   any // value of target on match
	}{{
		nil,
		func(err error) (any, bool) { return bfserr.AsType[*fs.PathError](err) },
		false,
		nil,
	}, {
		wrapped{"pitied the fool", errorT{"T"}},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{"T"},
	}, {
		errF,
		func(err error) (any, bool) { return bfserr.AsType[*fs.PathError](err) },
		true,
		errF,
	}, {
		errorT{},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{},
	}, {
		errorT{},
		func(err error) (any, bool) { return bfserr.AsType[*fs.PathError](err) },
		false,
		nil,
	}, {
		wrapped{"wrapped", nil},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		false,
		nil,
	}, {
		&poser{"error", nil},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{"poser"},
	}, {
		&poser{"path", nil},
		func(err error) (any, bool) { return bfserr.AsType[*fs.PathError](err) },
		true,
		poserPathErr,
	}, {
		poserErr,
		func(err error) (any, bool) { return bfserr.AsType[*poser](err) },
		true,
		poserErr,
	}, {
		multiErr{},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		false,
		nil,
	}, {
		multiErr{errors.New("a"), errorT{"T"}},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{"T"},
	}, {
		multiErr{errorT{"T"}, errors.New("a")},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{"T"},
	}, {
		multiErr{errorT{"a"}, errorT{"b"}},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{"a"},
	}, {
		multiErr{multiErr{errors.New("a"), errorT{"a"}}, errorT{"b"}},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		true,
		errorT{"b"},
	}, {
		multiErr{nil},
		func(err error) (any, bool) { return bfserr.AsType[errorT](err) },
		false,
		nil,
	}}
	for i, tc := range testCases {
		name := fmt.Sprintf("%d:As(Errorf(..., %v))", i, tc.err)
		t.Run(name, func(t *testing.T) {
			val, match := tc.asType(tc.err)

			if match != tc.match {
				t.Fatalf("match: got %v; want %v", match, tc.match)
			}
			if !match {
				return
			}
			if !reflect.DeepEqual(val, tc.want) {
				t.Fatalf("got %#v, want %#v", val, tc.want)
			}
		})
	}

}

func TestAsValidation(t *testing.T) {
	var s string
	testCases := []any{
		nil,
		(*int)(nil),
		"error",
		&s,
	}
	err := errors.New("error")
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%T(%v)", tc, tc), func(t *testing.T) {
			if tc, ok := bfserr.AsType[errorT](err); ok {
				t.Errorf("As(err, %T(%v)) = true, want false", tc, tc)
				return
			}
		})
	}
}

func TestAsType(t *testing.T) {
	var errT errorT
	var errP *fs.PathError
	type timeout interface {
		Timeout() bool
		error
	}
	_, errF := os.Open("non-existing")
	poserErr := &poser{"oh no", nil}

	testAsType(t,
		nil,
		errP,
		false,
	)
	testAsType(t,
		wrapped{"pitied the fool", errorT{"T"}},
		errorT{"T"},
		true,
	)
	testAsType(t,
		errF,
		errF,
		true,
	)
	testAsType(t,
		errT,
		errP,
		false,
	)
	testAsType(t,
		wrapped{"wrapped", nil},
		errT,
		false,
	)
	testAsType(t,
		&poser{"error", nil},
		errorT{"poser"},
		true,
	)
	testAsType(t,
		&poser{"path", nil},
		poserPathErr,
		true,
	)
	testAsType(t,
		poserErr,
		poserErr,
		true,
	)
	testAsType(t,
		errors.New("err"),
		timeout(nil),
		false,
	)
	testAsType(t,
		errF,
		errF.(timeout),
		true)
	testAsType(t,
		wrapped{"path error", errF},
		errF.(timeout),
		true,
	)
	testAsType(t,
		multiErr{},
		errT,
		false,
	)
	testAsType(t,
		multiErr{errors.New("a"), errorT{"T"}},
		errorT{"T"},
		true,
	)
	testAsType(t,
		multiErr{errorT{"T"}, errors.New("a")},
		errorT{"T"},
		true,
	)
	testAsType(t,
		multiErr{errorT{"a"}, errorT{"b"}},
		errorT{"a"},
		true,
	)
	testAsType(t,
		multiErr{multiErr{errors.New("a"), errorT{"a"}}, errorT{"b"}},
		errorT{"b"},
		true,
	)
	testAsType(t,
		multiErr{multiErr{errors.New("a"), errorT{"a"}}, errors.New("b")},
		errorT{"a"},
		true,
	)
	testAsType(t,
		multiErr{
			multiErr{
				multiErr{errors.New("a"), errorT{"a"}},
				errorT{"b"},
			},
		},
		errorT{"b"},
		true,
	)
	testAsType(t,
		multiErr{
			multiErr{
				errors.New("a"),
				multiErr{errors.New("b"), errorT{"b"}},
				multiErr{errors.New("c"), errorT{"c"}},
			},
		},
		errorT{"b"},
		true,
	)
	testAsType(t,
		multiErr{wrapped{"path error", errF}},
		errF.(timeout),
		true,
	)
	testAsType(t,
		multiErr{nil},
		errT,
		false,
	)
}

type compError interface {
	comparable
	error
}

func testAsType[E compError](t *testing.T, err error, want E, wantOK bool) {
	t.Helper()
	name := fmt.Sprintf("AsType[%T](Errorf(..., %v))", want, err)
	t.Run(name, func(t *testing.T) {
		got, gotOK := bfserr.AsType[E](err)
		if gotOK != wantOK || got != want {
			t.Fatalf("got %v, %t; want %v, %t", got, gotOK, want, wantOK)
		}
	})
}

func BenchmarkAs(b *testing.B) {
	err := multiErr{multiErr{multiErr{errors.New("a"), errorT{"a"}}, errorT{"b"}}}
	for b.Loop() {
		var target errorT
		if !errors.As(err, &target) {
			b.Fatal("As failed")
		}
	}
}

// Uncomment when AsType is released
// func BenchmarkAsType(b *testing.B) {
// 	err := multiErr{multiErr{multiErr{errors.New("a"), errorT{"a"}}, errorT{"b"}}}
// 	for range b.N {
// 		if _, ok := errors.AsType[errorT](err); !ok {
// 			b.Fatal("AsType failed")
// 		}
// 	}
// }

func BenchmarkAsType(b *testing.B) {
	err := multiErr{multiErr{multiErr{errors.New("a"), errorT{"a"}}, errorT{"b"}}}
	for b.Loop() {
		if _, ok := bfserr.AsType[errorT](err); !ok {
			b.Fatal("AsType failed")
		}
	}
}

func BenchmarkAsDeep(b *testing.B) {
	// Deep chain (no match) + shallow matching sibling.
	err := multiErr{
		deepWrapped(12),
		errorT{"hit-shallow"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var target errorT
		if !errors.As(err, &target) || target != (errorT{"hit-shallow"}) {
			b.Fatal("As failed")
		}
	}
}

func BenchmarkAsTypeDeep(b *testing.B) {
	// Same structure as above.
	err := multiErr{
		deepWrapped(12),
		errorT{"hit-shallow"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e, ok := bfserr.AsType[errorT](err)
		if !ok || e != (errorT{"hit-shallow"}) {
			b.Fatal("AsType failed")
		}
	}
}

func deepWrapped(depth int) error {
	e := errors.New("leaf")
	for i := range depth {
		e = wrapped{fmt.Sprintf("wrap %d", i), e}
	}
	return e
}

type errorT struct{ s string }

func (e errorT) Error() string { return fmt.Sprintf("errorT(%s)", e.s) }

type wrapped struct {
	msg string
	err error
}

func (e wrapped) Error() string { return e.msg }
func (e wrapped) Unwrap() error { return e.err }

type multiErr []error

func (m multiErr) Error() string   { return "multiError" }
func (m multiErr) Unwrap() []error { return []error(m) }
