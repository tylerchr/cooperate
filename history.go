package cooperate

type (
	// History implements storage of a sequence of Operations.
	History interface {
		// SequenceNumber returns the sequence number of the current state.
		SequenceNumber() int

		// Store appends an operation to the history, and returns its seqno.
		Store(op Operation) (seqno int, err error)

		// Iterate traverses through all operations between startingSeqno and
		// SequenceNumber() inclusive.
		Iterate(startingSeqno int, cb func(seqno int, op Operation) error) error
	}

	// MemoryHistory is the simplest possible History implementation, storing
	// a sequence of Operations in an in-memory slice.
	MemoryHistory []Operation
)

func (mh *MemoryHistory) SequenceNumber() int {
	return len(*mh)
}

func (mh *MemoryHistory) Store(op Operation) (int, error) {
	*mh = append(*mh, op)
	return mh.SequenceNumber(), nil
}

func (mh *MemoryHistory) Iterate(startingSeqno int, cb func(seqno int, op Operation) error) error {
	for i := startingSeqno; i < len(*mh); i++ {
		if err := cb(i, (*mh)[i]); err != nil {
			return err
		}
	}
	return nil
}
