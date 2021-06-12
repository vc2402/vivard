package postgre

import (
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"
)

type Service struct {
	db    *sqlx.DB
	guard sync.RWMutex
	log   *logrus.Entry
	dp    dep.Provider
}

func (ss *Service) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	ss.dp = prov
	if ss.db == nil {
		ss.tryConnectPostgre()
	}
	return
}

func (ss *Service) Start(eng *vivard.Engine, prov dep.Provider) error {
	return nil
}

func (ss *Service) DB() *sqlx.DB {
	return ss.db
}

func (ss *Service) getPostgreDB() (*sqlx.DB, error) {
	if ss.db == nil {
		ss.tryConnectPostgre()
	}
	ss.guard.RLock()
	defer ss.guard.RUnlock()
	if ss.db != nil {
		return ss.db, nil
	}
	return nil, errors.New("can't connect to Postgre")
}

func (ss *Service) tryConnectPostgre() {
	if ss.db == nil {
		ss.guard.Lock()
		defer ss.guard.Unlock()
		connectString := "user=guest password=postgre host=192.168.150.15 port=5432 dbname=om "

		if cs, ok := ss.dp.Config().GetConfig("Postgre.connectString").(string); ok && cs != "" {
			connectString = cs
		}
		ss.log.Tracef("pgsql: trying to connect with connect string '%s", connectString)
		var err error
		ss.db, err = sqlx.Open("pgx", connectString)
		if err != nil {
			ss.log.Warnf("pgsql connection: %v", err)
		}
	}
}
