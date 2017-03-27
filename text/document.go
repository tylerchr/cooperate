package text

import (
	"fmt"

	"github.com/tylerchr/cooperate"
)

// A TextDocument is a string that implements the cooperate.Document interface.
type TextDocument struct {
	contents string
}

// NewTextDocument initializes a TextDocument with a starting value of initial.
func NewTextDocument(initial string) *TextDocument {
	return &TextDocument{
		contents: initial,
	}
}

// String returns the current contents of the document.
func (td *TextDocument) String() string {
	return td.contents
}

// Apply performs op against the TextDocument.
func (td *TextDocument) Apply(op cooperate.Operation) error {

	// verify that operation will apply cleanly to document
	// TODO(tylerchr): Assert that the post length also matches.
	if pre, _ := Lengths(op); len(td.contents) != pre {
		return cooperate.ErrDocumentSizeMismatch
	}

	iter := cooperate.NewOperationIterator(op)

	var cursor int

	for iter.More() {

		nextType, next := iter.PeekType(), iter.Peek()

		switch nextType {
		case Retain:
			ret := int(next.(RetainAction))
			cursor += ret
			iter.Consume()

		case Insert:
			text := string(next.(InsertAction))
			td.contents = td.contents[:cursor] + text + td.contents[cursor:]
			cursor += len(text)
			iter.Consume()

		case Delete:
			expectedText := string(next.(DeleteAction))
			actualText := td.contents[cursor : cursor+len(expectedText)]
			if expectedText != actualText {
				panic(fmt.Errorf("failed delete assertion: %s != %s", expectedText, actualText))
			}
			td.contents = td.contents[:cursor] + td.contents[cursor+len(expectedText):]
			iter.Consume()

		default:
			fmt.Printf("unknown action type: %T", nextType)
			return cooperate.ErrUnknownAction
		}

	}

	return nil

}
