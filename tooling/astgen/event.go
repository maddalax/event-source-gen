package main

import (
	"fmt"
	"sort"
)

type Event struct {
	entity   string
	name     string
	fields   []Field
	mappings []TypeMapping
}

func (e *Event) GetFields() []Field {
	sort.Slice(e.fields, func(i, j int) bool {
		return e.fields[i].Name < e.fields[j].Name
	})
	return e.fields
}

func (e *Event) GetFields2() []NameType {
	return Map(e.GetFields(), func(f Field) NameType {
		return NameType{Name: f.Name, Type: f.Type}
	})
}

func (e *Event) Compare(value string) bool {
	return fmt.Sprintf("%s.%s", e.entity, e.name) == value
}

func (e *Event) GetIdField() *Field {
	for _, f := range e.GetFields() {
		if f.IsIdField {
			return &f
		}
	}
	return nil
}

func (e *Event) ReceiverName() string {
	return "e"
}

func (e *Event) StructName() string {
	return FormatFieldName(fmt.Sprintf("%s%s", PascalCase(e.entity), PascalCase(e.name)))
}

func (e *Event) ParseField(name string, value string) Field {
	return ParseField(&EventImports, e.mappings, e.ReceiverName(), name, value)
}

func (e *Event) Validate() {
	idField := ""
	for _, f := range e.GetFields() {
		if f.IsIdField {
			idField = f.Name
		}
	}

	if idField == "" {
		PanicF("event %s.%s is missing an id field", e.entity, e.name)
	}
}
