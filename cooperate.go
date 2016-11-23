package cooperate

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	// ErrDocumentSizeMismatch indicates that the operation does not account for
	// the entire length of the existing document.
	ErrDocumentSizeMismatch = errors.New("document size mismatch")

	// ErrDeleteMismatch indicates that an action requested to delete data that
	// was not present in the existing document.
	ErrDeleteMismatch = errors.New("delete mismatch")
)

type (
	// A document is a document.
	Document []byte

	// An Operation is a list of component actions that together define one
	// complete iteration through a document.
	Operation []Action

	// An Action is something that can be performed as part of an operation.
	Action interface {
		isAction()
	}

	// RetainAction moves the cursor forward a specified number of elements
	RetainAction int

	// InsertAction inserts the given string at the current location
	InsertAction string

	// DeleteAction asserts that the given string immediately follows the
	// cursor, and then removes it.
	DeleteAction string
)

func (RetainAction) isAction() {}
func (InsertAction) isAction() {}
func (DeleteAction) isAction() {}

func (a RetainAction) GoString() string { return fmt.Sprintf("R(%d)", a) }
func (a InsertAction) GoString() string { return fmt.Sprintf("I(%s)", a) }
func (a DeleteAction) GoString() string { return fmt.Sprintf("D(%s)", a) }

// lengths calculates the lengths of the document op expects to be applied to
// and the length of that document after applying op.
func lengths(op Operation) (pre, post int) {
	for _, a := range []Action(op) {
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

// Expand inflates op such that each action affects only one element. For
// example, a RetainAction(6) becomes six consecutive RetainAction(1).
func Expand(op Operation) Operation {

	var actions []Action

	for _, a := range []Action(op) {
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
	}

	return Operation(actions)

}

// Reduce collapses op into the minumum possible number of actions that have
// an identical effect.
func Reduce(op Operation) Operation {

	var actions []Action

	for _, a := range []Action(op) {
		if len(actions) == 0 {
			actions = append(actions, a)
			continue
		}

		// pop off the latest action
		lastAction := actions[len(actions)-1]

		switch a := a.(type) {
		case InsertAction:
			switch l := lastAction.(type) {
			case InsertAction:
				// merge them!
				actions[len(actions)-1] = InsertAction(string(l) + string(a))
			default:
				// just add it
				actions = append(actions, a)
			}
		case DeleteAction:
			switch l := lastAction.(type) {
			case DeleteAction:
				// merge them!
				actions[len(actions)-1] = DeleteAction(string(l) + string(a))
			default:
				// just add it
				actions = append(actions, a)
			}
		case RetainAction:
			switch l := lastAction.(type) {
			case RetainAction:
				// merge them!
				actions[len(actions)-1] = RetainAction(int(l) + int(a))
			default:
				// just add it
				actions = append(actions, a)
			}
		}
	}

	return Operation(actions)

}

// Compose merges a and b into a single operation c such that the effect of
// applying c is equal to that of applying a then b.
func Compose(a, b Operation) (Operation, error) {

	// first, it must be the case that post(a) = pre(b)
	_, post := lengths(a)
	pre, _ := lengths(b)
	if post != pre {
		return Operation{}, ErrDocumentSizeMismatch
	}

	// expand them both, as composition is easier one element at a time
	aa, bb := []Action(Expand(a)), []Action(Expand(b))

	var ca, cb int               // cursors for each array
	var composedActions []Action // new list of actions

	for {

		// if we are out of actions from B, we are finished
		if cb == len(bb) {
			break
		}

		// if we are out of actions from A, there might still be inserts in B
		if ca == len(aa) {
			if s, ok := bb[cb].(InsertAction); ok {
				composedActions = append(composedActions, s)
				cb++
				continue
			}
			break
		}

		first, second := aa[ca], bb[cb]

		switch f := first.(type) {
		case InsertAction:
			switch s := second.(type) {
			case InsertAction:
				composedActions = append(composedActions, s, f)
				ca++
				cb++
			case DeleteAction:
				if string(f) != string(s) {
					panic(ErrDeleteMismatch)
				}
				ca++
				cb++
			case RetainAction:
				composedActions = append(composedActions, f)
				ca++
				cb++
			}
		case DeleteAction:

			// It happens to be the case that we could do
			//
			//   composedActions = append(composedActions, f)
			//   ca++
			//
			// for any of these cases where the first action is a delete, and the
			// second action would still get accounted for correctly next round.
			//
			// There is also an equally-true inverse property: whenever the second
			// operation is an insert, it can be handled in a consistent way.
			//
			// We could exploit these properties to simplify this handling
			// code, if we wanted to.

			switch s := second.(type) {
			case InsertAction:
				composedActions = append(composedActions, f, s)
				ca++
				cb++
			case DeleteAction:
				composedActions = append(composedActions, f, s)
				ca++
				cb++
			case RetainAction:
				composedActions = append(composedActions, f)
				ca++
			}
		case RetainAction:
			switch s := second.(type) {
			case InsertAction:
				composedActions = append(composedActions, s)
				cb++
			case DeleteAction:
				composedActions = append(composedActions, s)
				ca++
				cb++
			case RetainAction:
				composedActions = append(composedActions, f)
				ca++
				cb++
			}
		}
	}

	// a document size mismatch occurs if we didn't process everything
	if len(aa) != ca || len(bb) != cb {
		return nil, ErrDocumentSizeMismatch
	}

	return Reduce(Operation(composedActions)), nil
}

// Transform takes operations a and b, and produces operations a' and b' such that
// the following equality holds:
//
//   Compose(a, b') == Compose(b, a')
//
// In other words, if operations a and b happen concurrently on a single parent,
// Transform generates operations a' and b' that will bring the divergent states
// back in sync if applied (to b and a, respectively).
//
// This implementation of Transform favors b; that is, if a and b both act on the
// same element, the effect is as though b's intended change was applied first.
func Transform(a, b Operation) (aPrime, bPrime Operation, err error) {

	// Since a and b must share a common parent (parent, not ancestor), it must
	// be the case that their expectant lengths match.
	preA, _ := lengths(a)
	preB, _ := lengths(b)
	if preA != preB {
		return nil, nil, ErrDocumentSizeMismatch
	}

	// expand them both, as composition is easier one element at a time
	actionsA, actionsB := []Action(Expand(a)), []Action(Expand(b))

	var ca, cb int      // cursors for each array
	var aa, bb []Action // new list of actions (become aPrime and bPrime respectively)

	for {

		// if we reach the ends at the same time, we are done
		if ca == len(actionsA) && cb == len(actionsB) {
			break
		}

		// if we are at the end of one but not the other, that "other" must contain inserts
		if ca == len(actionsA) || cb == len(actionsB) {

			// a better have inserts
			if ca < len(actionsA) {
				if first, ok := actionsA[ca].(InsertAction); ok {
					aa = append(aa, first)
					bb = append(bb, RetainAction(1))
					ca++
					continue
				}
			}

			// b better have inserts
			if cb < len(actionsB) {
				if second, ok := actionsB[cb].(InsertAction); ok {
					aa = append(aa, RetainAction(1))
					bb = append(bb, second)
					ca++
					continue
				}

			}

			return nil, nil, ErrDocumentSizeMismatch
		}

		first, second := actionsA[ca], actionsB[cb]

		switch f := first.(type) {
		case InsertAction:
			switch s := second.(type) {
			case InsertAction:
				// InsertInsert
				aa = append(aa, RetainAction(1), f)
				bb = append(bb, s, RetainAction(1))
				ca++
				cb++
			case DeleteAction:
				// InsertDelete
				aa = append(aa, f)
				bb = append(bb, RetainAction(1))
				ca++
			case RetainAction:
				// InsertRetain
				aa = append(aa, f)
				bb = append(bb, RetainAction(1))
				ca++
			}
		case DeleteAction:
			switch s := second.(type) {
			case InsertAction:
				// DeleteInsert
				aa = append(aa, RetainAction(1))
				bb = append(bb, s)
				cb++
			case DeleteAction:
				// DeleteDelete
				ca++
				cb++
			case RetainAction:
				// DeleteRetain
				aa = append(aa, f)
				ca++
				cb++
			}
		case RetainAction:
			switch s := second.(type) {
			case InsertAction:
				// RetainInsert
				aa = append(aa, RetainAction(1))
				bb = append(bb, s)
				cb++
			case DeleteAction:
				// RetainDelete
				bb = append(aa, s)
				ca++
				cb++
			case RetainAction:
				// RetainRetain
				aa = append(aa, f)
				bb = append(bb, s)
				ca++
				cb++
			}
		}

	}

	return Reduce(Operation(aa)), Reduce(Operation(bb)), nil

}

// appliesCleanly indicates whether the operation op is valid against the current
// state of doc.
func appliesCleanly(doc []byte, op Operation) error {

	// assert that op's assumed length matches doc's length
	if pre, _ := lengths(op); pre != len(doc) {
		return ErrDocumentSizeMismatch
	}

	// ensure that any deletes in op align with current data in doc
	var cursor int
	for _, a := range []Action(op) {
		switch a := a.(type) {
		case InsertAction: // no change to verification cursor
		case DeleteAction:
			if string(doc[cursor:cursor+len(string(a))]) != string(a) {
				return ErrDeleteMismatch
			}
			cursor += len(string(a))
		case RetainAction:
			cursor += int(a)
		}
	}

	return nil

}

// Apply performs operation op against the document doc, or returns an error
// if op is invalid given the current state of doc.
func Apply(doc []byte, op Operation) ([]byte, error) {

	// assert that the operation will apply cleanly before mutating the doc
	if err := appliesCleanly(doc, op); err != nil {
		return nil, err
	}

	var cursor int

	for _, action := range []Action(op) {
		switch action := action.(type) {
		case RetainAction:
			cursor += int(action)
		case InsertAction:
			var buf bytes.Buffer
			buf.Write(doc[:cursor])
			buf.WriteString(string(action))
			buf.Write(doc[cursor:])
			doc = buf.Bytes()
			cursor += len(string(action))
		case DeleteAction:
			if string(doc[cursor:cursor+len(string(action))]) != string(action) {
				return nil, ErrDeleteMismatch
			}
			copy(doc[cursor:], doc[cursor+len(string(action)):])
			doc = doc[:len(doc)-len(string(action))]

		}
	}

	if cursor != len(doc) {
		return nil, ErrDocumentSizeMismatch
	}

	return doc, nil

}
