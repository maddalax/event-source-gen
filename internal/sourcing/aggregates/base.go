package aggregates

import (
	"eventgen/internal/sourcing"
	"log"
)

func LoadEvents(aggregateId string, agg any) {
	iter, _ := sourcing.Cursor(aggregateId)
	iter.Each(func(e *sourcing.BaseEvent[any]) {
		err := LoadEvent(agg, e)
		if err != nil {
			log.Fatal(err)
		}
	})
}
