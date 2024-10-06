package projections

import "eventgen/internal/sourcing/events"

type EmailsUsed struct{ BaseProjection }

func (a *EmailsUsed) OnUserCreated(e *events.UserCreated) {
	return
}
func (a *EmailsUsed) OnUserEmailChanged(e *events.UserEmailChanged) {
	return
}
