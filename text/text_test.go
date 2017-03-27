package text

import (
	"reflect"
	"testing"

	"github.com/tylerchr/cooperate"
)

func TestReduce(t *testing.T) {

	cases := []struct {
		First, Second cooperate.Action
		Result        cooperate.Action
		Mergeable     bool
	}{
		{
			First:     RetainAction(1),
			Second:    InsertAction("a"),
			Mergeable: false,
		},
		{
			First:     InsertAction("a"),
			Second:    DeleteAction("a"),
			Mergeable: false,
		},
		{
			First:     RetainAction(1),
			Second:    RetainAction(2),
			Result:    RetainAction(3),
			Mergeable: true,
		},
		{
			First:     InsertAction("a"),
			Second:    InsertAction("b"),
			Result:    InsertAction("ab"),
			Mergeable: true,
		},
		{
			First:     DeleteAction("a"),
			Second:    DeleteAction("b"),
			Result:    DeleteAction("ab"),
			Mergeable: true,
		},
	}

	var th TextHandler

	for i, c := range cases {

		action, ok := th.Reduce(c.First, c.Second)
		if ok != c.Mergeable {
			t.Errorf("[case %d] unexpected mergeability result: expected %t but got %t", i, c.Mergeable, ok)
		} else if !reflect.DeepEqual(action, c.Result) {
			t.Errorf("[case %d] unexpected reduction: expected '%#v' but got '%#v'", i, c.Result, action)
		}

	}

}

func TestLengths(t *testing.T) {

	cases := []struct {
		Operation cooperate.Operation
		Before    int
		After     int
	}{
		{
			Operation: cooperate.Operation([]cooperate.Action{}),
			Before:    0,
			After:     0,
		},
		{
			Operation: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
			}),
			Before: 1,
			After:  1,
		},
		{
			Operation: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				InsertAction("foo"),
				RetainAction(2),
			}),
			Before: 3,
			After:  6,
		},
		{
			Operation: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				DeleteAction("foo"),
			}),
			Before: 4,
			After:  1,
		},
	}

	for i, c := range cases {
		if pre, post := Lengths(c.Operation); pre != c.Before || post != c.After {
			t.Errorf("[case %d] unexpected prelength: expected (%d -> %d) but got (%d -> %d)", i, c.Before, c.After, pre, post)
		}
	}

}

func TestCompose(t *testing.T) {

	cases := []struct {
		First       cooperate.Operation
		Second      cooperate.Operation
		Composition cooperate.Operation
		Error       error
	}{
		{
			First: cooperate.Operation([]cooperate.Action{
				InsertAction("foo"),
			}),
			Second: cooperate.Operation([]cooperate.Action{
				DeleteAction("foo"),
			}),
			Composition: nil,
		},
		{
			First: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				InsertAction("l"),
				RetainAction(2),
			}),
			Second: cooperate.Operation([]cooperate.Action{
				RetainAction(2),
				InsertAction("e"),
				RetainAction(2),
			}),
			Composition: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				InsertAction("le"),
				RetainAction(2),
			}),
		},
	}

	var th TextHandler

	for i, c := range cases {

		aa, bb := cooperate.NewOperationIterator(cooperate.Expand(th, c.First)), cooperate.NewOperationIterator(cooperate.Expand(th, c.Second))

		if sum, err := th.Compose(aa, bb); err != c.Error {
			t.Errorf("[case %d] unexpected error state: expected '%s' but got '%s'", i, c.Error, err)
		} else if !reflect.DeepEqual(sum, c.Composition) {
			t.Errorf("[case %d] unexpected composition: expected '%#v' but got '%#v'", i, c.Composition, sum)
		}
	}

}

func TestTransform(t *testing.T) {

	cases := []struct {
		A, B           cooperate.Operation
		APrime, BPrime cooperate.Operation
		Error          error
	}{
		{
			A: cooperate.Operation([]cooperate.Action{
				RetainAction(2),
				InsertAction("t"),
			}),
			B: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				InsertAction("ro"),
				RetainAction(1),
			}),
			APrime: cooperate.Operation([]cooperate.Action{
				RetainAction(4),
				InsertAction("t"),
			}),
			BPrime: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				InsertAction("ro"),
				RetainAction(2),
			}),
		},
		{
			A: cooperate.Operation([]cooperate.Action{
				RetainAction(2),
				InsertAction("t"),
			}),
			B: cooperate.Operation([]cooperate.Action{
				RetainAction(2),
				InsertAction("a"),
			}),
			APrime: cooperate.Operation([]cooperate.Action{
				RetainAction(3),
				InsertAction("t"),
			}),
			BPrime: cooperate.Operation([]cooperate.Action{
				RetainAction(2),
				InsertAction("a"),
				RetainAction(1),
			}),
		},
	}

	var th TextHandler

	for i, c := range cases {

		a, b := cooperate.NewOperationIterator(cooperate.Expand(th, c.A)), cooperate.NewOperationIterator(cooperate.Expand(th, c.B))

		if aPrime, bPrime, err := th.Transform(a, b); err != c.Error {
			t.Errorf("[case %d] unexpected error state: expected '%s' but got '%s'", i, c.Error, err)
		} else if !reflect.DeepEqual(aPrime, c.APrime) || !reflect.DeepEqual(bPrime, c.BPrime) {
			t.Errorf("[case %d] unexpected transformation: expected (a':%#v, b1:%#v) but got (a':%#v, b':%#v)", i, c.APrime, c.BPrime, aPrime, bPrime)
		}
	}

}
