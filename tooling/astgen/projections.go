package main

import (
	"fmt"
	"go/ast"
	"path"
)

func WriteProjectionHandlers(parsed *Parsed) {
	for _, projection := range parsed.Projections {
		fileName := fmt.Sprintf("%s.go", projection.name)
		WriteFile(path.Join(ProjectionsPath, fileName), func(content *ast.File) string {
			return generateProjectionHandlerFile(parsed, &projection, content)
		})
	}

	WriteFile(path.Join(ProjectionsPath, "base.go"), func(content *ast.File) string {
		return updateBaseProjection(parsed, content)
	})
}

func updateBaseProjection(parsed *Parsed, existing *ast.File) string {
	astu := NewAstUtil(existing)
	builder := NewCodeBuilder(astu)

	builder.astu.SetPackageName("projections")
	builder.astu.AddImport(GetSourcingPath(""))
	builder.astu.AddImport(GetSourcingPath("events"))
	builder.SyncFromAst()

	for _, event := range parsed.Events {
		methodName := fmt.Sprintf("On%s", event.StructName())
		if !builder.astu.HasMethod(methodName) {
			builder.AppendLineF(`
			func (p *BaseProjection) On%s(e *events.%s) {
				// Override this method in your projection
			}`, event.StructName(), event.StructName())
			builder.SyncToAst()
		} else {
			builder.astu.UpdateMethodSignature(UpdateSignature{
				MethodName: methodName,
				NewName:    methodName,
				NewParams: []NameType{
					{Name: "e", Type: fmt.Sprintf("*events.%s", event.StructName())},
				},
				NewResults: []ReturnType{},
			})
			builder.SyncFromAst()
		}
	}
	return builder.String()
}

func generateProjectionHandlerFile(parsed *Parsed, projection *Projection, existing *ast.File) string {
	astu := NewAstUtil(existing)
	builder := NewCodeBuilder(astu)

	builder.astu.SetPackageName("projections")
	builder.astu.AddImport(GetSourcingPath("events"))
	builder.SyncFromAst()

	structName := FormatFieldName(projection.name)

	if !builder.astu.HasStruct(structName) {
		builder.AppendLineF(`
		type %s struct {
			BaseProjection
		}`, structName)
		builder.SyncToAst()
	}

	for _, event := range projection.events {
		e := addProjectionHandler(FormatFieldName(projection.name), event, builder.astu)
		builder.SyncFromAst()
		builder.AppendLine(e)
		builder.SyncToAst()
	}

	return builder.String()
}

func addProjectionHandler(structName string, event *Event, astu *AstUtil) string {
	builder := NewCodeBuilder(astu)

	methodName := fmt.Sprintf("On%s", event.StructName())

	m := Method{
		StructReceiver: "a",
		StructName:     structName,
		MethodName:     methodName,
		Parameters: []NameType{
			{Name: "e", Type: fmt.Sprintf("*events.%s", event.StructName())},
		},
		Return: []ReturnType{},
	}

	if astu.HasMethod(methodName) {
		astu.UpdateMethodSignature(UpdateSignature{
			MethodName: methodName,
			NewName:    methodName,
			NewParams:  m.Parameters,
			NewResults: m.Return,
		})
		newMethod := astu.GetMethodAsString(methodName)
		astu.DeleteMethod(methodName)
		return newMethod
	} else {
		return builder.BuildMethod(m)
	}
}
