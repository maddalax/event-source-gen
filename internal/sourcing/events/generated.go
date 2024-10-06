package events

import "encoding/json"
import "fmt"

type UserCreated struct {
	Email     string
	FirstName string
	Id        string
	LastName  string
}

func (e *UserCreated) GetId() string {
	return e.Id
}

func (e *UserCreated) GetType() string {
	return "UserCreated"
}

type UserEmailChanged struct {
	Email string
	Id    string
}

func (e *UserEmailChanged) GetId() string {
	return e.Id
}

func (e *UserEmailChanged) GetType() string {
	return "UserEmailChanged"
}
func Deserialize(data string, eventType string) (any, error) {
	if eventType == "UserCreated" {
		var d UserCreated
		err := json.Unmarshal([]byte(data), &d)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
	if eventType == "UserEmailChanged" {
		var d UserEmailChanged
		err := json.Unmarshal([]byte(data), &d)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
	if eventType == "MyNewEvent" {
		var d UserCreated
		err := json.Unmarshal([]byte(data), &d)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
	return nil, fmt.Errorf("unknown event type: %s", eventType)
}
