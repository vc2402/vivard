package nats

import (
	"time"

	ng "github.com/nats-io/nats.go"
	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"
)

type Service struct {
	conn *ng.Conn
	dp   dep.Provider
	err  error
}

func ForConnection(conn *ng.Conn) *Service {
	return &Service{conn: conn}
}

func (ns *Service) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	ns.dp = prov
	return
}

func (ns *Service) Start(eng *vivard.Engine, prov dep.Provider) error {
	ns.Connect()
	return nil
}

func (ns *Service) Conn() *ng.Conn {
	return ns.conn
}

func (ns *Service) Err() error {
	return ns.err
}

func (ns *Service) Connect() error {
	if ns.conn == nil {
		o := &ng.Options{
			Url:            "nats://localhost:4222",
			AllowReconnect: true,
			ReconnectWait:  time.Second * 30,
		}
		skipConnectError := false
		if cfg, ok := dp.Config().GetConfig("nats").(map[string]interface{}); ok {
			if url, ok := cfg["url"].(string); ok {
				o.Url = url
			}
			if ar, ok := cfg["allowReconnect"].(bool); ok {
				o.AllowReconnect = ar
			}
			if sce, ok := cfg["skipConnectError"].(bool); ok {
				skipConnectError = sce
			}
		}
	}
	ns.conn, ns.err = o.Connect()
	if !skipConnectError {
		return ns.err
	} else {
		return nil
	}
}

func (ns *Service) Reconnect() error {
	ns.conn = nil
	return ns.Connect()
}
