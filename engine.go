package vivard

import (
	"github.com/sirupsen/logrus"
	dep "github.com/vc2402/vivard/dependencies"
)

// well known services
const (
	ServiceGQL = "gql"

	ServiceSequenceProvider = "sequence"
	ServiceScripting        = "scripting"
	ServiceCRON             = "cron"
	ServiceLoggingLogrus    = "logging:logrus"
	ServiceNATS             = "nats"
	ServiceSQLX             = "sqlx"
)

type Service interface {
	Prepare(eng *Engine, provider dep.Provider) error
	Start(eng *Engine, provider dep.Provider) error
}

type SubEngine interface {
	Name() string
	Prepare(engine *Engine) error
	Start() error
}

type Engine struct {
	gql *GQLEngine
	//TODO: change it to slice for order guaranty
	services map[string]Service
	engines  map[string]SubEngine
	config   *configProvider
	logger   *logrus.Logger
}

type Generator func(eng *Engine) error

//NewEngine creates new empty Engine object
func NewEngine() *Engine {
	eng := &Engine{
		services: map[string]Service{},
		engines:  map[string]SubEngine{},
	}

	if logrusCfg := eng.ConfValue("logrusConfig"); logrusCfg != nil {
		if lc, ok := logrusCfg.(map[string]interface{}); ok {
			eng.logger, _ = initLogrus(lc)
		}
	}
	if eng.logger == nil {
		eng.logger = logrus.StandardLogger()
	}
	return eng
}

//WithService add service to services list
func (eng *Engine) WithService(name string, srv Service) *Engine {
	eng.services[name] = srv
	return eng
}

//WithEngine add subengine to list
func (eng *Engine) WithEngine(se SubEngine) *Engine {
	eng.engines[se.Name()] = se
	return eng
}

//Start performs procedure of starting the engine
func (eng *Engine) Start() error {
	for _, s := range eng.services {
		err := s.Prepare(eng, eng)
		if err != nil {
			return err
		}
	}
	for _, s := range eng.engines {
		err := s.Prepare(eng)
		if err != nil {
			return err
		}
	}
	for _, s := range eng.services {
		err := s.Start(eng, eng)
		if err != nil {
			return err
		}
	}
	for _, s := range eng.engines {
		err := s.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

//GetService looks for registered service and returns it; returns nil if not found
func (eng *Engine) GetService(tip string) Service {
	return eng.services[tip]
}

//GetConfig looks for config with given name and returns it
func (eng *Engine) GetConfig(name string) interface{} {
	return eng.ConfValue(name)
}

// Engine looks for subengine with given name
func (eng *Engine) Engine(name string) SubEngine {
	return eng.engines[name]
}

func (eng *Engine) Config() dep.ConfigProvider {
	return eng
}

func (eng *Engine) Logger(name string) *logrus.Entry {
	return eng.logger.WithField("m", name)
}

// func (eng *Engine) RegisterSequenceProvider(sp SequenceProvider) {
// 	eng.sequenceProvider = sp
// }

// func (eng *Engine) Sequence(ctx context.Context, name string) (Sequence, error) {
// 	if eng.sequenceProvider == nil {
// 		return nil, ErrNoSequenceProvider
// 	}
// 	return eng.sequenceProvider.Sequence(ctx, name)
// }
