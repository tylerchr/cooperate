package cooperate

import "fmt"

type Server struct {
	Document Document
	History  History

	// these implement the core OT operations
	ExpandReducer
	ComposeTransformer
}

// Apply applies the received Operation.
func (s *Server) Apply(root int, op Operation) error {

	// we need some way of knowing which state the operation is rooted at,

	// then we need to look up everything since that state,
	// compose it all together,
	var meanwhile Operation
	s.History.Iterate(root, func(seqno int, op Operation) (err error) {
		if meanwhile == nil {
			meanwhile = op
		} else {
			meanwhile, err = s.Compose(
				NewOperationIterator(Expand(s.ExpandReducer, meanwhile)),
				NewOperationIterator(Expand(s.ExpandReducer, op)),
			)
		}
		return
	})

	// transform op against it,
	if meanwhile != nil {
		_, opPrime, err := s.Transform(
			NewOperationIterator(Expand(s.ExpandReducer, meanwhile)),
			NewOperationIterator(Expand(s.ExpandReducer, op)),
		)
		if err != nil {
			return err
		}
		op = opPrime
	}

	// apply op' to our copy of the state,
	if err := s.Document.Apply(op); err != nil {
		return err
	}

	// save op' to the history,
	if _, err := s.History.Store(op); err != nil {
		return err
	}

	// and finally broadcast op' to everyone.
	fmt.Printf("hey everyone apply this: %#v\n", op)

	return nil

}
