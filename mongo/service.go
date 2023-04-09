package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"

	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ServiceMongo = "mongo"
)

type ConnectionConfig struct {
	Alias         string
	ConnectString string
	DBName        string
}

type Service struct {
	db          *mongo.Database
	config      map[string]ConnectionConfig
	connections map[string]*mongo.Client
	aliases     map[string]*connection
	guard       sync.RWMutex
	log         *zap.Logger
	dp          dep.Provider
}

type connection struct {
	url string
	db  *mongo.Database
}

// New creates new mongo service
// params may be:
//
//	*mongo.Database (should be first)
//	db-name (connect string will be build for the local host)
//	pair connect string - db-name
//	ConnectionConfig object
func New(params ...any) (*Service, error) {
	s := &Service{}
	cs := map[string]ConnectionConfig{}
	for i := 0; i < len(params); i++ {
		switch v := params[i].(type) {
		case *mongo.Database:
			s.db = v
		case string:
			conf := ConnectionConfig{Alias: "default"}
			if i <= len(params)-2 {
				if dbn, ok := params[i+1].(string); ok {
					i++
					conf.DBName = dbn
					conf.ConnectString = v
				}
			} else {
				conf.DBName = v
			}
			cs[conf.Alias] = conf
		case ConnectionConfig:
			cs[v.Alias] = v
		default:
			return nil, fmt.Errorf("invalid param: %v (%T)", params[i], params[i])
		}
	}
	if len(cs) > 0 {
		s.config = cs
	}
	return s, nil
}

func (ms *Service) With(db *mongo.Database) *Service {
	ms.db = db
	return ms
}

func (ms *Service) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	ms.aliases = make(map[string]*connection)
	ms.connections = make(map[string]*mongo.Client)
	ms.log = prov.Logger("mongo")
	ms.dp = prov
	if ms.db == nil {
		ms.db, err = ms.GetDefaultMongo(context.Background())
	}
	return
}

func (ms *Service) Start(eng *vivard.Engine, prov dep.Provider) error {
	return nil
}

func (ms *Service) Provide() interface{} {
	return ms.DB()
}

func (ms *Service) DB() *mongo.Database {
	if ms.db == nil {
		panic("mongo service not initialized")
	}
	return ms.db
}

// GetMongo returns database with given alias
func (ms *Service) GetMongo(ctx context.Context, alias string) (*mongo.Database, error) {
	return ms.getMongoDB(ctx, alias)
}

// GetDefaultMongo returns default database
func (ms *Service) GetDefaultMongo(ctx context.Context) (*mongo.Database, error) {
	return ms.getMongoDB(ctx, "default")
}

func (ms *Service) getMongoDB(ctx context.Context, alias string) (*mongo.Database, error) {
	ms.guard.RLock()
	db, exists := ms.aliases[alias]
	ms.guard.RUnlock()
	if !exists {
		return ms.registerNewDB(ctx, alias)
	}
	return db.db, nil
}

func (ms *Service) registerNewDB(ctx context.Context, alias string) (*mongo.Database, error) {
	ms.guard.Lock()
	defer ms.guard.Unlock()
	var conf ConnectionConfig
	confFound := false
	if cf, ok := ms.dp.Config().GetConfig("mongo.aliases." + alias).(map[string]interface{}); ok {
		conf.Alias = alias
		if cs, ok := cf["connectstring"].(string); ok {
			conf.ConnectString = cs
			confFound = true
		} else if cs, ok := cf["connect-string"].(string); ok {
			conf.ConnectString = cs
			confFound = true
		}
		if db, ok := cf["dbname"].(string); ok {
			conf.DBName = db
		} else if db, ok := cf["db-name"].(string); ok {
			conf.ConnectString = db
		}
	}
	if !confFound && ms.config != nil {
		if cf, ok := ms.config[alias]; ok {
			conf = cf
			confFound = true
		}
	}
	if !confFound && alias != "default" {
		ms.log.Error("Mongo: alias not found", zap.String("alias", alias))
		return nil, fmt.Errorf("no configuration found for alias: %s", alias)
	}
	if !confFound || conf.ConnectString == "" {
		conf.ConnectString = "mongodb://localhost:27017"
	}
	client, exists := ms.connections[conf.ConnectString]
	if !exists {
		var err error
		client, err = ms.createClient(ctx, conf.ConnectString)
		if err != nil {
			return nil, err
		}
		ms.log.Info("Mongo: got new connection for alias", zap.String("alias", alias))
	}

	if conf.DBName == "" {
		dbnames, err := client.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			ms.log.Error("Mongo: ListDatabaseNames", zap.String("cs", conf.ConnectString), zap.Error(err))
			return nil, err
		} else if len(dbnames) == 0 {
			ms.log.Error("Mongo: there is no databases for %s", zap.String("cs", conf.ConnectString))
			return nil, errors.New("no database")
		}
		conf.DBName = dbnames[0]
		ms.log.Info("Mongo: no db name set for alias. Using the first one", zap.String("alias", alias), zap.String("result", conf.DBName))
	}
	db := client.Database(conf.DBName)
	ms.aliases[alias] = &connection{conf.ConnectString, db}
	return db, nil
}

func (ms *Service) createClient(ctx context.Context, connectString string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectString))
	if err != nil {
		ms.log.Error("Mongo: NewClient", zap.String("cs", connectString), zap.Error(err))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		ms.log.Error("Mongo: client.Connect", zap.String("cs", connectString), zap.Error(err))
		return nil, err
	}
	ms.connections[connectString] = client
	return client, nil
}
