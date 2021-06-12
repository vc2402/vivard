package vivard

import "fmt"

type KVPair struct {
	Key interface{}
	Val interface{}
}

type ArrKV []KVPair

func MapStringStringToArrKV(m map[string]string) ArrKV {
	ret := make([]KVPair, len(m))
	i := 0
	for k, v := range m {
		ret[i] = KVPair{Key: k, Val: v}
		i++
	}
	return ret
}

func (akv ArrKV) ToMapStringString() (map[string]string, error) {
	ret := map[string]string{}
	var k, v string
	var err error
	for _, kv := range akv {
		k, err = kv.KeyString()
		if err != nil {
			return nil, err
		}
		v, err = kv.KeyString()
		if err != nil {
			return nil, err
		}
		ret[k] = v
	}
	return ret, nil
}

func (kv KVPair) KeyString() (string, error) {
	if s, ok := kv.Key.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("not a string: %v (%T)", kv.Key, kv.Key)
}

func (kv KVPair) ValString() (string, error) {
	if s, ok := kv.Val.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("not a string: %v (%T)", kv.Val, kv.Val)
}

func (kv KVPair) KeyStr() string {
	if s, ok := kv.Key.(string); ok {
		return s
	}
	return ""
}

func (kv KVPair) ValStr() string {
	if s, ok := kv.Val.(string); ok {
		return s
	}
	return ""
}

func MapStringIntToArrKV(m map[string]int) ArrKV {
	ret := make([]KVPair, len(m))
	i := 0
	for k, v := range m {
		ret[i] = KVPair{Key: k, Val: v}
		i++
	}
	return ret
}

func (akv ArrKV) ToMapStringInt() (map[string]int, error) {
	ret := map[string]int{}
	var k string
	var v int
	var err error
	for _, kv := range akv {
		k, err = kv.KeyString()
		if err != nil {
			return nil, err
		}
		v, err = kv.ValInteger()
		if err != nil {
			return nil, err
		}
		ret[k] = v
	}
	return ret, nil
}

func (kv KVPair) KeyInteger() (int, error) {
	if s, ok := kv.Key.(int); ok {
		return s, nil
	}
	return 0, fmt.Errorf("not a int: %v (%T)", kv.Key, kv.Key)
}

func (kv KVPair) ValInteger() (int, error) {
	if s, ok := kv.Val.(int); ok {
		return s, nil
	}
	return 0, fmt.Errorf("not a int: %v (%T)", kv.Val, kv.Val)
}

func (kv KVPair) KeyInt() int {
	if s, ok := kv.Key.(int); ok {
		return s
	}
	return 0
}

func (kv KVPair) ValInt() int {
	if s, ok := kv.Val.(int); ok {
		return s
	}
	return 0
}
