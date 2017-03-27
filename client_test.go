package cooperate_test

import (
	"reflect"
	"testing"

	"github.com/tylerchr/cooperate"
	"github.com/tylerchr/cooperate/text"
)

func TestClient_Buffering(t *testing.T) {

	cases := []struct {
		ExistingDocument string
		Operations       []cooperate.Operation
		ExpectedDocument string
		ExpectedInFlight cooperate.Operation
		ExpectedBuffer   cooperate.Operation
	}{
		// first operation applied is immediately placed in-flight;
		// future operations are held in the buffer
		{
			ExistingDocument: "",
			Operations: []cooperate.Operation{
				cooperate.Operation([]cooperate.Action{
					text.InsertAction("lorem"),
				}),
				cooperate.Operation([]cooperate.Action{
					text.RetainAction(5),
					text.InsertAction(" ipsum"),
				}),
			},
			ExpectedDocument: "lorem ipsum",
			ExpectedInFlight: cooperate.Operation([]cooperate.Action{
				text.InsertAction("lorem"),
			}),
			ExpectedBuffer: cooperate.Operation([]cooperate.Action{
				text.RetainAction(5),
				text.InsertAction(" ipsum"),
			}),
		},

		// buffer composition
		{
			ExistingDocument: "",
			Operations: []cooperate.Operation{
				cooperate.Operation([]cooperate.Action{
					text.InsertAction("lorem"),
				}),
				cooperate.Operation([]cooperate.Action{
					text.RetainAction(5),
					text.InsertAction(" ipsum"),
				}),
				cooperate.Operation([]cooperate.Action{
					text.RetainAction(11),
					text.InsertAction(" dolor"),
				}),
			},
			ExpectedDocument: "lorem ipsum dolor",
			ExpectedInFlight: cooperate.Operation([]cooperate.Action{
				text.InsertAction("lorem"),
			}),
			ExpectedBuffer: cooperate.Operation([]cooperate.Action{
				text.RetainAction(5),
				text.InsertAction(" ipsum dolor"),
			}),
		},
	}

	for i, c := range cases {

		client := &cooperate.Client{
			Document:           text.NewTextDocument(c.ExistingDocument),
			ExpandReducer:      text.TextHandler{},
			ComposeTransformer: text.TextHandler{},
		}

		for _, op := range c.Operations {
			client.ApplyLocal(op)
		}

		if client.Document.(*text.TextDocument).String() != c.ExpectedDocument {
			t.Errorf("[case %d] unexpected document: expected '%s' but got '%s'", i, c.ExpectedDocument, client.Document.(*text.TextDocument).String())
		}

		if !reflect.DeepEqual(client.InFlight, c.ExpectedInFlight) {
			t.Errorf("[case %d] unexpected in-flight operation: expected '%s' but got '%s'", i, c.ExpectedInFlight, client.InFlight)
		}

		if !reflect.DeepEqual(client.Buffer, c.ExpectedBuffer) {
			t.Errorf("[case %d] unexpected buffered operation: expected '%s' but got '%s'", i, c.ExpectedBuffer, client.Buffer)
		}
	}

}

func TestClient_ApplyReceived(t *testing.T) {

	cases := []struct {
		ExistingDocument string
		ExistingInFlight cooperate.Operation
		ExistingBuffer   cooperate.Operation

		Operations []cooperate.Operation

		ExpectedDocument string
		ExpectedInFlight cooperate.Operation
		ExpectedBuffer   cooperate.Operation
	}{
		// first operation applied is immediately placed in-flight;
		// future operations are held in the buffer
		{
			ExistingDocument: "redblue",
			ExistingInFlight: cooperate.Operation([]cooperate.Action{
				text.InsertAction("red"),
			}),
			ExistingBuffer: cooperate.Operation([]cooperate.Action{
				text.RetainAction(3),
				text.InsertAction("blue"),
			}),
			Operations: []cooperate.Operation{
				cooperate.Operation([]cooperate.Action{
					text.InsertAction("green"),
				}),
			},
			ExpectedDocument: "greenredblue",
			ExpectedInFlight: cooperate.Operation([]cooperate.Action{
				text.RetainAction(5),
				text.InsertAction("red"),
			}),
			ExpectedBuffer: cooperate.Operation([]cooperate.Action{
				text.RetainAction(8),
				text.InsertAction("blue"),
			}),
		},
	}

	for i, c := range cases {

		client := &cooperate.Client{
			Document: text.NewTextDocument(c.ExistingDocument),
			InFlight: c.ExistingInFlight,
			Buffer:   c.ExistingBuffer,

			ExpandReducer:      text.TextHandler{},
			ComposeTransformer: text.TextHandler{},
		}

		for _, op := range c.Operations {
			client.ApplyReceived(op)
		}

		if client.Document.(*text.TextDocument).String() != c.ExpectedDocument {
			t.Errorf("[case %d] unexpected document: expected '%s' but got '%s'", i, c.ExpectedDocument, client.Document.(*text.TextDocument).String())
		}

		if !reflect.DeepEqual(client.InFlight, c.ExpectedInFlight) {
			t.Errorf("[case %d] unexpected in-flight operation: expected '%s' but got '%s'", i, c.ExpectedInFlight, client.InFlight)
		}

		if !reflect.DeepEqual(client.Buffer, c.ExpectedBuffer) {
			t.Errorf("[case %d] unexpected buffered operation: expected '%s' but got '%s'", i, c.ExpectedBuffer, client.Buffer)
		}
	}

}
