package vivard

import "errors"

var (
	//ErrNoSequenceProvider - there is not SequenceProvider registered
	ErrNoSequenceProvider = errors.New("no SequenceProvider registered")
	// ErrItemNotFound may be returned for dictionary items (wrapped with information about dictionary and id)
	ErrItemNotFound = errors.New("item not found")
)
