package mongo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vc2402/vivard"
	dep "github.com/vc2402/vivard/dependencies"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ServiceMongo = "mongo"
)

type Service struct {
	db          *mongo.Database
	connections map[string]*mongo.Client
	aliases     map[string]*connection
	guard       sync.RWMutex
	log         *logrus.Entry
	dp          dep.Provider
}

type connection struct {
	url string
	db  *mongo.Database
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
	connectString, _ := ms.dp.Config().GetConfig("Mongo.Aliases." + alias + ".ConnectString").(string)
	if connectString == "" && alias != "default" {
		ms.log.Warnf("getMongo: alias not found: %s", alias)
		return nil, errors.New("invalid alias")
	}
	if connectString == "" {
		connectString = "mongodb://localhost:27017"
	}
	client, exists := ms.connections[connectString]
	if !exists {
		var err error
		client, err = ms.createClient(ctx, connectString)
		if err != nil {
			return nil, err
		}
	}
	dbName, _ := ms.dp.Config().GetConfig("Mongo.Aliases." + alias + ".DBName").(string)
	if dbName == "" {
		dbnames, err := client.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			ms.log.Warnf("getMongo: ListDatabaseNames(%s): %s", connectString, err.Error())
			return nil, err
		} else if len(dbnames) == 0 {
			ms.log.Warnf("getMongo: there is no databases %s", connectString)
			return nil, errors.New("no database")
		}
		dbName = dbnames[0]
		ms.log.Warnf("getMongo: no db name set for alias %s. Using the first one: %s", alias, dbName)
	}
	db := client.Database(dbName)
	ms.aliases[alias] = &connection{connectString, db}
	return db, nil
}

func (ms *Service) createClient(ctx context.Context, connectString string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectString))
	if err != nil {
		ms.log.Warnf("getMongo: NewClient(%s): %s", connectString, err.Error())
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		ms.log.Warnf("getMongo: client.Connect(%s): %s", connectString, err.Error())
		return nil, err
	}
	ms.connections[connectString] = client
	return client, nil
}
