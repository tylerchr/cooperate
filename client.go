package cooperate

import "fmt"

// A Client serves as the editing entity in the OT paradigm. It maintains
// its own independent state and proposes updates to some OT server.
type Client struct {
	Document Document

	InFlight Operation
	Buffer   Operation

	// these implement the core OT operations
	ExpandReducer
	ComposeTransformer
}

// ApplyLocal applies an operation that this client produced. If no pending
// operations exist, this operation is immediately proposed; otherwise it is
// composed into the buffer and held for future proposal.
func (c *Client) ApplyLocal(op Operation) error {

	// apply the transformation to the document
	if err := c.Document.Apply(op); err != nil {
		return err
	}

	switch {
	case c.InFlight == nil:
		c.InFlight = op
		fmt.Printf("[sent op] %#v\n", c.InFlight)

	case c.InFlight != nil && c.Buffer == nil:
		c.Buffer = op
		fmt.Printf("[set buffer] %#v\n", c.Buffer)

	case c.InFlight != nil && c.Buffer != nil:
		composedOp, err := c.ComposeTransformer.Compose(NewOperationIterator(Expand(c.ExpandReducer, c.Buffer)), NewOperationIterator(Expand(c.ExpandReducer, op)))
		if err != nil {
			return err
		}
		c.Buffer = Reduce(c.ExpandReducer, composedOp)
		fmt.Printf("[composed into buffer] %#v\n", c.Buffer)
	}

	return nil

}

// ApplyReceived transforms an Operation received from the server for local
// application and adapts InFlight and Buffer accordingly.
func (c *Client) ApplyReceived(op Operation) error {

	// TODO(tylerchr): We are currently assuming that an inflight and buffer exist. What to do if one/both don't?

	// transform op against inflight --> this is our new inflight + temp state
	// transform temp state against buffer --> this is our new buffer

	if_aa, if_bb, err := c.Transform(NewOperationIterator(Expand(c.ExpandReducer, c.InFlight)), NewOperationIterator(Expand(c.ExpandReducer, op)))
	// fmt.Printf(">> %#v %#v %s\n", if_aa, if_bb, err)
	if err != nil {
		return err
	}

	// a' is our new InFlight operation
	c.InFlight = if_aa
	fmt.Printf("[inflight] %#v\n", c.InFlight)

	// b' is now useful for transforming the buffer
	buf_aa, buf_bb, err := c.Transform(NewOperationIterator(Expand(c.ExpandReducer, c.Buffer)), NewOperationIterator(Expand(c.ExpandReducer, if_bb)))
	if err != nil {
		return err
	}

	// buf_aa is our new Buffer operation
	c.Buffer = buf_aa
	fmt.Printf("[buffer] %#v\n", c.Buffer)

	// buf_bb is the operation we should apply to our document
	fmt.Printf("[apply] %#v\n", buf_bb)

	if err := c.Document.Apply(buf_bb); err != nil {
		return err
	}

	fmt.Printf("[document] %#v\n", c.Document)

	return nil

}
