package main

import (
	"fmt"
	"sort"
)

func GenerateProjectionsHandlerInterface(parsed *Parsed) string {
	builder := NewCodeBuilder(nil)
	builder.AppendLine("package projections")

	builder.AddImport(GetSourcingPath(""))
	builder.AddImport(GetSourcingPath("events"))

	funcs := make([]Function, 0)

	for _, event := range parsed.Events {
		name := fmt.Sprintf("On%s", event.StructName())
		funcs = append(funcs, Function{
			Name: name,
			Parameters: []NameType{
				{Name: "e", Type: fmt.Sprintf("*events.%s", event.StructName())},
			},
		})
	}

	additional := parsed.ProjectionAdditionalMethods["methods"]

	type Pair struct {
		Method       string
		ParamString  string
		ReturnString string
	}

	sorted := make([]Pair, 0)

	for method, values := range additional {
		sorted = append(sorted, Pair{
			Method:       method,
			ParamString:  values["params"],
			ReturnString: values["return"],
		})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Method < sorted[j].Method
	})

	for _, pair := range sorted {
		funcs = append(funcs, Function{
			Name: pair.Method,
			Parameters: []NameType{
				{Name: "", Type: pair.ParamString},
			},
			Return: []ReturnType{
				{Type: pair.ReturnString},
			},
		})
	}

	i := builder.BuildInterface("ProjectionHandler", funcs)
	builder.Append(i)

	builder.BuildEventHandler(parsed.Events)
	builder.BuildGetProjections(Map(parsed.Projections, func(t Projection) string {
		return FormatFieldName(t.name)
	}))

	return builder.String()
}
