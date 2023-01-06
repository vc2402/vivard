package vivard

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	dep "github.com/vc2402/vivard/dependencies"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
)

type JobID int

type Job struct {
	id              JobID
	name            string
	command         func(ctx context.Context) error
	spec            string
	entry           cron.EntryID
	cancelFn        context.CancelFunc
	runCount        int
	duration        time.Duration
	lastRunAt       time.Time
	lastRunDuration time.Duration
	lastError       error
	lastErrorTime   time.Time
}

//CRONService provides robfig/cron functionality as a vivard service
type CRONService struct {
	cron *cron.Cron
	jobs map[JobID]*Job
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
	return cs
}

func (cs *CRONService) Cron() *cron.Cron {
	return cs.cron
}

func (cs *CRONService) AddFunc(spec string, cmd func(ctx context.Context) error) (JobID, error) {
	name := runtime.FuncForPC(reflect.ValueOf(cmd).Pointer()).Name()
	return cs.AddNamedFunc(spec, name, cmd)
}

func (cs *CRONService) AddNamedFunc(spec string, name string, cmd func(ctx context.Context) error) (JobID, error) {
	if cs.jobs == nil {
		cs.jobs = map[JobID]*Job{}
	}
	job := &Job{
		id:      JobID(len(cs.jobs)),
		name:    name,
		command: cmd,
		spec:    spec,
	}
	cs.jobs[job.id] = job
	var err error
	job.entry, err = cs.cron.AddJob(spec, job)
	if err != nil {
		job.lastError = err
	}
	return job.id, err
}

func (j *Job) Run() {
	defer j.recover()
	var ctx context.Context
	ctx, j.cancelFn = context.WithCancel(context.Background())
	j.lastRunAt = time.Now()
	err := j.command(ctx)
	if err != nil {
		j.lastError = err
		j.lastErrorTime = time.Now()
	}
	j.lastRunDuration = time.Since(j.lastRunAt)
	j.runCount++
	j.duration += j.lastRunDuration
	j.cancelFn = nil
}

func (j *Job) Cancel() {
	if j.cancelFn != nil {
		j.cancelFn()
	}
}

func (j *Job) recover() {
	if r := recover(); r != nil {
		j.lastError = fmt.Errorf("recovered: %v\n  stack trace: %s", r, string(debug.Stack()))
		j.lastErrorTime = time.Now()
	}
}
