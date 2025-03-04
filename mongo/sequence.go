package mongo

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"sync"

	"go.mongodb.org/mongo-driver/mongo/options"

	dep "github.com/vc2402/vivard/dependencies"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/vc2402/vivard"
	"go.mongodb.org/mongo-driver/mongo"
)

const sequencesCollectionName = "_sequences"
const sequencesAreCacheable = true

// Sequence allows to create sequential numbers
//
//	current value is cached!
type Sequence struct {
	p    *SequenceProvider
	name string
	curr int
}

type SequenceProvider struct {
	db        *mongo.Database
	ms        *Service
	log       *zap.Logger
	sequences map[string]*Sequence
	seqMux    sync.RWMutex
}

func SequenceForDB(db *mongo.Database) *SequenceProvider {
	return &SequenceProvider{db: db}
}

func SequenceForService(ms *Service) *SequenceProvider {
	return &SequenceProvider{ms: ms}
}

func (msp *SequenceProvider) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	msp.log = prov.Logger("mongo-seq")
	msp.sequences = map[string]*Sequence{}
	if msp.db == nil {
		if msp.ms == nil {
			mongo, ok := eng.GetService(ServiceMongo).(*Service)
			if !ok {
				return errors.New("MongoService is required for SequenceProvider")
			}
			msp.ms = mongo
		}
		if msp.ms.db == nil {
			err = msp.ms.Prepare(eng, prov)
			if err != nil {
				return
			}
		}

		msp.db = msp.ms.DB()
	}
	return
}

func (msp *SequenceProvider) Start(eng *vivard.Engine, prov dep.Provider) error {
	if msp.db == nil {
		return errors.New("SequenceProvider is not initialized")
	}
	return nil
}

// Sequence returns Sequence object with given name
func (msp *SequenceProvider) Sequence(ctx context.Context, name string) (vivard.Sequence, error) {
	return msp.sequence(ctx, name)
}

// ListSequences returns sequences with names containing mask (case-insensitive)
func (msp *SequenceProvider) ListSequences(ctx context.Context, mask string) (map[string]int, error) {
	query := bson.M{}
	if mask != "" {
		query["_id"] = bson.M{"$regex": mask, "$options": "i"}
	}
	cur, err := msp.db.Collection(sequencesCollectionName).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	ret := map[string]int{}
	for cur.Next(ctx) {
		seq := map[string]interface{}{}
		err = cur.Decode(&seq)
		if err != nil {
			return ret, err
		}
		name, _ := seq["_id"].(string)
		val, _ := seq["current"].(int32)
		ret[name] = int(val)
		if len(ret) > 100 {
			return ret, errors.New("too many records")
		}
	}
	return ret, nil
}

func (msp *SequenceProvider) sequence(ctx context.Context, seqName string) (*Sequence, error) {
	seq := msp.lookForSequence(seqName)
	if seq == nil {
		seq = msp.createSequence(seqName)
	}
	return seq, nil
}

// Next increments current value of Sequence and returns it
// return -1 on error
func (s *Sequence) Next(ctx context.Context) (int, error) {
	var err error
	if s.curr == -1 || !sequencesAreCacheable {
		err = s.load(ctx)
		if err != nil {
			return -1, err
		}
	}
	curr := s.curr
	s.curr++
	err = s.save(ctx)
	if err != nil {
		s.p.log.Error("on update", zap.String("sequence", s.name), zap.Error(err))
		return -1, err
	}
	return curr, nil
}

// Current returns current sequence value
func (s *Sequence) Current(ctx context.Context) (int, error) {
	if s.curr == -1 || !sequencesAreCacheable {
		err := s.load(ctx)
		if err != nil {
			return -1, err
		}
	}
	return s.curr, nil
}

// SetCurrent sets current value of Sequence to value
func (s *Sequence) SetCurrent(ctx context.Context, value int) (int, error) {
	if s.curr == -1 || !sequencesAreCacheable {
		err := s.load(ctx)
		if err != nil {
			return -1, err
		}
	}
	s.curr = value
	err := s.save(ctx)
	if err != nil {
		s.p.log.Error("on update", zap.String("sequence", s.name), zap.Error(err))
		return -1, err
	}
	return s.curr, nil
}

func (s *Sequence) load(ctx context.Context) error {
	m := map[string]interface{}{}
	err := s.p.db.Collection(sequencesCollectionName).
		FindOne(
			ctx,
			bson.M{"_id": s.name},
			options.FindOne().SetProjection(bson.D{{"current", 1}}),
		).
		Decode(&m)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			s.p.log.Error("load: FindOne", zap.String("sequence", s.name), zap.Error(err))
			return err
		}
		s.curr = 1
		s.p.log.Debug("load: initializing new", zap.String("sequence", s.name))
		_, err := s.p.db.Collection(sequencesCollectionName).InsertOne(ctx, bson.M{"_id": s.name, "current": s.curr})
		if err != nil {
			s.p.log.Error("load: InsertOne", zap.String("sequence", s.name), zap.Error(err))
		}
		return err
	}

	s.curr = int(m["current"].(int32))
	if s.curr == -1 {
		s.curr = 1
	}
	return err
}

func (s *Sequence) save(ctx context.Context) (err error) {
	_, err = s.p.db.Collection(sequencesCollectionName).
		UpdateOne(
			ctx,
			bson.M{"_id": s.name}, bson.M{"$set": bson.M{"current": s.curr}},
		)
	return
}

func (msp *SequenceProvider) lookForSequence(seqName string) *Sequence {
	msp.seqMux.RLock()
	defer msp.seqMux.RUnlock()
	return msp.sequences[seqName]
}

func (msp *SequenceProvider) createSequence(seqName string) *Sequence {
	msp.seqMux.Lock()
	defer msp.seqMux.Unlock()
	se, ok := msp.sequences[seqName]
	if !ok {
		se = &Sequence{p: msp, name: seqName, curr: -1}
		msp.sequences[seqName] = se
	}
	return se
}
