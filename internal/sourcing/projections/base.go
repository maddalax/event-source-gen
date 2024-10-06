package projections

import (
	"eventgen/internal/database"
	"eventgen/internal/sourcing"
	"eventgen/internal/sourcing/events"
	"reflect"
)

type BaseProjection struct {
	hydrating  bool
	paused     bool
	catchingUp bool
}

func (p *BaseProjection) Name() string {
	return reflect.TypeOf(p).String()
}
func (p *BaseProjection) SetPaused(paused bool) {
	p.paused = paused
}
func (p *BaseProjection) SetHydrating(hydrating bool) {
	p.hydrating = hydrating
}
func (p *BaseProjection) IsPaused() bool {
	return p.paused
}
func (p *BaseProjection) IsCatchingUp() bool {
	return p.catchingUp
}
func (p *BaseProjection) SetCatchingUp(catchingUp bool) {
	p.catchingUp = catchingUp
}
func (p *BaseProjection) SetLastEventVersion(version sourcing.Version) {
	err := database.HSet("projection-last-event", p.Name(), version)
	if err != nil {
		return
	}
}
func (p *BaseProjection) GetLastEventVersion() sourcing.Version {
	data, err := database.HGet[uint64]("projection-last-event", p.Name())
	if err != nil {
		return 0
	}
	return sourcing.Version(*data)
}
func (p *BaseProjection) GetGlobalEventVersion() sourcing.Version {
	return sourcing.GetGlobalEventVersion()
}
func (p *BaseProjection) IsHydrating() bool {
	return p.hydrating
}
func (p *BaseProjection) IsBehind() bool {
	if p.IsHydrating() {
		return false
	}
	return p.GetLastEventVersion() < p.GetGlobalEventVersion()
}
func (p *BaseProjection) GetDiffBehind() sourcing.Version {
	if p.IsHydrating() {
		return 0
	}
	return p.GetGlobalEventVersion() - p.GetLastEventVersion()
}
func (p *BaseProjection) Debug() bool {
	return false
}
func (p *BaseProjection) OnUserCreated(e *events.UserCreated) {
}
func (p *BaseProjection) OnUserEmailChanged(e *events.UserEmailChanged) {
}
