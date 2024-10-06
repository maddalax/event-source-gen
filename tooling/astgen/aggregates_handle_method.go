package main

import (
	"fmt"
	"strings"
)

func GenerateAggregatesHandleFile(parsed *Parsed) string {
	builder := NewCodeBuilder(nil)
	builder.AppendLine("package aggregates")

	builder.AddImport(GetSourcingPath("commands"))
	builder.AddImport(GetSourcingPath("events"))
	builder.AddImport(GetSourcingPath(""))
	builder.AddImport("errors")

	builder.AppendLine(buildHandleCommand(builder, parsed))
	builder.AppendLine(buildLoadEvent(builder, parsed))

	return builder.String()
}

func buildHandleCommand(builder *CodeBuilder, parsed *Parsed) string {
	cases := ""

	for _, command := range Unique(parsed.Commands, func(item Command) string {
		return fmt.Sprintf("%s.%s", item.entity, item.StructName())
	}) {
		cases += buildCase(&command)
	}

	cases += `
		default:
			return errors.New("unknown command type")
	`

	body := fmt.Sprintf(`switch c := command.(type) {
		%s
	}`, cases)

	method := Function{
		Name: "Handle",
		Parameters: []NameType{
			{Name: "command", Type: "sourcing.Command"},
		},
		Return: []ReturnType{
			{Type: "error"},
		},
		Body: body,
	}

	return builder.BuildFunction(method)
}

func buildLoadEvent(builder *CodeBuilder, parsed *Parsed) string {
	eventMap := NewOrderedMap[string, []*Event]()

	for _, event := range parsed.Commands {
		key := PascalCase(event.entity)
		exists, ok := eventMap.Get(key)
		if !ok {
			exists = make([]*Event, 0)
			eventMap.Set(key, exists)
		}
		exists = append(exists, event.events...)
		eventMap.Set(key, exists)
	}

	cases := ""

	for _, entry := range eventMap.Entries() {
		cases += buildEventCase(entry.Key, entry.Value)
		cases += "\n"
	}

	body := fmt.Sprintf(`
		switch a := agg.(type) {
			%s
			default:
				return nil
		}
	`, cases)

	f := Function{
		Name: "LoadEvent",
		Parameters: []NameType{
			{Name: "agg", Type: "any"},
			{Name: "event", Type: "*sourcing.BaseEvent[any]"},
		},
		Return: []ReturnType{
			{Type: "error"},
		},
		Body: body,
	}

	return builder.BuildFunction(f)
}

func buildEventCase(structName string, events []*Event) string {
	builder := NewCodeBuilder(nil)
	builder.AppendLineF("case *%s:", structName)
	builder.AppendLineF("switch e := event.Data.(type) {")
	for _, event := range Unique(events, func(item *Event) string {
		return fmt.Sprintf("%s:%s", item.entity, item.StructName())
	}) {
		eventHandler := fmt.Sprintf("On%s", event.StructName())
		builder.AppendLineF(`
		case *events.%s:
			a.Version = event.Version
			return a.%s(e)
		`, event.StructName(), eventHandler)
	}
	builder.AppendLine(`
		default:
			return nil
	`)
	builder.AppendLine("}")
	return builder.String()
}

func buildCase(command *Command) string {
	aggType := PascalCase(command.entity)
	builder := NewCodeBuilder(nil)

	builder.AppendLineF(`
	case *commands.%s:
		agg := %s{}
		LoadEvents(command.GetId(), &agg)
		agg.Version = agg.Version + 1
	`, command.StructName(), aggType)

	fields := make([]string, 0)
	for i := range command.events {
		fields = append(fields, fmt.Sprintf("e%d", i))
	}
	fields = append(fields, "err")
	builder.AppendLineF("%s := agg.Handle%s(c)", strings.Join(fields, ", "), command.StructName())
	builder.AppendLineF("if err != nil { return err }")

	for i, field := range fields {
		if i == len(fields)-1 {
			continue
		}
		builder.AppendLineF(`
		if %s == nil {
			return errors.New("command.%s -> command.Handle -> events.%s is nil. ensure it is returned from the aggregate")
		}
		`, field, command.StructName(), command.events[i].StructName())
	}

	for i, field := range fields {
		if i == len(fields)-1 {
			continue
		}
		builder.AppendLineF(`
			agg.AppendAndCommit(%s)
		`, field)
	}

	builder.AppendLineF("return nil")

	return builder.String()
}
