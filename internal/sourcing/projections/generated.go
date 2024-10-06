package projections

import "eventgen/internal/sourcing"
import "eventgen/internal/sourcing/events"

type ProjectionHandler interface {
	OnUserCreated(e *events.UserCreated)
	OnUserEmailChanged(e *events.UserEmailChanged)
	Debug() bool
	GetDiffBehind() sourcing.Version
	GetLastEventVersion() sourcing.Version
	IsBehind() bool
	IsCatchingUp() bool
	IsHydrating() bool
	IsPaused() bool
	Name() string
	SetCatchingUp(catchingUp bool)
	SetHydrating(hydrating bool)
	SetLastEventVersion(version sourcing.Version)
	SetPaused(paused bool)
}

func Handle(event *sourcing.BaseEvent[any], handler *ProjectionHandler) {
	switch e := event.Data.(type) {
	case events.UserCreated:
		(*handler).OnUserCreated(&e)
	case events.UserEmailChanged:
		(*handler).OnUserEmailChanged(&e)
	}
}

func GetProjections() []ProjectionHandler {
	return []ProjectionHandler{&EmailsUsed{}}
}
