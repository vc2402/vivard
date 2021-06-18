package vivard

import (
	"github.com/sirupsen/logrus"
	dep "github.com/vc2402/vivard/dependencies"
)

type LogrusService struct {
	log *logrus.Entry
}

func NewLogrusService(log *logrus.Entry) *LogrusService {
	return &LogrusService{log: log}
}

func (ls *LogrusService) Prepare(eng *Engine, _ dep.Provider) (err error) {
	if ls.log == nil {
		var logger *logrus.Logger
		logger, err = ls.initLogrus(eng)
		if err == nil {
			ls.log = logrus.NewEntry(logger)
		}
	}
	return
}

func (ls *LogrusService) Start(eng *Engine, _ dep.Provider) error {
	return nil
}

func (ls *LogrusService) Log() *logrus.Entry {
	return ls.log
}

func (ls *LogrusService) Named(name string) *logrus.Entry {
	return ls.log.WithField("mod", name)
}

func (ls *LogrusService) initLogrus(eng *Engine) (logger *logrus.Logger, err error) {
	if logrusCfg := eng.ConfValue("logrusConfig"); logrusCfg != nil {
		if lc, ok := logrusCfg.(map[string]interface{}); ok {
			logger, err = initLogrus(lc)
		}
	}
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return
}
