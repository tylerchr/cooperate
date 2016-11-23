package cooperate

import (
	"bytes"
	"reflect"
	"testing"
)

func TestLengths(t *testing.T) {

	cases := []struct {
		Operation Operation
		Before    int
		After     int
	}{
		{
			Operation: Operation([]Action{}),
			Before:    0,
			After:     0,
		},
		{
			Operation: Operation([]Action{
				RetainAction(1),
			}),
			Before: 1,
			After:  1,
		},
		{
			Operation: Operation([]Action{
				RetainAction(1),
				InsertAction("foo"),
				RetainAction(2),
			}),
			Before: 3,
			After:  6,
		},
		{
			Operation: Operation([]Action{
				RetainAction(1),
				DeleteAction("foo"),
			}),
			Before: 4,
			After:  1,
		},
	}

	for i, c := range cases {
		if pre, post := lengths(c.Operation); pre != c.Before || post != c.After {
			t.Errorf("[case %d] unexpected prelength: expected (%d -> %d) but got (%d -> %d)", i, c.Before, c.After, pre, post)
		}
	}

}

func TestExpand(t *testing.T) {

	cases := []struct {
		Original Operation
		Expanded Operation
	}{
		{
			Original: Operation([]Action{
				RetainAction(1),
			}),
			Expanded: Operation([]Action{
				RetainAction(1),
			}),
		},
		{
			Original: Operation([]Action{
				RetainAction(3),
			}),
			Expanded: Operation([]Action{
				RetainAction(1),
				RetainAction(1),
				RetainAction(1),
			}),
		},
		{
			Original: Operation([]Action{
				InsertAction("foo"),
			}),
			Expanded: Operation([]Action{
				InsertAction("f"),
				InsertAction("o"),
				InsertAction("o"),
			}),
		},
		{
			Original: Operation([]Action{
				DeleteAction("foo"),
			}),
			Expanded: Operation([]Action{
				DeleteAction("f"),
				DeleteAction("o"),
				DeleteAction("o"),
			}),
		},
	}

	for i, c := range cases {
		if exp := Expand(c.Original); !reflect.DeepEqual(exp, c.Expanded) {
			t.Errorf("[case %d] unexpected expansion: expected '%#v' but got '%#v'", i, c.Expanded, exp)
		}
	}

}

func TestReduce(t *testing.T) {

	cases := []struct {
		Original Operation
		Reduced  Operation
	}{
		{
			Original: Operation([]Action{
				RetainAction(1),
			}),
			Reduced: Operation([]Action{
				RetainAction(1),
			}),
		},
		{
			Original: Operation([]Action{
				RetainAction(1),
				RetainAction(1),
				RetainAction(1),
			}),
			Reduced: Operation([]Action{
				RetainAction(3),
			}),
		},
		{
			Original: Operation([]Action{
				InsertAction("f"),
				InsertAction("o"),
				InsertAction("o"),
			}),
			Reduced: Operation([]Action{
				InsertAction("foo"),
			}),
		},
		{
			Original: Operation([]Action{
				DeleteAction("f"),
				DeleteAction("o"),
				DeleteAction("o"),
			}),
			Reduced: Operation([]Action{
				DeleteAction("foo"),
			}),
		},
	}

	for i, c := range cases {
		if exp := Reduce(c.Original); !reflect.DeepEqual(exp, c.Reduced) {
			t.Errorf("[case %d] unexpected expansion: expected '%#v' but got '%#v'", i, c.Reduced, exp)
		}
	}

}

func TestCompose(t *testing.T) {

	cases := []struct {
		First       Operation
		Second      Operation
		Composition Operation
		Error       error
	}{
		{
			First: Operation([]Action{
				InsertAction("foo"),
			}),
			Second: Operation([]Action{
				DeleteAction("foo"),
			}),
			Composition: nil,
		},
		{
			First: Operation([]Action{
				RetainAction(1),
				InsertAction("l"),
				RetainAction(2),
			}),
			Second: Operation([]Action{
				RetainAction(2),
				InsertAction("e"),
				RetainAction(2),
			}),
			Composition: Operation([]Action{
				RetainAction(1),
				InsertAction("le"),
				RetainAction(2),
			}),
		},
	}

	for i, c := range cases {
		if sum, err := Compose(c.First, c.Second); err != c.Error {
			t.Errorf("[case %d] unexpected error state: expected '%s' but got '%s'", i, c.Error, err)
		} else if !reflect.DeepEqual(sum, c.Composition) {
			t.Errorf("[case %d] unexpected composition: expected '%#v' but got '%#v'", i, c.Composition, sum)
		}
	}

}

func TestTransform(t *testing.T) {

	cases := []struct {
		A, B           Operation
		APrime, BPrime Operation
		Error          error
	}{
		{
			A: Operation([]Action{
				RetainAction(2),
				InsertAction("t"),
			}),
			B: Operation([]Action{
				RetainAction(1),
				InsertAction("ro"),
				RetainAction(1),
			}),
			APrime: Operation([]Action{
				RetainAction(4),
				InsertAction("t"),
			}),
			BPrime: Operation([]Action{
				RetainAction(1),
				InsertAction("ro"),
				RetainAction(2),
			}),
		},
		{
			A: Operation([]Action{
				RetainAction(2),
				InsertAction("t"),
			}),
			B: Operation([]Action{
				RetainAction(2),
				InsertAction("a"),
			}),
			APrime: Operation([]Action{
				RetainAction(3),
				InsertAction("t"),
			}),
			BPrime: Operation([]Action{
				RetainAction(2),
				InsertAction("a"),
				RetainAction(1),
			}),
		},
	}

	for i, c := range cases {
		if aPrime, bPrime, err := Transform(c.A, c.B); err != c.Error {
			t.Errorf("[case %d] unexpected error state: expected '%s' but got '%s'", i, c.Error, err)
		} else if !reflect.DeepEqual(aPrime, c.APrime) || !reflect.DeepEqual(bPrime, c.BPrime) {
			t.Errorf("[case %d] unexpected transformation: expected (a':%#v, b1:%#v) but got (a':%#v, b':%#v)", i, c.APrime, c.BPrime, aPrime, bPrime)
		}
	}

}

func TestApply(t *testing.T) {

	cases := []struct {
		Base      []byte
		Operation Operation
		Target    []byte
		Error     error
	}{
		{
			Base: []byte("got"),
			Operation: Operation([]Action{
				RetainAction(3),
			}),
			Target: []byte("got"),
		},
		{
			Base: []byte(""),
			Operation: Operation([]Action{
				InsertAction("hello"),
			}),
			Target: []byte("hello"),
		},
		{
			Base: []byte("got"),
			Operation: Operation([]Action{
				DeleteAction("got"),
			}),
			Target: []byte(""),
		},
		{
			Base: []byte("got"),
			Operation: Operation([]Action{
				RetainAction(2),
				InsertAction("a"),
				RetainAction(1),
			}),
			Target: []byte("goat"),
		},
		{
			Base: []byte("goat"),
			Operation: Operation([]Action{
				RetainAction(2),
				DeleteAction("a"),
				RetainAction(1),
			}),
			Target: []byte("got"),
		},
		{
			Base: []byte("got"),
			Operation: Operation([]Action{
				RetainAction(2),
				InsertAction("a"),
			}),
			Error: ErrDocumentSizeMismatch,
		},
		{
			Base: []byte("goat"),
			Operation: Operation([]Action{
				RetainAction(3),
				DeleteAction("d"),
			}),
			Error: ErrDeleteMismatch,
		},

		{
			Base: []byte("it"),
			Operation: Operation([]Action{
				InsertAction("a"),
				DeleteAction("i"),
				RetainAction(1),
			}),
			Target: []byte("at"),
		},
		{
			Base: []byte("it"),
			Operation: Operation([]Action{
				DeleteAction("i"),
				InsertAction("a"),
				RetainAction(1),
			}),
			Target: []byte("at"),
		},
	}

	for i, c := range cases {

		if doc, err := Apply(c.Base, c.Operation); err != c.Error {
			t.Errorf("[case %d] unexpected error state: expected '%s' but got '%s'", i, c.Error, err)
		} else if c.Error == nil && !bytes.Equal(doc, c.Target) {
			t.Errorf("[case %d] unexpected result: expected '%s' but got '%s'", i, c.Target, doc)
		}

	}

}
