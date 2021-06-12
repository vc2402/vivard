package vivard

import "errors"

var (
	//ErrNoSequenceProvider - there is not SequenceProvider registered
	ErrNoSequenceProvider = errors.New("no SequenceProvider registered")
)
