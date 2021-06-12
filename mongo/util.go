package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

func MapToBsonM(m interface{}) (bson.M, error) {
	ret := bson.M{}
	switch m.(type) {
	case *map[string]interface{}:
		mm, ok := m.(*map[string]interface{})
		if !ok {
			return nil, errors.New("problem while converting type")
		}
		if mm == nil {
			return nil, nil
		}
		for k, v := range *mm {
			ret[k] = v
		}
	case bson.M:
		return m.(bson.M), nil
	default:
		return nil, errors.New("can not convert type")
	}
	return ret, nil
}
