package mongo

import (
	"context"
	"errors"
	"sync"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
	dep "github.com/vc2402/vivard/dependencies"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/vc2402/vivard"
	"go.mongodb.org/mongo-driver/mongo"
)

const sequencesCollectionName = "_sequences"
const sequencesAreCacheable = true

// Sequence allows to create sequential numbers
//  current value is cached!
type Sequence struct {
	p    *SequenceProvider
	name string
	curr int
}

type SequenceProvider struct {
	db        *mongo.Database
	log       *logrus.Entry
	sequences map[string]*Sequence
	seqMux    sync.RWMutex
}

func MongoSequenceForDB(db *mongo.Database) *SequenceProvider {
	return &SequenceProvider{db: db}
}

func (msp *SequenceProvider) Prepare(eng *vivard.Engine, prov dep.Provider) (err error) {
	msp.log = prov.Logger("mng-seq")
	msp.sequences = map[string]*Sequence{}
	if msp.db == nil {
		mongo, ok := eng.GetService(ServiceMongo).(*Service)
		if !ok {
			return errors.New("MongoService is required for SequenceProvider")
		}
		if mongo.db == nil {
			err = mongo.Prepare(eng, prov)
			if err != nil {
				return
			}
		}

		msp.db = mongo.DB()
	}
	return
}

func (msp *SequenceProvider) Start(eng *vivard.Engine, prov dep.Provider) error {
	if msp.db == nil {
		return errors.New("SequenceProvider is not initialized")
	}
	return nil
}

func (msp *SequenceProvider) Provide() interface{} {
	return msp
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
	s.curr++
	err = s.save(ctx)
	if err != nil {
		s.p.log.Warnf("Sequence<%s>.Next: Update: %v", s.name, err)
		return -1, err
	}
	return s.curr, nil
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

//SetCurrent sets current value of Sequence to value
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
		s.p.log.Warnf("Sequence<%s>.Next: Update: %v", s.name, err)
		return -1, err
	}
	return s.curr, nil
}

func (s *Sequence) load(ctx context.Context) error {
	m := map[string]interface{}{}
	err := s.p.db.Collection(sequencesCollectionName).
		FindOne(ctx,
			bson.M{"_id": s.name},
			options.FindOne().SetProjection(bson.D{{"current", 1}}),
		).
		Decode(&m)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			s.p.log.Warnf("Sequence<%s>.load: FindOne: %v", s.name, err)
			return err
		}
		s.curr = 1
		s.p.log.Tracef("Sequence<%s>.load: Iniatializing", s.name)
		_, err := s.p.db.Collection(sequencesCollectionName).InsertOne(ctx, bson.M{"_id": s.name, "current": s.curr})
		if err != nil {
			s.p.log.Warnf("Sequence<%s>.load: Insert: %v", s.name, err)
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
		UpdateOne(ctx,
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
