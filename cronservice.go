package vivard

import (
	"github.com/robfig/cron/v3"
	dep "github.com/vc2402/vivard/dependencies"
)

//CRONService provides robfig/cron functionality as a vivard service
type CRONService struct {
	cron *cron.Cron
}

func (cs *CRONService) Prepare(eng *Engine, _ dep.Provider) (err error) {
	if cs.cron == nil {
		cs.cron = cron.New()
	}
	return
}

func (cs *CRONService) Start(eng *Engine, _ dep.Provider) error {
	cs.cron.Start()
	return nil
}

func (cs *CRONService) Provide() interface{} {
	return cs.Cron()
}

func (cs *CRONService) Cron() *cron.Cron {
	return cs.cron
}
