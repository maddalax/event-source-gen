package main

import "fmt"

func GenerateEventsFile(parsed *Parsed) string {
	builder := NewCodeBuilder(nil)
	builder.AppendLine("package events")

	for _, eventImport := range SortImports(&EventImports) {
		builder.AddImport(eventImport)
	}

	builder.AddImport("encoding/json")
	builder.AddImport("fmt")

	for _, event := range parsed.Events {
		buildEvent(&event, builder)
	}

	buildDeserialize(parsed, builder)

	return builder.String()
}

func buildDeserialize(parsed *Parsed, builder *CodeBuilder) {

	eventMap := NewOrderedMap[string, string]()

	for _, event := range parsed.Events {
		eventMap.Set(event.StructName(), event.StructName())
	}

	for eventFrom, to := range parsed.EventMigrations {
		for _, s := range to {
			f, _ := eventMap.Get(eventFrom)
			eventMap.Set(s, f)
		}
	}

	body := ""

	for _, entry := range eventMap.Entries() {
		eventName := entry.Key
		structName := entry.Value

		body += fmt.Sprintf(`
			if eventType == "%s" {
				var d %s
				err := json.Unmarshal([]byte(data), &d)
				if err != nil {
					return nil, err
				}
				return d, nil
			}
		`, eventName, structName)
	}

	body += `
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	`

	f := Function{
		Name: "Deserialize",
		Parameters: []NameType{
			{Name: "data", Type: "string"},
			{Name: "eventType", Type: "string"},
		},
		Return: []ReturnType{
			{Type: "any"},
			{Type: "error"},
		},
		Body: body,
	}

	builder.AppendLine(builder.BuildFunction(f))
}

func buildEvent(event *Event, builder *CodeBuilder) {
	_struct := Struct{
		Name:   event.StructName(),
		Fields: event.GetFields2(),
	}
	builder.Append(builder.BuildStruct(_struct))

	builder.Append(builder.BuildMethod(Method{
		StructReceiver: "e",
		StructName:     _struct.Name,
		MethodName:     "GetId",
		Body:           fmt.Sprintf("return e.%s", event.GetIdField().Name),
		Return: []ReturnType{
			{Type: "string"},
		},
	}))

	builder.Append(builder.BuildMethod(Method{
		StructReceiver: "e",
		StructName:     _struct.Name,
		MethodName:     "GetType",
		Body:           fmt.Sprintf(`return "%s"`, event.StructName()),
		Return: []ReturnType{
			{Type: "string"},
		},
	}))

}
