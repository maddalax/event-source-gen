package eventstore

import (
	"context"
	"eventgen/internal/sourcing"
	events2 "eventgen/internal/sourcing/events"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
)

// StreamIterator is a simple iterator to read through all entries in a Redis Stream
type StreamIterator struct {
	rdb       *redis.Client
	stream    string
	lastID    string
	batchSize int64
}

// NewStreamIterator creates a new StreamIterator
func NewStreamIterator(rdb *redis.Client, stream string, batchSize int64) *StreamIterator {
	return &StreamIterator{
		rdb:       rdb,
		stream:    stream,
		lastID:    "0", // Start from the beginning of the stream
		batchSize: batchSize,
	}
}

// HasNext checks if there are more entries to read in the stream
func (it *StreamIterator) HasNext() bool {
	return it.lastID != ""
}

func toEvent(messages []redis.XMessage) []*sourcing.BaseEvent[any] {
	events := make([]*sourcing.BaseEvent[any], len(messages))
	for i, message := range messages {
		t := message.Values["type"].(string)
		version, _ := strconv.ParseUint(message.Values["version"].(string), 10, 64)
		globalVersion, _ := strconv.ParseUint(message.Values["globalVersion"].(string), 10, 64)
		instance, err := events2.Deserialize(message.Values["payload"].(string), t)
		if err != nil {
			log.Fatalf("could not unmarshal event: %v", err)
		}
		if instance == nil {
			log.Fatalf(fmt.Sprintf("could not find event type: %s", t))
		}
		events[i] = &sourcing.BaseEvent[any]{
			Version:       sourcing.Version(version),
			GlobalVersion: sourcing.Version(globalVersion),
			Data:          instance,
		}
	}
	return events
}

func (it *StreamIterator) Close() {
	// noop
}

func (it *StreamIterator) Each(cb func(event *sourcing.BaseEvent[any])) {
	for {
		if !it.HasNext() {
			break
		}
		entries, err := it.Next()
		if err != nil {
			log.Fatalf("could not read entries: %v", err)
		}
		for _, entry := range entries {
			cb(entry)
		}
	}
}

// Next fetches the next batch of entries from the stream
func (it *StreamIterator) Next() ([]*sourcing.BaseEvent[any], error) {
	entries, err := it.rdb.XRead(context.Background(), &redis.XReadArgs{
		Streams: []string{it.stream, it.lastID},
		Count:   it.batchSize,
		Block:   -1,
	}).Result()

	if entries == nil {
		it.lastID = ""
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if len(entries) > 0 && len(entries[0].Messages) > 0 {
		lastEntry := entries[0].Messages[len(entries[0].Messages)-1]
		it.lastID = lastEntry.ID
		return toEvent(entries[0].Messages), nil
	}

	// No more entries to read
	it.lastID = ""
	return nil, nil
}
