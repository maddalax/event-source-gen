package main

import (
	"fmt"
)

func GenerateCommandsFile(parsed *Parsed) string {
	builder := NewCodeBuilder(nil)
	builder.AppendLine("package commands")

	for _, commandImport := range SortImports(&CommandImports) {
		builder.AddImport(commandImport)
	}

	builder.AddImport(GetSourcingPath("events"))

	for _, command := range parsed.Commands {
		fields := command.GetFields2()

		_struct := Struct{
			Name:   command.StructName(),
			Fields: fields,
		}
		builder.Append(builder.BuildStruct(_struct))

		/*
			func (c *CreatePatient) ToPatientCreated() *events.PatientCreated {
				var event *events.PatientCreated = &events.PatientCreated{}
				event.Email = c.Email
				event.Id = c.Id
				return event
			}
		*/
		for _, event := range command.events {
			eventStructPath := fmt.Sprintf("events.%s", event.StructName())
			eventName := fmt.Sprintf("To%s", event.StructName())
			builder.Append(builder.BuildToEventMethod(_struct.Name, eventName, eventStructPath, fields))
		}

		/*
			func (c *CreatePatient) GetId() string {
				return c.Id
			}
		*/
		builder.Append(builder.BuildMethod(Method{
			StructReceiver: "c",
			StructName:     _struct.Name,
			MethodName:     "GetId",
			Parameters:     []NameType{},
			Body:           fmt.Sprintf("return c.%s", command.GetIdField().Name),
			Return: []ReturnType{
				{Type: "string"},
			},
		}))

		/*
			func (c *CreatePatient) Validate() error {
				if err := validators.IsEmpty("Id", c.Id); err != nil {
					return err
				}
				return nil
			}
		*/
		builder.Append(builder.BuildValidation(_struct.Name, "Validate",
			FlatMap(command.GetFields(), func(f Field) []Validation {
				return f.Validations
			}),
		))

		/*
			func NewCreatePatient(Email string, Id string) *CreatePatient {
				result := &CreatePatient{}
				result.Email = Email
				result.Id = Id
				return result
			}
		*/
		builder.Append(builder.BuildNewCommandMethod(_struct))
	}

	return builder.String()
}
