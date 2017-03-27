package text

import (
	"fmt"
	"reflect"

	"github.com/tylerchr/cooperate"
)

var (
	Retain = reflect.TypeOf(RetainAction(0))
	Insert = reflect.TypeOf(InsertAction(""))
	Delete = reflect.TypeOf(DeleteAction(""))
)

type (
	// A TextHandler implements cooperate.ComposeTransformer and cooperate.ExpandReducer
	// for the standard text-based operations: retain, insert, and delete.
	TextHandler struct{}

	// RetainAction moves the cursor forward a specified number of elements
	RetainAction int

	// InsertAction inserts the given string at the current location
	InsertAction string

	// DeleteAction asserts that the given string immediately follows the
	// cursor, and then removes it.
	DeleteAction string
)

func (a RetainAction) GoString() string { return fmt.Sprintf("R(%d)", a) }
func (a InsertAction) GoString() string { return fmt.Sprintf("I(%s)", a) }
func (a DeleteAction) GoString() string { return fmt.Sprintf("D(%s)", a) }

func (th TextHandler) Expand(a cooperate.Action) []cooperate.Action {
	var actions []cooperate.Action
	switch a := a.(type) {
	case RetainAction:
		for i := 0; i < int(a); i++ {
			actions = append(actions, RetainAction(1))
		}
	case InsertAction:
		for _, x := range string(a) {
			actions = append(actions, InsertAction(x))
		}
	case DeleteAction:
		for _, x := range string(a) {
			actions = append(actions, DeleteAction(x))
		}
	}
	return actions
}

// Reduce combines two actions of identical type into a single action with the same
// effect, or indicates whether that the action could not be accomplished.
//
// Reduce will always fail if a and b are not of identical types.
func (th TextHandler) Reduce(a, b cooperate.Action) (cooperate.Action, bool) {

	// if the actions aren't the same type, then they can't be reduced
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		// fmt.Printf("not same type: %s != %s\n", reflect.TypeOf(a), reflect.TypeOf(b))
		return nil, false
	}

	// TODO(tylerchr): Check the type of b, because this solution will panic on bad input.

	switch a.(type) {
	case InsertAction:
		return InsertAction(string(a.(InsertAction)) + string(b.(InsertAction))), true
	case DeleteAction:
		return DeleteAction(string(a.(DeleteAction)) + string(b.(DeleteAction))), true
	case RetainAction:
		return RetainAction(int(a.(RetainAction)) + int(b.(RetainAction))), true
	}

	return nil, false // we can't merge actions we can't identify
}

// Compose merges a and b into a single operation c such that the effect of
// applying c is equal to that of applying a then b.
func (th TextHandler) Compose(a, b *cooperate.OperationIterator) (cooperate.Operation, error) {

	var composedActions []cooperate.Action // new list of actions

ComposeLoop:
	for {

		switch {

		// if we are out of actions from B, we are finished
		case !b.More():
			break ComposeLoop

		case !a.More():
			if b.PeekType() == Insert {
				composedActions = append(composedActions, b.Consume())
				continue ComposeLoop
			}
			break ComposeLoop

		// // optimization case
		// // anytime the first action is a delete, we can apply it immediately
		// case a.PeekType() == Delete:
		// 	composedActions = append(composedActions, a.Consume())

		// // optimization case
		// // anytime the second action is an insert, we can apply it immediately
		// case b.PeekType() == Insert:
		// 	composedActions = append(composedActions, b.Consume())

		case a.PeekType() == Insert && b.PeekType() == Insert:
			composedActions = append(composedActions, b.Consume(), a.Consume())

		case a.PeekType() == Insert && b.PeekType() == Delete:
			f, s := a.Peek(), b.Peek()
			if string(f.(InsertAction)) != string(s.(DeleteAction)) {
				panic(cooperate.ErrDeleteMismatch)
			}
			a.Consume()
			b.Consume()

		case a.PeekType() == Insert && b.PeekType() == Retain:
			composedActions = append(composedActions, a.Consume())
			b.Consume()

		case a.PeekType() == Delete && b.PeekType() == Insert:
			composedActions = append(composedActions, a.Consume(), b.Consume())

		case a.PeekType() == Delete && b.PeekType() == Delete:
			composedActions = append(composedActions, a.Consume(), b.Consume())

		case a.PeekType() == Delete && b.PeekType() == Retain:
			composedActions = append(composedActions, a.Consume())

		case a.PeekType() == Retain && b.PeekType() == Insert:
			composedActions = append(composedActions, b.Consume())

		case a.PeekType() == Retain && b.PeekType() == Delete:
			a.Consume()
			composedActions = append(composedActions, b.Consume())

		case a.PeekType() == Retain && b.PeekType() == Retain:
			composedActions = append(composedActions, a.Consume())
			b.Consume()
		}
	}

	// a document size mismatch occurs if we didn't process everything
	if a.More() || b.More() {
		return nil, cooperate.ErrDocumentSizeMismatch
	}

	return cooperate.Reduce(th, cooperate.Operation(composedActions)), nil
}

// Transform implements cooperate.Transformer for the basic text operations
// defined in this package.
//
// This implementation favors b; that is, if a and b both act on the
// same element, the effect is as though b's intended change was applied first.
func (th TextHandler) Transform(a, b *cooperate.OperationIterator) (aa, bb cooperate.Operation, err error) {

	var aPrime, bPrime []cooperate.Action // new list of actions

TransformLoop:
	for {

		switch {

		// if we reach the ends at the same time, we are done
		case !a.More() && !b.More():
			break TransformLoop

		// if we are at the end of b but not a, then a must contain only inserts
		case a.More() && !b.More():
			if a.PeekType() == Insert {
				aPrime = append(aPrime, a.Consume())
				bPrime = append(bPrime, RetainAction(1))
				continue
			}
			break TransformLoop

		// if we are at the end of a but not b, then b must contain only inserts
		case !a.More() && b.More():
			if b.PeekType() == Insert {
				aPrime = append(aPrime, RetainAction(1))
				bPrime = append(bPrime, b.Consume())
				continue
			}
			break TransformLoop

		case a.PeekType() == Insert && b.PeekType() == Insert:
			aPrime = append(aPrime, RetainAction(1))
			bPrime = append(bPrime, b.Consume())

		case a.PeekType() == Insert && b.PeekType() == Delete:
			aPrime = append(aPrime, a.Consume())
			bPrime = append(bPrime, RetainAction(1))

		case a.PeekType() == Insert && b.PeekType() == Retain:
			aPrime = append(aPrime, a.Consume())
			bPrime = append(bPrime, RetainAction(1))

		case a.PeekType() == Delete && b.PeekType() == Insert:
			aPrime = append(aPrime, RetainAction(1))
			bPrime = append(bPrime, b.Consume())

		case a.PeekType() == Delete && b.PeekType() == Delete:
			a.Consume()
			b.Consume()

		case a.PeekType() == Delete && b.PeekType() == Retain:
			aPrime = append(aPrime, a.Consume())
			b.Consume()

		case a.PeekType() == Retain && b.PeekType() == Insert:
			aPrime = append(aPrime, RetainAction(1))
			bPrime = append(bPrime, b.Consume())

		case a.PeekType() == Retain && b.PeekType() == Delete:
			a.Consume()
			bPrime = append(bPrime, b.Consume())

		case a.PeekType() == Retain && b.PeekType() == Retain:
			aPrime = append(aPrime, a.Consume())
			bPrime = append(bPrime, b.Consume())

		}

	}

	// a document size mismatch occurs if we didn't process everything
	if a.More() || b.More() {
		return nil, nil, cooperate.ErrDocumentSizeMismatch
	}

	return cooperate.Reduce(th, cooperate.Operation(aPrime)), cooperate.Reduce(th, cooperate.Operation(bPrime)), nil

}

// Lengths calculates the lengths of the document op expects to be applied to
// and the length of that document after applying op.
func Lengths(op cooperate.Operation) (pre, post int) {
	for _, a := range []cooperate.Action(op) {
		switch a := a.(type) {
		case RetainAction:
			pre += int(a)
			post += int(a)
		case InsertAction:
			post += int(len(a))
		case DeleteAction:
			pre += int(len(a))
		}
	}
	return
}
