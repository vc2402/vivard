package vivard

import "context"

// Sequence - interface for named sequence of integer
type Sequence interface {
	//Next incerements current value and returns it as next value for sequence
	Next(ctx context.Context) (int, error)
	//Current returns current (last returned by Next) value
	Current(ctx context.Context) (int, error)
	//SetCurrent sets current value of Sequence to value
	SetCurrent(ctx context.Context, value int) (int, error)
}

//SequenceProvider provides sequences
type SequenceProvider interface {
	//Sequence returns Sequence object for given name
	Sequence(ctx context.Context, name string) (Sequence, error)
}
