package eventstore

import (
	"context"
	"eventgen/internal/database"
	"eventgen/internal/sourcing"
	"eventgen/internal/util"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisEventStore struct {
	lock *redislock.Client
}

func NewRedisEventStore() *RedisEventStore {
	lock := redislock.New(database.Connect())
	return &RedisEventStore{
		lock: lock,
	}
}

func (r *RedisEventStore) RunWithLock(key string, timeout time.Duration, tries int, cb func() error) error {
	ctx := context.Background()
	lock, err := r.lock.Obtain(ctx, key, timeout, nil)
	if err != nil {
		if tries <= 0 {
			return fmt.Errorf("could not obtain lock: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		return r.RunWithLock(key, timeout, tries-1, cb)
	}
	defer lock.Release(ctx)
	return cb()
}

func (r *RedisEventStore) Save(events []sourcing.Event, version sourcing.Version) error {
	ctx := context.Background()
	aggregateIds := make(map[string]bool)
	for _, event := range events {
		aggregateIds[event.GetId()] = true
	}

	if len(aggregateIds) > 1 {
		return fmt.Errorf("cannot save events for multiple aggregates")
	}

	aggregateId := events[0].GetId()

	if aggregateId == "" {
		return fmt.Errorf("aggregate id is empty")
	}

	streamName := fmt.Sprintf("events-stream-%s", aggregateId)
	lockName := fmt.Sprintf("event-stream-lock")

	l, err := r.lock.Obtain(ctx, lockName, time.Second*3, nil)

	if err != nil {
		fmt.Printf("could not obtain lock: %v\n", err)
		time.Sleep(100 * time.Millisecond)
		return r.Save(events, version)
	}

	defer l.Release(ctx)

	rdb := database.Connect()

	rdb.HSet(context.Background(), "event-streams", streamName, time.Now().Unix())

	lastVersion := rdb.XLen(context.Background(), streamName).Val()

	fmt.Printf("last version: %d, new version: %d\n", lastVersion, version)

	if lastVersion+1 != int64(version) {
		return fmt.Errorf("version mismatch: expected %d, got %d", lastVersion+1, version)
	}

	for i := range events {
		e := events[i]

		serialized := util.ToJson(e)
		globalVersion := rdb.Incr(context.Background(), "global-event-version").Val()

		data := map[string]interface{}{
			"type":          e.GetType(),
			"globalVersion": globalVersion,
			"version":       int64(version) + int64(i),
			"payload":       serialized,
		}

		id, err := rdb.XAdd(context.Background(), &redis.XAddArgs{
			Stream: streamName,
			Values: data,
		}).Result()

		if err != nil {
			return err
		}

		rdb.XAdd(context.Background(), &redis.XAddArgs{
			Stream: "latest-events-pubsub",
			Values: data,
		})

		rdb.RPush(context.Background(), "global-event-ordering", fmt.Sprintf("%s:id:%s", streamName, id))
	}

	return nil
}

func (r *RedisEventStore) Get(aggregateId string) (sourcing.Iterator, error) {
	rdb := database.Connect()
	streamName := fmt.Sprintf("events-stream-%s", aggregateId)
	return NewStreamIterator(rdb, streamName, 50), nil
}

func (r *RedisEventStore) All(start sourcing.Version, cb func(event *sourcing.BaseEvent[any])) error {
	rdb := database.Connect()
	ctx := context.Background()

	startInt := int64(start)
	endInt := int64(start + 50)
	for {
		result := rdb.LRange(context.Background(), "global-event-ordering", startInt, endInt)
		streams := result.Val()

		if len(streams) == 0 {
			break
		}

		startInt = endInt + 1
		endInt = startInt + 50

		for _, stream := range streams {
			split := strings.Split(stream, ":id:")
			streamId := split[0]
			messageId := split[1]
			entry, err := rdb.XRange(ctx, streamId, messageId, messageId).Result()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			events := toEvent(entry)
			for _, event := range events {
				cb(event)
			}
		}
	}

	return nil
}
