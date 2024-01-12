package vivard

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"go.uber.org/zap"
	"hash/fnv"
	"regexp"
	"runtime/debug"
	"sync"
	"time"
)

type queryStatistics struct {
	operation    string
	query        string
	started      time.Time
	finished     time.Time
	duration     time.Duration
	isSuccessful bool
	errors       []gqlerrors.FormattedError
}

type statistic struct {
	from        time.Time
	to          time.Time
	count       int
	duration    time.Duration
	maxDuration time.Duration
	minDuration time.Duration
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
				gqe.log.Error("error", zap.String("problem", err.Error()), zap.String("request", opName))
			}
		} else {
			for i, err := range qs.errors {
				fmt.Printf("error sent to client for request '%s' %d: %s", opName, i+1, err.Error())
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
		st.history.PushFront(&list.Element{Value: st.current})
		st.current = statistic{}
		if st.history.Len() > gqe.options.StatisticsSnapshotsCount {
			st.history.Remove(st.history.Back())
		}
	}
}

func (qs *queryStatistics) finish(result *graphql.Result) {
	qs.finished = time.Now()
	qs.isSuccessful = result != nil && !result.HasErrors()
	qs.errors = result.Errors
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
	}
	if st.maxDuration < qs.duration {
		st.maxDuration = qs.duration
	}
	if st.minDuration > qs.duration || st.minDuration == 0 {
		st.minDuration = qs.duration
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
					Type: graphql.NewNonNull(graphql.Int),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return int64(s.duration), nil
					},
				},
				"minDuration": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Int),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return int64(s.minDuration), nil
					},
				},
				"maxDuration": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Int),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						s := p.Source.(statistic)
						return int64(s.maxDuration), nil
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
				"history": &graphql.Field{
					Type: graphql.NewList(statisticType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if _, ok := p.Args["query"]; !ok {
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
							}
						}
						return ret, nil
					},
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
	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
		//Mutation: graphql.NewObject(mutation),
	}
	return graphql.NewSchema(schemaConfig)
}
