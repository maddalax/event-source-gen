package aggregates

import "eventgen/internal/sourcing/commands"
import "eventgen/internal/sourcing/events"
import "eventgen/internal/sourcing"
import "errors"

func Handle(command sourcing.Command) error {
	switch c := command.(type) {
	case *commands.CreateUser:
		agg := User{}
		LoadEvents(command.GetId(), &agg)
		agg.Version = agg.Version + 1
		e0, err := agg.HandleCreateUser(c)
		if err != nil {
			return err
		}
		if e0 == nil {
			return errors.New("command.CreateUser -> command.Handle -> events.UserCreated is nil. ensure it is returned from the aggregate")
		}
		agg.AppendAndCommit(e0)
		return nil
	case *commands.UpdateEmailUser:
		agg := User{}
		LoadEvents(command.GetId(), &agg)
		agg.Version = agg.Version + 1
		e0, err := agg.HandleUpdateEmailUser(c)
		if err != nil {
			return err
		}
		if e0 == nil {
			return errors.New("command.UpdateEmailUser -> command.Handle -> events.UserCreated is nil. ensure it is returned from the aggregate")
		}
		agg.AppendAndCommit(e0)
		return nil
	default:
		return errors.New("unknown command type")
	}
}
func LoadEvent(agg any, event *sourcing.BaseEvent[any]) error {
	switch a := agg.(type) {
	case *User:
		switch e := event.Data.(type) {
		case *events.UserCreated:
			a.Version = event.Version
			return a.OnUserCreated(e)
		default:
			return nil
		}
	default:
		return nil
	}
}
