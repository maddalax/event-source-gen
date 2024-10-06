package sourcing

import (
	"encoding/json"
	"eventgen/internal/database"
	"time"
)

type Version uint64

type Projection interface {
	OnEvent(event *BaseEvent[any])
	SetLastEventVersion(version Version)
	IsBehind() bool
	GetDiffBehind() Version
	GetLastEventVersion() Version
	Name() string
	IsHydrating() bool
	SetHydrating(hydrating bool)
	SetPaused(paused bool)
	IsPaused() bool
	IsCatchingUp() bool
	SetCatchingUp(catchingUp bool)
	Debug() bool
}

type SerializedEvent struct {
	Type          string `json:"type"`
	Payload       any    `json:"payload"`
	GlobalVersion uint64 `json:"globalVersion"`
	Version       uint64 `json:"Version"`
}

type Aggregate interface {
	Handle(command Command) error
	OnEvent(event Event)
}

type Command interface {
	GetId() string
}

type Iterator interface {
	HasNext() bool
	Next() ([]*BaseEvent[any], error)
	Close()
	Each(func(event *BaseEvent[any]))
}

type BaseEvent[T any] struct {
	Version       Version
	GlobalVersion Version
	Data          T
}

func Stringify(base *BaseEvent[any]) string {
	e := base.Data.(Event)
	t := e.GetType()
	id := e.GetId()
	data := map[string]any{
		"id":            id,
		"type":          t,
		"globalVersion": base.Version,
		"version":       base.Version,
		"payload":       e,
	}
	jsonStr, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(jsonStr)
}

type Event interface {
	GetId() string
	GetType() string
}

type EventStore interface {
	Save(events []Event, version Version) error
	Get(aggregateId string) (Iterator, error)
	All(start Version, cb func(event *BaseEvent[any])) error
	RunWithLock(key string, timeout time.Duration, tries int, cb func() error) error
}

var Store *EventStore

func Save(event []Event, version Version) error {
	err := (*Store).Save(event, version)
	if err != nil {
		return err
	}
	return nil
}

func Cursor(aggregateId string) (Iterator, error) {
	result, _ := (*Store).Get(aggregateId)
	return result, nil
}

func All(start Version, cb func(event *BaseEvent[any])) error {
	return (*Store).All(start, cb)
}

func SetEventStore(store EventStore) {
	Store = &store
}

func GetGlobalEventVersion() Version {
	latestEventVersion, err := database.Get[uint64]("global-event-version")
	if err != nil {
		return 0
	}
	return Version(*latestEventVersion)
}
