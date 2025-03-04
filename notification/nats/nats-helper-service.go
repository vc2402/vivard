package nats

import (
	"github.com/vc2402/go-natshelper"
	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"
)

type natsHelperWrapper struct {
	nats *natshelper.Server
}

func (n natsHelperWrapper) Prepare(eng *vivard.Engine, provider dep.Provider) error {
	return nil
}

func (n natsHelperWrapper) Start(eng *vivard.Engine, provider dep.Provider) error {
	return nil
}

func (n natsHelperWrapper) Provide() interface{} {
	return n.nats
}
