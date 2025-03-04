package vivard

import (
	"context"
	dep "github.com/vc2402/vivard/dependencies"
)

// Sequence - interface for named sequence of integer
type Sequence interface {
	// Next increments current value and returns it as next value for sequence
	Next(ctx context.Context) (int, error)
	// Current returns current (last returned by Next) value
	Current(ctx context.Context) (int, error)
	// SetCurrent sets current value of Sequence to value
	SetCurrent(ctx context.Context, value int) (int, error)
}

// SequenceProvider provides sequences
type SequenceProvider interface {
	// Sequence returns Sequence object for given name
	Sequence(ctx context.Context, name string) (Sequence, error)
	// ListSequences returns sequences with names containing mask (case-insensitivity may depend on implementation)
	// return map with Sequence name as key and current value as value
	ListSequences(ctx context.Context, mask string) (map[string]int, error)
}

// SequenceService provides sequence provider
type SequenceService struct {
	provider SequenceProvider
}

func NewSequenceService(provider SequenceProvider) *SequenceService {
	return &SequenceService{provider: provider}
}

func (ss *SequenceService) Prepare(eng *Engine, prov dep.Provider) (err error) {
	if ip, ok := ss.provider.(InitProcessor); ok {
		return ip.Prepare(eng, prov)
	}
	return
}

func (ss *SequenceService) Start(eng *Engine, prov dep.Provider) error {
	if ip, ok := ss.provider.(InitProcessor); ok {
		return ip.Start(eng, prov)
	}
	return nil
}

func (ss *SequenceService) Provide() interface{} {
	return ss.provider
}
