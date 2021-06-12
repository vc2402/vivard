package vivard

import "github.com/robfig/cron/v3"

//CRONService provides robfig/cron functionality as a vivard service
type CRONService struct {
	cron *cron.Cron
}

func (cs *CRONService) Prepare(eng *Engine) (err error) {
	if cs.cron == nil {
		cs.cron = cron.New()
	}
	return
}

func (cs *CRONService) Start(eng *Engine) error {
	cs.cron.Start()
	return nil
}

func (cs *CRONService) Cron() *cron.Cron {
	return cs.cron
}
