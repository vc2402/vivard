package vivard

import (
	dep "github.com/vc2402/vivard/dependencies"
	"go.uber.org/zap"
)

// well known services
const (
	ServiceGQL = "gql"

	ServiceSequenceProvider = "sequence"
	ServiceScripting        = "scripting"
	ServiceCRON             = "cron"
	ServiceLoggingLogrus    = "logging:logrus"
	ServiceLoggingZap       = "logging:zap"
	ServiceNATS             = "nats"
	ServiceSQL              = "sql"
	ServiceSQLX             = "sqlx"
)

const configZapLogger = "zap-logger"

type Service interface {
	// Prepare will be called for each registered service before SubEngine's Prepare
	Prepare(eng *Engine, provider dep.Provider) error
	Start(eng *Engine, provider dep.Provider) error
	// Provide should return some low level object, e.g. *sql.DB for sql Service
	Provide() interface{}
}

type SubEngine interface {
	Name() string
	Prepare(engine *Engine) error
	Start() error
}

// ServiceProvider may be implemented by SubEngine to register services
type ServiceProvider interface {
	ProvideServices(engine *Engine)
}

type Engine struct {
	gql *GQLEngine
	//TODO: change it to slice for order guaranty
	services map[string]Service
	engines  map[string]SubEngine
	config   *configProvider
	logger   *zap.Logger
}

type Generator func(eng *Engine) error

// NewEngine creates new empty Engine object
func NewEngine() *Engine {
	eng := &Engine{
		services: map[string]Service{},
		engines:  map[string]SubEngine{},
	}

	if loggerCfg := eng.ConfValue(configZapLogger); loggerCfg != nil {
		if lc, ok := loggerCfg.(map[string]interface{}); ok {
			eng.logger, _ = InitZapLogger(lc)
		}
	}
	if eng.logger == nil {
		eng.logger = zap.L()
	}
	return eng
}

// WithLogger sets l as Engine's logger
func (eng *Engine) WithLogger(l *zap.Logger) *Engine {
	eng.logger = l
	return eng
}

// WithService add service to services list
func (eng *Engine) WithService(name string, srv Service) *Engine {
	eng.services[name] = srv
	return eng
}

// WithEngine add SubEngine to list
func (eng *Engine) WithEngine(se SubEngine) *Engine {
	eng.engines[se.Name()] = se
	return eng
}

//Start performs procedure of starting the engine
func (eng *Engine) Start() error {
	for _, s := range eng.engines {
		if sp, ok := s.(ServiceProvider); ok {
			sp.ProvideServices(eng)
		}
	}
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

func (eng *Engine) Logger(name string) *zap.Logger {
	if name == "" {
		return eng.logger
	}
	return eng.logger.With(zap.String("m", name))
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

func InitZapLogger(cfg map[string]interface{}) (*zap.Logger, error) {
	// so far nothing
	return zap.L(), nil
}
