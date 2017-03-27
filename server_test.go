package cooperate_test

import (
	"testing"

	"github.com/tylerchr/cooperate"
	"github.com/tylerchr/cooperate/text"
)

func TestServer(t *testing.T) {

	s := &cooperate.Server{
		Document: text.NewTextDocument(""),
		History:  &cooperate.MemoryHistory{},
	}

	ops := []struct {
		Root      int
		Operation cooperate.Operation
	}{
		{
			Root: 0,
			Operation: cooperate.Operation([]cooperate.Action{
				text.InsertAction("foo"),
			}),
		},
		{
			Root: 1,
			Operation: cooperate.Operation([]cooperate.Action{
				text.RetainAction(3),
				text.InsertAction(" "),
			}),
		},
		{
			Root: 2,
			Operation: cooperate.Operation([]cooperate.Action{
				text.RetainAction(4),
				text.InsertAction("bar"),
			}),
		},
	}

	for _, op := range ops {
		if err := s.Apply(op.Root, op.Operation); err != nil {
			t.Fatalf("apply error: %s", err)
		}
	}

	if doc := s.Document.(*text.TextDocument); doc.String() != "foo bar" {
		t.Errorf("unexpected document: %s\n", doc.String())
	}

}

func TestServer_OutOfOrder(t *testing.T) {

	s := &cooperate.Server{
		Document:           text.NewTextDocument(""),
		History:            &cooperate.MemoryHistory{},
		ExpandReducer:      &text.TextHandler{},
		ComposeTransformer: &text.TextHandler{},
	}

	ops := []struct {
		Root      int
		Operation cooperate.Operation
	}{
		{
			Root: 0,
			Operation: cooperate.Operation([]cooperate.Action{
				text.InsertAction("red"),
			}),
		},
		{
			Root: 1,
			Operation: cooperate.Operation([]cooperate.Action{
				text.RetainAction(3),
				text.InsertAction("blue"),
			}),
		},
		{
			Root: 0,
			Operation: cooperate.Operation([]cooperate.Action{
				text.InsertAction("green"),
			}),
		},
	}

	for _, op := range ops {
		if err := s.Apply(op.Root, op.Operation); err != nil {
			t.Fatalf("apply error: %s", err)
		}
	}

	if doc := s.Document.(*text.TextDocument); doc.String() != "greenredblue" {
		t.Errorf("unexpected document: %s\n", doc.String())
	}

	t.Logf("Server {seqno=%d}\n", s.History.SequenceNumber())

}
