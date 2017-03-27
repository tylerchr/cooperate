// Package cooperate is a library for implementing Operational Transformation.
package cooperate

import (
	"errors"
	"reflect"
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
	// A Document is data that may be collaboratively edited via
	// a series of distributed Operations.
	Document interface {
		Apply(Operation) error
	}

	// An Operation is a list of component actions that together define one
	// complete iteration through a document.
	Operation []Action

	// An Action is something that can be performed as part of an operation.
	Action interface{}

	// A Composer provides an application-aware implementation of the OT compose
	// function. Mathematically, it's defined as
	//
	//   Compose(a, b) ≡ a ◦ b
	//
	// In other words, given two operations a and b, it is expected to produce a
	// single operation c that produces the same effect as applying a then b.
	Composer interface {
		Compose(a, b OperationIterator) (Operation, error)
	}

	// A Transformer provides an application-aware implementation of the OT transform
	// function by producing the operational inverses of a and b. Mathematically:
	//
	//   transform(a, b) = (a', b') where a ◦ b' ≡ b ◦ a'
	//
	// Implementations of Transformer are expected to behave such that the above
	// equality holds.
	Transformer interface {
		Transform(a, b OperationIterator) (aa, bb Operation, err error)
	}

	// A ComposeTransformer implements the two core OT functions.
	ComposeTransformer interface {
		Composer
		Transformer
	}

	// An ExpandReducer converts an Operation to its most and least verbose forms,
	// respectively. It may help simplify implementations of Compose and Transform.
	ExpandReducer interface {
		// Expand inflates a single action to its atomic parts.
		Expand(a Action) []Action

		// Reduce merges actions a and b, or indicates that they are unmergable. The
		// arguments are provided such that the action a happens before action b.
		Reduce(a, b Action) (Action, bool)
	}
)

// An OperationIterator provides an interface to the actions within an Operation
// that helps simplify Composer and Transformer implementations.
type OperationIterator struct {
	Cursor  int
	Actions []Action
}

func NewOperationIterator(op Operation) *OperationIterator {
	return &OperationIterator{Actions: []Action(op)}
}

// More indicates whether any unconsumed actions remain.
func (oit *OperationIterator) More() bool {
	return oit.Len() > 0
}

// Len reports the number of unconsumed actions remaining.
func (oit *OperationIterator) Len() int {
	return len(oit.Actions) - oit.Cursor
}

// Peek returns the foremost action. It panics if none remain.
func (oit *OperationIterator) Peek() Action {
	if oit.Cursor >= len(oit.Actions) {
		panic(errors.New("no actions remain in iterator"))
	}
	return oit.Actions[oit.Cursor]
}

// PeekType returns the reflect.Type of the foremost action, or nil if none remain.
func (oit *OperationIterator) PeekType() reflect.Type {
	if oit.More() {
		return reflect.TypeOf(oit.Peek())
	}
	return nil
}

// Consume advances the iterator over foremost action and returns it. It panics if
// no actions remain.
func (oit *OperationIterator) Consume() Action {
	if oit.Cursor >= len(oit.Actions) {
		panic(errors.New("no actions remain in iterator"))
	}
	oit.Cursor++
	return oit.Actions[oit.Cursor-1]
}

// Expand inflates op such that each action affects only one element. For
// example, a RetainAction(6) becomes six consecutive RetainAction(1).
func Expand(er ExpandReducer, op Operation) Operation {
	var actions []Action
	for _, a := range []Action(op) {
		actions = append(actions, er.Expand(a)...)
	}
	return Operation(actions)
}

// Reduce collapses op into the minumum possible number of actions that have
// an identical effect.
func Reduce(er ExpandReducer, op Operation) Operation {
	var actions []Action
	for _, a := range []Action(op) {
		if len(actions) == 0 {
			actions = append(actions, a)
		} else if reduced, ok := er.Reduce(actions[len(actions)-1], a); ok {
			actions[len(actions)-1] = reduced
		} else {
			actions = append(actions, a)
		}
	}
	return Operation(actions)
}
