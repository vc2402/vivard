package scripting

import (
	"go.uber.org/zap"
	"sync"

	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"
)

type Service struct {
	prefix  string
	suffix  string
	scripts map[string]*script
	context map[string]interface{}
	modules map[string]interface{}
	locker  sync.Mutex
	log     *zap.Logger
}

func (s *Service) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	s.prefix = "./scripts/"
	s.suffix = ".js"
	s.scripts = make(map[string]*script)
	s.modules = make(map[string]interface{})
	s.context = nil
	s.log = prov.Logger("scripting")

	config, _ := prov.Config().GetConfig("scripting").(map[string]interface{})
	if config != nil {
		if pr, ok := config["filePrefix"]; ok {
			s.prefix = pr.(string)
		}
		if sf, ok := config["fileSuffix"]; ok {
			s.prefix = sf.(string)
		}
	}
	s.modules["@logger"] = map[string]interface{}{"log": s.log.Sugar()}
	return
}

func (s *Service) Start(eng *vivard.Engine, prov dep.Provider) error {
	return nil
}

func (s *Service) Provide() interface{} {
	return s
}
