package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

type Command struct {
	entity   string
	name     string
	fields   []Field
	events   []*Event
	mappings []TypeMapping
}

func (c *Command) GetFields() []Field {
	sort.Slice(c.fields, func(i, j int) bool {
		return c.fields[i].Name < c.fields[j].Name
	})
	return c.fields
}

func (c *Command) GetFields2() []NameType {
	return Map(c.GetFields(), func(f Field) NameType {
		return NameType{Name: f.Name, Type: f.Type}
	})
}

func (c *Command) Validate(raw *Aggregate) {
	idField := ""
	for _, f := range c.GetFields() {
		if f.IsIdField {
			idField = f.Name
		}
	}

	if idField == "" {
		PanicF("command %s.%s is missing an id field", c.entity, c.name)
	}

	if len(c.events) == 0 {
		PanicF("command %s.%s is missing events", c.entity, c.name)
	}
}

func (c *Command) BuildStruct() *ast.GenDecl {
	// Create the struct name by combining `name` and `entity`
	structName := ast.NewIdent(c.StructName())

	// Create a list of fields
	var fields []*ast.Field
	for _, field := range c.GetFields() {
		field := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(PascalCase(field.Name))},
			Type:  ast.NewIdent(field.Type),
		}
		fields = append(fields, field)
	}

	// Define the struct type
	structType := &ast.StructType{
		Fields: &ast.FieldList{
			List: fields,
		},
	}

	// Create a new type spec for the struct
	typeSpec := &ast.TypeSpec{
		Name: structName,
		Type: structType,
	}

	// Create a new declaration for the struct type
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			typeSpec,
		},
	}
}

func (c *Command) GetIdField() *Field {
	for _, f := range c.GetFields() {
		if f.IsIdField {
			return &f
		}
	}
	return nil
}

func (c *Command) ReceiverName() string {
	return "c"
}

func (c *Command) StructName() string {
	return FormatFieldName(fmt.Sprintf("%s%s", PascalCase(c.name), PascalCase(c.entity)))
}

func (c *Command) ParseField(name string, value string) Field {
	return ParseField(&CommandImports, c.mappings, c.ReceiverName(), name, value)
}
