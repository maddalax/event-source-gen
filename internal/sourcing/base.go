package sourcing

import (
	"log"
)

type AggregateBase struct {
	Version Version
	events  []Event
}

func (a *AggregateBase) appendEvent(event Event) {
	if event == nil {
		log.Fatalf("event is nil\n")
		return
	}
	a.events = append(a.events, event)
}

func (a *AggregateBase) getUncommittedEvents() []Event {
	return a.events
}

func (a *AggregateBase) clearUncommittedEvents() {
	a.events = []Event{}
}

func (a *AggregateBase) commitEvents() {
	l := len(a.events) - 1
	firstEventVersion := a.Version - Version(l)
	err := Save(a.events, firstEventVersion)
	if err != nil {
		log.Fatalf("error saving events: %v\n", err)
		return
	}
	a.clearUncommittedEvents()
}

func (a *AggregateBase) AppendAndCommit(event Event) {
	a.appendEvent(event)
	a.commitEvents()
}

func (a *AggregateBase) Append(event Event) {
	a.appendEvent(event)
}

func (a *AggregateBase) AppendManyAndCommit(events ...Event) {
	for _, event := range events {
		a.appendEvent(event)
	}
	a.commitEvents()
}

func (a *AggregateBase) Commit() {
	a.commitEvents()
}
