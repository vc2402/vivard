package vivard

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	dep "github.com/vc2402/vivard/dependencies"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type JobID int

type Job struct {
	ID              JobID
	Name            string
	command         func(ctx context.Context) (interface{}, error)
	Spec            string
	entry           cron.EntryID
	cancelFn        context.CancelFunc
	RunCount        int
	Duration        time.Duration
	LastRunAt       time.Time
	LastRunDuration time.Duration
	LastError       error
	LastErrorTime   time.Time
	LastResult      interface{}
	cs              *CRONService
}

//CRONService provides robfig/cron functionality as a vivard service
type CRONService struct {
	cron     *cron.Cron
	jobs     map[JobID]*Job
	jobsLock sync.RWMutex
	ctx      context.Context
}

func NewCronService(ctx context.Context) *CRONService {
	return &CRONService{ctx: ctx}
}

func (cs *CRONService) SetJobContext(ctx context.Context) {
	cs.ctx = ctx
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

func (cs *CRONService) AddFunc(spec string, cmd func(ctx context.Context) (interface{}, error)) (JobID, error) {
	name := runtime.FuncForPC(reflect.ValueOf(cmd).Pointer()).Name()
	return cs.AddNamedFunc(spec, name, cmd)
}

func (cs *CRONService) AddNamedFunc(spec string, name string, cmd func(ctx context.Context) (interface{}, error)) (JobID, error) {
	job := &Job{
		ID:      JobID(len(cs.jobs)),
		Name:    name,
		command: cmd,
		Spec:    spec,
		cs:      cs,
	}
	cs.jobsLock.Lock()
	defer cs.jobsLock.Unlock()
	if cs.jobs == nil {
		cs.jobs = map[JobID]*Job{}
	}
	cs.jobs[job.ID] = job
	var err error
	job.entry, err = cs.cron.AddJob(spec, job)
	if err != nil {
		job.LastError = err
	}
	return job.ID, err
}

func (cs *CRONService) ListJobs() []*Job {
	cs.jobsLock.RLock()
	defer cs.jobsLock.RUnlock()
	jobs := make([]*Job, len(cs.jobs))
	i := 0
	for _, job := range cs.jobs {
		jobs[i] = job
		i++
	}
	return jobs
}

func (j *Job) Run() {
	defer j.recover()
	ctx := j.cs.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, j.cancelFn = context.WithCancel(ctx)
	j.LastRunAt = time.Now()
	var err error
	j.LastResult, err = j.command(ctx)
	if err != nil {
		j.LastError = err
		j.LastErrorTime = time.Now()
	}
	j.LastRunDuration = time.Since(j.LastRunAt)
	j.RunCount++
	j.Duration += j.LastRunDuration
	j.cancelFn = nil
}

func (j *Job) Cancel() {
	if j.cancelFn != nil {
		j.cancelFn()
	}
}

func (j *Job) IsRunning() bool {
	return j.cancelFn != nil
}

func (j *Job) ScheduledAt() *time.Time {
	if j.entry != 0 {
		entry := j.cs.cron.Entry(j.entry)
		if entry.Valid() {
			return &entry.Next
		}
	}
	return nil
}

func (j *Job) recover() {
	if r := recover(); r != nil {
		j.LastError = fmt.Errorf("recovered: %v\n  stack trace: %s", r, string(debug.Stack()))
		j.LastErrorTime = time.Now()
	}
}
