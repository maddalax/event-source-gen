package database

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	once sync.Once
	rdb  *redis.Client
)

func Connect() *redis.Client {
	once.Do(func() {
		var ctx = context.Background()
		var err error
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		if err != nil {
			panic(err)
		}

		cmd := rdb.Ping(ctx)

		if cmd.Err() != nil {
			panic(err)
		}
	})
	return rdb
}

func Pipeline(cb func(pipe redis.Pipeliner)) error {
	db := Connect()
	pipe := db.Pipeline()
	cb(pipe)
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func Set[T any](key string, value T) error {
	db := Connect()
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	result := db.Set(context.Background(), key, serialized, time.Duration(0))
	return result.Err()
}

func HSet[T any](set string, key string, value T) error {
	db := Connect()
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	result := db.HSet(context.Background(), set, key, serialized)
	return result.Err()
}

func Get[T any](key string) (*T, error) {
	db := Connect()
	val, err := db.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	result := new(T)
	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func HGet[T any](set string, key string) (*T, error) {
	db := Connect()
	val, err := db.HGet(context.Background(), set, key).Result()
	if err != nil {
		return nil, err
	}
	result := new(T)
	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func HExists(set string, key string) (bool, error) {
	db := Connect()
	e := db.HExists(context.Background(), set, key)
	return e.Result()
}

func HList[T any](key string) ([]*T, error) {
	db := Connect()
	val, err := db.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	result := make([]*T, len(val))

	count := 0
	for _, t := range val {
		item := new(T)
		err = json.Unmarshal([]byte(t), item)
		if err != nil {
			return nil, err
		}
		result[count] = item
		count++
	}
	return result, nil
}

// LPush pushes a value onto the head of a list.
func LPush[T any](list string, value T) error {
	db := Connect()
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	result := db.LPush(context.Background(), list, serialized)
	return result.Err()
}

// RPush pushes a value onto the tail of a list.
func RPush[T any](list string, value T) error {
	db := Connect()
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	result := db.RPush(context.Background(), list, serialized)
	return result.Err()
}

// LPop pops a value from the head of a list.
func LPop[T any](list string) (*T, error) {
	db := Connect()
	val, err := db.LPop(context.Background(), list).Result()
	if err != nil {
		return nil, err
	}
	result := new(T)
	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RPop pops a value from the tail of a list.
func RPop[T any](list string) (*T, error) {
	db := Connect()
	val, err := db.RPop(context.Background(), list).Result()
	if err != nil {
		return nil, err
	}
	result := new(T)
	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LRange retrieves elements from a list within the specified range.
func LRange[T any](list string, start, stop int64) ([]*T, error) {
	db := Connect()
	vals, err := db.LRange(context.Background(), list, start, stop).Result()
	if err != nil {
		return nil, err
	}
	result := make([]*T, len(vals))
	for i, v := range vals {
		item := new(T)
		err = json.Unmarshal([]byte(v), item)
		if err != nil {
			return nil, err
		}
		result[i] = item
	}
	return result, nil
}
