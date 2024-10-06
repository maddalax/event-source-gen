package main

import (
	"encoding/json"
	"eventgen/internal/sourcing"
	"eventgen/internal/sourcing/eventstore"
	"eventgen/internal/sourcing/projections"
	"log"
	"time"
)

func main() {
	redis := eventstore.NewRedisEventStore()
	sourcing.SetEventStore(redis)

	projectionManager := projections.NewProjectionManager(projections.GetProjections())

	err := projectionManager.Hydrate()
	if err != nil {
		log.Fatal(err)
	}

	projectionManager.ListenForEvents()

	for {
		metrics := projectionManager.GetMetrics()
		for i, metric := range metrics {
			serialized, _ := json.Marshal(metric)
			log.Printf("Projection %d: %s\n", i, string(serialized))
		}
		time.Sleep(5 * time.Second)
	}
}
