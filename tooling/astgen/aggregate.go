package main

import (
	"fmt"
	"go/ast"
	"path"
	"strings"
)

func WriteAggregateFile(parsed *Parsed) {
	m := NewOrderedMap[string, []Command]()

	for _, command := range parsed.Commands {
		get, has := m.Get(command.entity)
		if !has {
			get = make([]Command, 0)
			m.Set(command.entity, get)
		}
		get = append(get, command)
		m.Set(command.entity, get)
	}

	for _, entry := range m.Entries() {
		key := entry.Key
		commands := entry.Value
		fileName := fmt.Sprintf("%s.go", key)
		fullPath := path.Join(AggregatesPath, fileName)
		WriteFile(fullPath, func(content *ast.File) string {
			result := GenerateAggregateMethod(PascalCase(key), commands, parsed, content)
			return result
		})
	}
}

func GenerateAggregateMethod(structName string, commands []Command, parsed *Parsed, content *ast.File) string {
	astu := NewAstUtil(content)
	builder := NewCodeBuilder(astu)

	builder.astu.SetPackageName("aggregates")
	builder.astu.AddImport(GetSourcingPath("events"))
	builder.astu.AddImport(GetSourcingPath("commands"))
	builder.astu.AddImport(GetSourcingPath(""))
	builder.SyncFromAst()

	if !builder.astu.HasStruct(structName) {
		builder.AppendLineF(`
		type %s struct {
			sourcing.AggregateBase
	    }`, structName)
		builder.SyncToAst()
	}

	for _, command := range commands {
		c := addCommand(structName, &command, builder.astu)
		builder.SyncFromAst()
		builder.AppendLine(c)
		builder.SyncToAst()

		for _, event := range command.events {
			e := addEventHandler(structName, event, builder.astu)
			builder.SyncFromAst()
			builder.AppendLine(e)
			builder.SyncToAst()
		}

	}

	return builder.String()
}

func addCommand(structName string, command *Command, astu *AstUtil) string {
	builder := NewCodeBuilder(astu)

	returns := make([]ReturnType, 0)

	for _, ret := range command.events {
		returns = append(returns, ReturnType{
			fmt.Sprintf("*events.%s", ret.StructName()),
		})
	}

	returns = append(returns, ReturnType{
		"error",
	})

	methodName := fmt.Sprintf("Handle%s", command.StructName())

	m := Method{
		StructReceiver: "a",
		StructName:     structName,
		MethodName:     methodName,
		Parameters: []NameType{
			{Name: "c", Type: fmt.Sprintf("*commands.%s", command.StructName())},
		},
		Body: fmt.Sprintf("return %s", strings.Join(Map(returns, func(rt ReturnType) string {
			return "nil"
		}), ", ")),
		Return: returns,
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

func addEventHandler(structName string, event *Event, astu *AstUtil) string {
	builder := NewCodeBuilder(astu)

	returns := make([]ReturnType, 0)

	returns = append(returns, ReturnType{
		"error",
	})

	methodName := fmt.Sprintf("On%s", event.StructName())

	m := Method{
		StructReceiver: "a",
		StructName:     structName,
		MethodName:     methodName,
		Parameters: []NameType{
			{Name: "e", Type: fmt.Sprintf("*events.%s", event.StructName())},
		},
		Body:   fmt.Sprintf("return nil"),
		Return: returns,
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
