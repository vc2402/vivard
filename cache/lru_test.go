package cache

import (
	"fmt"
	"reflect"
	"testing"
)

var testCache *LRU[int, int]

const size = 10

func init() {
	testCache = NewLRU[int, int](size)
	for i := 0; i < size; i++ {
		testCache.Set(i+1, size-i)
	}
}

func TestCache_Get(t *testing.T) {
	type args[K Key] struct {
		key K
	}
	type testCase[K Key, V any] struct {
		name      string
		c         *LRU[K, V]
		args      args[K]
		wantValue V
		headValue V
	}
	tests := []testCase[int, int]{
		{
			name:      "first",
			c:         testCache,
			args:      args[int]{1},
			wantValue: size,
			headValue: size,
		},
		{
			name:      "last",
			c:         testCache,
			args:      args[int]{size},
			wantValue: 1,
			headValue: 1,
		},
		{
			name:      "any",
			c:         testCache,
			args:      args[int]{2},
			wantValue: size - 2 + 1,
			headValue: size - 2 + 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := tt.c.Get(tt.args.key); !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("Get() = %v, want %v", gotValue, tt.wantValue)
			}
			if tt.c.head.value != tt.headValue {
				t.Errorf("head = %v, want %v", tt.c.head.value, tt.headValue)
			}
			err := tt.c.checkList()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCache_Remove(t *testing.T) {
	type args[K Key] struct {
		key K
	}
	type testCase[K Key, V any] struct {
		name      string
		c         *LRU[K, V]
		args      args[K]
		wantValue V
		wantSize  int
	}
	tests := []testCase[int, int]{
		{
			name:      "remove first",
			c:         testCache,
			args:      args[int]{10},
			wantValue: 1,
			wantSize:  size - 1,
		},
		{
			name:      "remove middle",
			c:         testCache,
			args:      args[int]{3},
			wantValue: size - 3 + 1,
			wantSize:  size - 2,
		},
		{
			name:      "remove last",
			c:         testCache,
			args:      args[int]{1},
			wantValue: 10,
			wantSize:  size - 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValue := tt.c.Remove(tt.args.key); !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("Remove() = %v, want %v", gotValue, tt.wantValue)
			}
			if len(tt.c.index) != tt.wantSize {
				t.Errorf("size = %v, want %v", len(tt.c.index), tt.wantSize)
			}
			err := tt.c.checkList()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCache_Put(t *testing.T) {
	type args[K Key, V any] struct {
		key   K
		value V
	}
	type testCase[K Key, V any] struct {
		name string
		c    *LRU[K, V]
		args args[K, V]
		head V
	}
	tests := []testCase[int, int]{
		{
			name: "put 6",
			c:    testCache,
			args: args[int, int]{6, 6},
			head: 6,
		},
		{
			name: "put 7",
			c:    testCache,
			args: args[int, int]{7, 7},
			head: 7,
		},
		{
			name: "replace 6",
			c:    testCache,
			args: args[int, int]{6, 8},
			head: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Set(tt.args.key, tt.args.value)
			if len(tt.c.index) > tt.c.size {
				t.Errorf("len = %v, size %v", len(tt.c.index), tt.c.size)
			}
			if tt.c.head.value != tt.head {
				t.Errorf("head = %v, want %v", tt.c.head.value, tt.head)
			}
			err := tt.c.checkList()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCache_addHead(t *testing.T) {
	type args[K Key, V any] struct {
		item *lruItem[K, V]
	}
	type testCase[K Key, V any] struct {
		name string
		c    *LRU[K, V]
		args args[K, V]
	}
	tests := []testCase[int, int]{
		{
			name: "add head",
			c:    testCache,
			args: args[int, int]{
				&lruItem[int, int]{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.addHead(tt.args.item)
			err := tt.c.checkList()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCache_moveToHead(t *testing.T) {
	type args[K Key, V any] struct {
		item *lruItem[K, V]
	}
	type testCase[K Key, V any] struct {
		name string
		c    *LRU[K, V]
		args args[K, V]
	}
	tests := []testCase[int, int]{
		{
			name: "move to head head",
			c:    testCache,
			args: args[int, int]{
				testCache.head,
			},
		},
		{
			name: "move to head tail",
			c:    testCache,
			args: args[int, int]{
				testCache.tail,
			},
		},
		{
			name: "move to head",
			c:    testCache,
			args: args[int, int]{
				testCache.head.next,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.moveToHead(tt.args.item)
			err := tt.c.checkList()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCache_remove(t *testing.T) {
	type args[K Key, V any] struct {
		item *lruItem[K, V]
	}
	type testCase[K Key, V any] struct {
		name string
		c    *LRU[K, V]
		args args[K, V]
	}
	tests := []testCase[int, int]{
		{
			name: "remove head",
			c:    testCache,
			args: args[int, int]{
				testCache.head,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.remove(tt.args.item)
			err := tt.c.checkList()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCache_removeTail(t *testing.T) {
	type testCase[K Key, V any] struct {
		name     string
		c        *LRU[K, V]
		wantItem *lruItem[K, V]
	}
	tests := []testCase[int, int]{
		{
			name:     "remove tail",
			c:        testCache,
			wantItem: testCache.tail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotItem := tt.c.removeTail(); !reflect.DeepEqual(gotItem, tt.wantItem) {
				t.Errorf("removeTail() = %v, want %v", gotItem, tt.wantItem)
				err := tt.c.checkList()
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func (c *LRU[K, V]) checkList() error {
	var prev *lruItem[K, V]
	count := 0
	for it := c.head; it != nil; it = it.next {
		if prev != nil {
			if it.prev != prev {
				return fmt.Errorf("checkList: at %d: prev incorrect", count)
			}
			if prev.next != it {
				return fmt.Errorf("checkList: at %d: next incorrect", count)
			}
		}
		prev = it
		count++
	}
	if prev != c.tail {
		return fmt.Errorf("checkList: tail incorrect")
	}
	return nil
}
