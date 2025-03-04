package vivard

import (
	"context"
	"errors"
	"fmt"
	"github.com/vc2402/go-natshelper"
	dep "github.com/vc2402/vivard/dependencies"
	"strconv"
	"strings"
	"time"
)

// provides sequences via NATS service as Vivard service
// requests are: sequence.<name>.<command>
// commands: next and current

const (
	processorName          = "SequenceProviderProcessor"
	sequenceTopic          = "sequence.>"
	sequenceTopicPrefix    = "sequence"
	sequenceCommandNext    = "next"
	sequenceCommandCurrent = "current"
)

const (
	sequenceDefaultTimeout = time.Second * 10
)

type NatsSequenceProviderMode int

const (
	NsmServer NatsSequenceProviderMode = iota
	NsmClient
)

type NatsSequence struct {
	name     string
	provider *NatsSequenceProvider
}

type NatsSequenceProvider struct {
	mode      NatsSequenceProviderMode
	sequences map[string]Sequence
	provider  SequenceProvider
	vivard    *Engine
	nats      *natshelper.Server
	timeout   time.Duration
}

type NatsSequenceOptionTimeout time.Duration

func NewNatsSequenceProvider(args ...any) *NatsSequenceProvider {
	ns := &NatsSequenceProvider{mode: NsmServer, timeout: sequenceDefaultTimeout}
	for _, arg := range args {
		switch arg := arg.(type) {
		case *natshelper.Server:
			ns.nats = arg
		case SequenceProvider:
			ns.provider = arg
		case NatsSequenceProviderMode:
			ns.mode = arg
		case NatsSequenceOptionTimeout:
			ns.timeout = time.Duration(arg)
		case time.Duration:
			ns.timeout = arg
		default:
			fmt.Printf("undefined parameter for NatsSequenceProvider: %v (%T)\n", arg, arg)
		}
	}
	return ns
}

// RegisterSequence register sequence
//
//	for server mode it is exported sequence
//	for client mode (sequence may be nil) it is sequence that should be got from server
func (ns *NatsSequenceProvider) RegisterSequence(name string, sequence Sequence) error {
	if ns.sequences == nil {
		ns.sequences = map[string]Sequence{}
	}
	ns.sequences[name] = sequence
	return nil
}

// RegisterProvider registers default provider
//
//	for server mode it is provider for sequences to export (if no sequences were registered with RegisterSequence)
//	for client mode it is provider for sequences that were not registered with RegisterSequence
func (ns *NatsSequenceProvider) RegisterProvider(provider SequenceProvider) error {
	ns.provider = provider
	return nil
}

func (ns *NatsSequenceProvider) Prepare(eng *Engine, prov dep.Provider) (err error) {
	ns.vivard = eng
	return
}

func (ns *NatsSequenceProvider) Start(eng *Engine, prov dep.Provider) error {
	if ns.nats == nil {
		nhw := eng.GetService("nats")
		if nhs, ok := nhw.Provide().(*natshelper.Server); ok {
			ns.nats = nhs
		}
	}
	if ns.nats == nil {
		return errors.New("nats service not found")
	}
	err := ns.nats.AddRequestProcessor(processorName, sequenceTopic, ns, false)
	if err != nil {
		return err
	}
	return nil
}

func (ns *NatsSequenceProvider) Provide() interface{} {
	return ns
}

func (ns *NatsSequenceProvider) ProcessNatsRequest(topic string, request []byte) (response []byte, err error) {
	parts := strings.Split(topic, ".")
	if parts[0] == sequenceTopicPrefix && len(parts) > 2 {
		ctx := context.Background()
		var sequence Sequence
		if ns.provider != nil {
			sequence, err = ns.provider.Sequence(ctx, parts[1])
			if err != nil {
				return nil, err
			}
		} else if s, ok := ns.sequences[parts[1]]; ok {
			sequence = s
		}
		if sequence == nil {
			return nil, errors.New("sequence not found")
		}
		var result int
		switch parts[2] {
		case sequenceCommandCurrent:
			result, err = sequence.Current(ctx)
		case sequenceCommandNext:
			result, err = sequence.Next(ctx)
		}
		return []byte(strconv.Itoa(result)), nil
	}
	return nil, nil
}

func (ns *NatsSequenceProvider) Sequence(ctx context.Context, name string) (Sequence, error) {
	if _, ok := ns.sequences[name]; !ok {
		return ns.provider.Sequence(ctx, name)
	}
	return NatsSequence{
		name:     name,
		provider: ns,
	}, nil
}

func (ns *NatsSequenceProvider) ListSequences(ctx context.Context, mask string) (map[string]int, error) {
	// so far only local sequences
	return ns.provider.ListSequences(ctx, mask)
}

func (ns *NatsSequenceProvider) nextForSequence(ctx context.Context, name string) (int, error) {
	return ns.commandForSequence(ctx, name, sequenceCommandNext)
}

func (ns *NatsSequenceProvider) currentForSequence(ctx context.Context, name string) (int, error) {
	return ns.commandForSequence(ctx, name, sequenceCommandCurrent)
}

func (ns *NatsSequenceProvider) commandForSequence(ctx context.Context, name string, command string) (int, error) {
	if ns.nats == nil {
		return 0, errors.New("nats service not found")
	}
	topic := sequenceTopicPrefix + "." + name + "." + command
	response, err := ns.nats.RequestSync(topic, nil, ns.timeout)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(response))
}

// NatsSequence implementation of Sequence

func (nseq NatsSequence) Next(ctx context.Context) (int, error) {
	return nseq.provider.nextForSequence(ctx, nseq.name)
}

func (nseq NatsSequence) Current(ctx context.Context) (int, error) {
	return nseq.provider.currentForSequence(ctx, nseq.name)
}

func (nseq NatsSequence) SetCurrent(ctx context.Context, value int) (int, error) {
	return 0, errors.New("SetCurrent is not implemented for remote sequence")
}
