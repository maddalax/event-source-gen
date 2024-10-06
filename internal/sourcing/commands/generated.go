package commands

import "eventgen/internal/validators"
import "eventgen/internal/sourcing/events"

type CreateUser struct {
	Email     string
	FirstName string
	Id        string
	LastName  string
}

func (c *CreateUser) ToUserCreated() *events.UserCreated {
	var event *events.UserCreated = &events.UserCreated{}
	event.Email = c.Email
	event.FirstName = c.FirstName
	event.Id = c.Id
	event.LastName = c.LastName
	return event
}

func (c *CreateUser) GetId() string {
	return c.Id
}

func (c *CreateUser) Validate() error {
	if err := validators.IsEmpty("Id", c.Id); err != nil {
		return err
	}
	return nil
}

func NewCreateUser(Email string, FirstName string, Id string, LastName string) *CreateUser {
	result := &CreateUser{}
	result.Email = Email
	result.FirstName = FirstName
	result.Id = Id
	result.LastName = LastName
	return result
}

type UpdateEmailUser struct {
	Email string
	Id    string
}

func (c *UpdateEmailUser) ToUserCreated() *events.UserCreated {
	var event *events.UserCreated = &events.UserCreated{}
	event.Email = c.Email
	event.Id = c.Id
	return event
}

func (c *UpdateEmailUser) GetId() string {
	return c.Id
}

func (c *UpdateEmailUser) Validate() error {
	if err := validators.IsEmpty("Id", c.Id); err != nil {
		return err
	}
	return nil
}

func NewUpdateEmailUser(Email string, Id string) *UpdateEmailUser {
	result := &UpdateEmailUser{}
	result.Email = Email
	result.Id = Id
	return result
}
