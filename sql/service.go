package postgre

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"
	"go.uber.org/zap"
)

var ErrUndefinedParam = errors.New("invalid param")
var ErrNoProvider = errors.New("ConnectionProvider was not provided")

type ConnectionProvider = func(conf any) (*sql.DB, error)

type Service struct {
	db   *sql.DB
	log  *zap.Logger
	dp   dep.Provider
	cp   ConnectionProvider
	conf any
}

// New creates new sql service
// params may be *sql.DB or ConnectionProvider;
//  in later case the next param (if any) will be used as conf argument in call to provider
func New(params ...any) (*Service, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("%w: no params given", ErrUndefinedParam)
	}
	s := &Service{}
	for i := 0; i < len(params); i++ {
		p := params[i]
		switch v := p.(type) {
		case *sql.DB:
			s.db = v
		case ConnectionProvider:
			s.cp = v
			if i < len(params)-1 {
				i++
				s.conf = params[i]
			}
		default:
			return nil, fmt.Errorf("%w: %T", ErrUndefinedParam, p)
		}
	}
	return s, nil
}

func (ss *Service) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	ss.log = prov.Logger("sql")
	ss.dp = prov
	if ss.db == nil {
		ss.tryConnect(false)
	}
	return
}

func (ss *Service) Start(eng *vivard.Engine, prov dep.Provider) error {
	return nil
}

func (ss *Service) Provide() any {
	return ss
}

func (ss *Service) DB() (*sql.DB, error) {
	if ss.db != nil {
		return ss.db, nil
	}
	err := ss.tryConnect(false)
	return ss.db, err
}

func (ss *Service) Reconnect(force bool) error {
	return ss.tryConnect(force)
}

func (ss *Service) tryConnect(force bool) error {
	if ss.db == nil || force {
		if ss.cp == nil {
			return ErrNoProvider
		}
		if ss.db != nil {
			ss.db.Close()
		}
		var err error
		ss.db, err = ss.cp(ss.conf)
		if err != nil {
			ss.log.Error("sql connection provider", zap.Error(err))
		}
		return err
	}
	return nil
}
