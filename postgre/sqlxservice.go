package postgre

import (
	"errors"
	"go.uber.org/zap"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"
)

type Service struct {
	db    *sqlx.DB
	guard sync.RWMutex
	log   *zap.Logger
	dp    dep.Provider
}

func (ss *Service) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	ss.log = prov.Logger("postgres")
	ss.dp = prov
	if ss.db == nil {
		ss.tryConnectPostgres()
	}
	return
}

func (ss *Service) Start(eng *vivard.Engine, prov dep.Provider) error {
	return nil
}

func (ss *Service) Provide() interface{} {
	return ss.DB()
}

func (ss *Service) DB() *sqlx.DB {
	return ss.db
}

func (ss *Service) getPostgresDB() (*sqlx.DB, error) {
	if ss.db == nil {
		ss.tryConnectPostgres()
	}
	ss.guard.RLock()
	defer ss.guard.RUnlock()
	if ss.db != nil {
		return ss.db, nil
	}
	return nil, errors.New("can't connect to Postgre")
}

func (ss *Service) tryConnectPostgres() {
	if ss.db == nil {
		ss.guard.Lock()
		defer ss.guard.Unlock()
		connectString := "user=guest password=postgre host=127.0.0.1 port=5432 dbname=postgre "

		if cs, ok := ss.dp.Config().GetConfig("Postgres.connectString").(string); ok && cs != "" {
			connectString = cs
		}
		ss.log.Debug("pgsql: trying to connect", zap.String("connectString", connectString))
		var err error
		ss.db, err = sqlx.Open("pgx", connectString)
		if err != nil {
			ss.log.Error("pgsql connection", zap.String("connectString", connectString), zap.Error(err))
		}
	}
}
