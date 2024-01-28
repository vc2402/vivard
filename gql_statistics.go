package vivard

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	"go.uber.org/zap"
	"hash/fnv"
	"regexp"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

const (
	optionLogClientErrors         = "LogClientErrors"
	optionStatisticsSnapshotStep  = "StatisticsSnapshotStep"
	optionStatisticsSnapshotCount = "StatisticsSnapshotsCount"
	optionCollectStatistics       = "CollectStatistics"
)

var durationType = graphql.Float
var minMaxType = graphql.Int

type queryStatistics struct {
	operation    string
	query        string
	started      time.Time
	finished     time.Time
	duration     time.Duration
	isSuccessful bool
	errors       []string
}

type statistic struct {
	from        time.Time
	to          time.Time
	count       int
	duration    time.Duration
	maxDuration time.Duration
	minDuration time.Duration
	maxAt       time.Time
	minAt       time.Time
	lastErrorAt time.Time
	lastError   []string
	errors      int
}

type statistics struct {
	name     string
	query    string
	overall  statistic
	current  statistic
	history  *list.List
	accesMux sync.RWMutex
}

func (gqe *GQLEngine) statisticsProcessor() {
	ticker := time.NewTicker(time.Minute)
	lastShiftAt := time.Now()
	for {
		select {
		case s, ok := <-gqe.statisticsChannel:
			if !ok {
				ticker.Stop()
				if gqe.log != nil {
					gqe.log.Debug("statisticsProcessor: exiting")
				}
				return
			}
			gqe.doProcessStatistics(s)
		case <-ticker.C:
			now := time.Now()
			if gqe.options.StatisticsSnapshotStep > 0 &&
				!lastShiftAt.Truncate(gqe.options.StatisticsSnapshotStep).Equal(now.Truncate(gqe.options.StatisticsSnapshotStep)) {
				gqe.doShiftStatistics()
				lastShiftAt = now
			}
		}
	}
}

func (gqe *GQLEngine) collectQueryStatistics(qs queryStatistics) {
	if len(gqe.statisticsChannel) < cap(gqe.statisticsChannel) {
		gqe.statisticsChannel <- qs
	} else if gqe.log != nil {
		gqe.log.Warn("collectQueryStatistics: statisticsChannel is overcrowded; skipping statistics")
	}
}

func (gqe *GQLEngine) startQueryStatistics(operation string, query string) queryStatistics {
	return queryStatistics{operation: operation, query: query, started: time.Now()}
}

func (gqe *GQLEngine) doProcessStatistics(s interface{}) {
	defer gqe.recoverer()
	switch s.(type) {
	case queryStatistics:
		gqe.doProcessQueryStatistics(s.(queryStatistics))
	}
}

func (gqe *GQLEngine) doProcessQueryStatistics(qs queryStatistics) {
	r := regexp.MustCompile(`(?s)(query|mutation)[^{]*{[^a-zA-Z0-9_]*([a-zA-Z0-9_]*)`)
	op := r.FindStringSubmatch(qs.query)
	opName := "<undefined>"
	if len(op) > 2 {
		opName = fmt.Sprintf("%s:%s", op[1][:1], op[2])
	}
	hash := gqe.hashForQuery(qs.query)
	gqe.statisticsMux.RLock()
	st, ok := gqe.statistics[hash]
	gqe.statisticsMux.RUnlock()
	if !ok {
		gqe.statisticsMux.Lock()
		st, ok = gqe.statistics[hash]
		if !ok {
			st = &statistics{name: opName, query: qs.query, history: list.New()}
			gqe.statistics[hash] = st
		}
		gqe.statisticsMux.Unlock()
	}
	qs.duration = qs.finished.Sub(qs.started) / time.Microsecond
	st.overall.update(qs)
	st.current.update(qs)
	if len(qs.errors) > 0 && gqe.options.LogClientErrors {
		if gqe.log != nil {
			gqe.log.Error("error sent to client", zap.String("request", opName))
			for _, err := range qs.errors {
				gqe.log.Error("error", zap.String("problem", err), zap.String("request", opName))
			}
		} else {
			for i, err := range qs.errors {
				fmt.Printf("error sent to client for request '%s' %d: %s", opName, i+1, err)
			}
		}
	}

}

func (gqe *GQLEngine) hashForQuery(query string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(query))
	return h.Sum32()
}

func (gqe *GQLEngine) doShiftStatistics() {
	defer gqe.recoverer()
	gqe.statisticsMux.RLock()
	defer gqe.statisticsMux.RUnlock()
	for _, st := range gqe.statistics {
		now := time.Now()
		st.current.to = now
		st.history.PushFront(st.current)
		st.current = statistic{from: now}
		if st.history.Len() > gqe.options.StatisticsSnapshotsCount {
			st.history.Remove(st.history.Back())
		}
	}
}

func (qs *queryStatistics) finish(result *graphql.Result) {
	qs.finished = time.Now()
	qs.isSuccessful = result != nil && !result.HasErrors()
	if len(result.Errors) > 0 {
		qs.errors = make([]string, len(result.Errors))
		for i, e := range result.Errors {
			qs.errors[i] = e.Error()
		}
	}
}

func (st *statistic) update(qs queryStatistics) {
	if st.from.IsZero() {
		st.from = qs.started
	}
	st.to = qs.finished
	st.count++
	st.duration += qs.duration
	if !qs.isSuccessful {
		st.errors++
		st.lastError = qs.errors
		st.lastErrorAt = qs.finished
	}
	if st.maxDuration < qs.duration {
		st.maxDuration = qs.duration
		st.maxAt = qs.finished
	}
	if st.minDuration > qs.duration || st.minDuration == 0 {
		st.minDuration = qs.duration
		st.minAt = qs.finished
	}
}

func (gqe *GQLEngine) recoverer() {
	if r := recover(); r != nil {
		if gqe.log != nil {
			gqe.log.Warn("GQLEngine: recovered", zap.Any("problem", r))
			gqe.log.Warn("\t", zap.String("stack", string(debug.Stack())))
		} else {
			fmt.Printf("GQLEngine: recovered: %v\n", r)
			fmt.Printf("\tstack trace: %s", string(debug.Stack()))
		}
	}
}

func (gqe *GQLEngine) getStatisticsSchema() (graphql.Schema, error) {
	var statisticType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Statistic",
			Fields: graphql.Fields{
				"from": &graphql.Field{
					Type: graphql.NewNonNull(graphql.DateTime),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return s.from, nil
					},
				},
				"to": &graphql.Field{
					Type: graphql.NewNonNull(graphql.DateTime),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return s.to, nil
					},
				},
				"count": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Int),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return s.count, nil
					},
				},
				"errors": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Int),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return s.errors, nil
					},
				},
				"duration": &graphql.Field{
					Type: graphql.NewNonNull(durationType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return int64(s.duration / time.Microsecond), nil
					},
				},
				"minDuration": &graphql.Field{
					Type: graphql.NewNonNull(minMaxType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return int64(s.minDuration / time.Microsecond), nil
					},
				},
				"maxDuration": &graphql.Field{
					Type: graphql.NewNonNull(minMaxType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return int64(s.maxDuration / time.Microsecond), nil
					},
				},
				"maxAt": &graphql.Field{
					Type: graphql.DateTime,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						if s.maxAt.IsZero() {
							return nil, nil
						}
						return s.maxAt, nil
					},
				},
				"minAt": &graphql.Field{
					Type: graphql.DateTime,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						if s.minAt.IsZero() {
							return nil, nil
						}
						return s.minAt, nil
					},
				},
				"lastErrorAt": &graphql.Field{
					Type: graphql.DateTime,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						if s.lastErrorAt.IsZero() {
							return nil, nil
						}
						return s.lastErrorAt, nil
					},
				},
				"lastError": &graphql.Field{
					Type: graphql.NewList(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return s.lastError, nil
					},
				},
			},
		},
	)
	var statisticsType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Statistics",
			Fields: graphql.Fields{
				"name": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(*statistics)
						return s.name, nil
					},
				},
				"query": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(*statistics)
						return s.query, nil
					},
				},
				"overall": &graphql.Field{
					Type: graphql.NewNonNull(statisticType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(*statistics)
						return s.overall, nil
					},
				},
				"current": &graphql.Field{
					Type: graphql.NewNonNull(statisticType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(*statistics)
						return s.current, nil
					},
				},
				"history": &graphql.Field{
					Type: graphql.NewList(statisticType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if q, ok := p.Info.VariableValues["query"]; !ok || q == nil {
							return nil, nil
						}
						s := p.Source.(*statistics)
						var ret []statistic
						if s.history != nil {
							curr := s.history.Front()
							ret = make([]statistic, s.history.Len())
							idx := 0
							for curr != nil {
								ret[idx] = curr.Value.(statistic)
								idx++
								curr = curr.Next()
							}
						}
						return ret, nil
					},
				},
			},
		},
	)
	var optionsType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Options",
			Fields: graphql.Fields{
				optionLogClientErrors: &graphql.Field{
					Type: graphql.NewNonNull(graphql.Boolean),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return gqe.options.LogClientErrors, nil
					},
				},
				optionCollectStatistics: &graphql.Field{
					Type: graphql.NewNonNull(graphql.Boolean),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return gqe.collectStatistics, nil
					},
				},
				optionStatisticsSnapshotStep: &graphql.Field{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Duration to push current state in the states history",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return gqe.options.StatisticsSnapshotStep.String(), nil
					},
				},
				optionStatisticsSnapshotCount: &graphql.Field{
					Type:        graphql.NewNonNull(graphql.Int),
					Description: "number of historic records to store",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return gqe.options.StatisticsSnapshotsCount, nil
					},
				},
			},
		},
	)
	var optionsInputType = graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name: "OptionsInputType",
			Fields: graphql.InputObjectConfigFieldMap{
				optionLogClientErrors: &graphql.InputObjectFieldConfig{
					Type: graphql.Boolean,
				},
				optionCollectStatistics: &graphql.InputObjectFieldConfig{
					Type: graphql.Boolean,
				},
				optionStatisticsSnapshotStep: &graphql.InputObjectFieldConfig{
					Type:        graphql.String,
					Description: "duration in golang duration format or integer in minutes",
				},
				optionStatisticsSnapshotCount: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
			},
		},
	)

	rootQuery := graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"statistics": &graphql.Field{
				Type:        graphql.NewList(statisticsType),
				Description: "List statistics",
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{
						Type:         graphql.String,
						DefaultValue: "",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					query := p.Args["query"].(string)
					gqe.statisticsMux.RLock()
					defer gqe.statisticsMux.RUnlock()
					if query != "" {
						hash := gqe.hashForQuery(query)
						st, ok := gqe.statistics[hash]
						if !ok {
							return nil, errors.New("invalid query")
						}
						return []*statistics{st}, nil
					}
					var res []*statistics
					for _, st := range gqe.statistics {
						res = append(res, st)
					}
					return res, nil
				},
			},
		},
	}
	mutation := graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"options": &graphql.Field{
				Type:        optionsType,
				Description: "set or get options",
				Args: graphql.FieldConfigArgument{
					"options": &graphql.ArgumentConfig{
						Type: optionsInputType,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if opts, ok := p.Args["options"].(map[string]interface{}); ok {
						if log, ok := opts[optionLogClientErrors].(bool); ok {
							gqe.options.LogClientErrors = log
						}
						if step, ok := opts[optionStatisticsSnapshotStep].(string); ok {
							minutes, err := strconv.ParseInt(step, 10, 32)
							if err == nil {
								gqe.options.StatisticsSnapshotStep = time.Duration(minutes) * time.Minute
							} else {
								dur, err := time.ParseDuration(step)
								if err != nil {
									return nil, fmt.Errorf(
										"%s should be integer (value of minutes) or golang duration format string: %s",
										optionStatisticsSnapshotStep,
										step,
									)
								}
								gqe.options.StatisticsSnapshotStep = dur
							}
						}
						if count, ok := opts[optionStatisticsSnapshotCount].(int); ok {
							gqe.options.StatisticsSnapshotsCount = count
						}
						if collect, ok := opts[optionCollectStatistics].(bool); ok {
							if collect != gqe.collectStatistics {
								gqe.CollectStatistics(collect)
							}
						}
					}
					return true, nil
				},
			},
		},
	}
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(mutation),
	}
	return graphql.NewSchema(schemaConfig)
}
