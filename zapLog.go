package vivard

import (
	dep "github.com/vc2402/vivard/dependencies"
	"go.uber.org/zap"
)

type LoggerService struct {
	log *zap.Logger
}

func NewLoggerService(log *zap.Logger) *LoggerService {
	return &LoggerService{log: log}
}

func (ls *LoggerService) Prepare(eng *Engine, _ dep.Provider) (err error) {
	if ls.log == nil {
		ls.log = eng.Logger("")
	}
	return
}

func (ls *LoggerService) Start(_ *Engine, _ dep.Provider) error {
	return nil
}

func (ls *LoggerService) Provide() any {
	return ls.Log()
}

func (ls *LoggerService) Log() *zap.Logger {
	return ls.log
}

func (ls *LoggerService) Named(name string) *zap.Logger {
	return ls.log.With(zap.String("mod", name))
}
