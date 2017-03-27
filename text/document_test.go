package text

import (
	"testing"

	"github.com/tylerchr/cooperate"
)

func TestTextDocument(t *testing.T) {

	cases := []struct {
		ExistingContents string
		Operation        cooperate.Operation
		ExpectedError    error
		ExpectedContents string
	}{
		{
			ExistingContents: "",
			Operation: cooperate.Operation([]cooperate.Action{
				InsertAction("foo"),
			}),
			ExpectedContents: "foo",
		},
		{
			ExistingContents: "foo",
			Operation: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				InsertAction("l"),
				RetainAction(2),
			}),
			ExpectedContents: "floo",
		},
		{
			ExistingContents: "spar",
			Operation: cooperate.Operation([]cooperate.Action{
				RetainAction(1),
				DeleteAction("p"),
				RetainAction(2),
			}),
			ExpectedContents: "sar",
		},
	}

	for i, c := range cases {

		doc := TextDocument{
			contents: c.ExistingContents,
		}

		if err := doc.Apply(c.Operation); err != c.ExpectedError {
			t.Errorf("[case %d] unexpected error: expected '%s' but got '%s'", i, c.ExpectedError, err)
		} else if doc.contents != c.ExpectedContents {
			t.Errorf("[case %d] unexpected document: expected '%s' but got '%s'", i, c.ExpectedContents, doc.contents)
		}
	}

}
