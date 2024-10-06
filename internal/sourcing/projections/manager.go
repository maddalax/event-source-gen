package projections

import (
	"eventgen/internal/sourcing"
	"fmt"
	"sync"
	"time"
)

type ProjectionMetrics struct {
	Name                 string
	LatestProcessedEvent sourcing.Version
	GlobalEventVersion   sourcing.Version
	IsCatchingUp         bool
	IsHydrating          bool
	IsPaused             bool
	DiffBehind           int64
}

type ProjectionManager struct {
	handlers []ProjectionHandler
	Paused   bool
}

func NewProjectionManager(handlers []ProjectionHandler) *ProjectionManager {
	return &ProjectionManager{
		handlers: handlers,
	}
}

func runEvent(source string, event *sourcing.BaseEvent[any], handler *ProjectionHandler) {

	h := *handler

	if h.Debug() {
		fmt.Printf("%s received event from %s: %s\n", h.Name(), source, sourcing.Stringify(event))
	}

	h.SetLastEventVersion(event.GlobalVersion)
	Handle(event, &h)
}

func (m *ProjectionManager) GetMetrics() []ProjectionMetrics {
	var metrics []ProjectionMetrics
	globalVersion := sourcing.GetGlobalEventVersion()
	for _, handler := range m.handlers {
		metrics = append(metrics, ProjectionMetrics{
			Name:                 handler.Name(),
			IsPaused:             handler.IsPaused(),
			LatestProcessedEvent: handler.GetLastEventVersion(),
			IsCatchingUp:         handler.IsCatchingUp(),
			IsHydrating:          handler.IsHydrating(),
			GlobalEventVersion:   globalVersion,
			DiffBehind:           int64(handler.GetDiffBehind()),
		})
	}
	return metrics
}

func (m *ProjectionManager) ListenForEvents() {
	go func() {
		for {
			if !m.Paused {
				m.ReplayBehind()
			}
			time.Sleep(150 * time.Millisecond)
		}
	}()

}

func (m *ProjectionManager) Hydrate() error {
	for _, projection := range m.handlers {
		projection.SetHydrating(true)
	}

	// Hydrate all the projections on startup
	err := sourcing.All(0, func(event *sourcing.BaseEvent[any]) {
		wg := sync.WaitGroup{}
		for _, handler := range m.handlers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runEvent("startup_hydration", event, &handler)
			}()
		}
		wg.Wait()
	})

	if err != nil {
		return err
	}

	for _, projection := range m.handlers {
		projection.SetHydrating(false)
	}

	return nil
}

func (m *ProjectionManager) ReplayBehind() {
	hasAnyBehind := false
	lastGlobal := sourcing.GetGlobalEventVersion()
	behindMap := make(map[string]sourcing.Version)
	minBehind := lastGlobal

	for _, projection := range m.handlers {
		lastVersion := projection.GetLastEventVersion()
		if lastVersion < lastGlobal {

			if lastVersion < minBehind {
				minBehind = lastVersion
			}

			projection.SetHydrating(false)
			projection.SetCatchingUp(true)
			behindMap[projection.Name()] = lastVersion
			hasAnyBehind = true
		}
	}

	if !hasAnyBehind {
		return
	}

	_ = sourcing.All(minBehind, func(event *sourcing.BaseEvent[any]) {
		wg := sync.WaitGroup{}
		for _, handler := range m.handlers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				d := handler
				latestProjectionVersion, exists := behindMap[d.Name()]
				if !exists {
					return
				}
				if event.GlobalVersion <= latestProjectionVersion {
					return
				}
				runEvent("replay_behind", event, &handler)
			}()
		}
		wg.Wait()
	})

	for _, projection := range m.handlers {
		projection.SetCatchingUp(false)
	}
}
