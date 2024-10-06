package aggregates

import (
	"eventgen/internal/sourcing"
	"eventgen/internal/sourcing/commands"
	"eventgen/internal/sourcing/events"
)

type User struct{ sourcing.AggregateBase }

func (a *User) OnUserEmailChanged(e *events.UserEmailChanged) error {
	return nil
}
func (a *User) HandleCreateUser(c *commands.CreateUser) (*events.UserCreated, error) {
	return nil, nil
}
func (a *User) HandleUpdateEmailUser(c *commands.UpdateEmailUser) (*events.UserCreated, error) {
	return nil, nil
}
func (a *User) OnUserCreated(e *events.UserCreated) error {
	return nil
}
