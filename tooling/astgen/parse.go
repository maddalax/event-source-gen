package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sort"
	"strings"
)

var CommandImports []string
var EventImports []string
var Config Configuration

func SortImports(imports *[]string) []string {
	sort.Strings(*imports)
	return *imports
}

type Field struct {
	Name        string
	Type        string
	IsIdField   bool
	Validations []Validation
}

func (f *Field) ToNameType() *NameType {
	return &NameType{
		Name: f.Name,
		Type: f.Type,
	}
}

type TypeMapping struct {
	field   string
	_type   string
	_import string
}

func (t *TypeMapping) ApplyMapping(field *Field, arr *[]string) {
	matches := field.Type == t.field

	if strings.HasSuffix(t.field, "*") {
		matches = strings.Contains(field.Type, strings.TrimSuffix(t.field, "*"))
	}

	if matches {
		if t._type != "" {
			field.Type = t._type
		}
		if t._import != "" {
			has := false
			for _, i := range *arr {
				if i == t._import {
					has = true
				}
			}
			if !has {
				*arr = append(*arr, t._import)
			}
		}
	}
}

type Raw struct {
	Events                      map[string]map[string]map[string]string `yaml:"events"`
	Aggregates                  map[string]map[string]Aggregate         `yaml:"aggregates"`
	Projections                 map[string][]string                     `yaml:"projections"`
	TypeMapping                 map[string]map[string]string            `yaml:"type_mapping"`
	EventMigrations             map[string][]string                     `yaml:"event_migrations"`
	ProjectionAdditionalMethods map[string]map[string]map[string]string `yaml:"projection_additional_methods"`
	Config                      Configuration                           `yaml:"config"`
}

type Aggregate struct {
	Fields map[string]string `yaml:"fields"`
	Events []string          `yaml:"events"`
}

type Projection struct {
	events []*Event
	name   string
}

func NewConfigParser(path string) *Parsed {
	raw := NewRaw()
	raw.Read(path)
	Config = raw.Config
	parsed := raw.ToParsed()
	return parsed
}

type Parsed struct {
	Commands                    []Command
	Events                      []Event
	Projections                 []Projection
	TypeMappings                []TypeMapping
	EventMigrations             map[string][]string
	ProjectionAdditionalMethods map[string]map[string]map[string]string
}

type Configuration struct {
	SourcingDir    string `yaml:"sourcing_dir"`
	ValidationsDir string `yaml:"validations_dir"`
}

func NewRaw() *Raw {
	return &Raw{
		Aggregates:  make(map[string]map[string]Aggregate),
		Projections: make(map[string][]string),
		Events:      make(map[string]map[string]map[string]string),
		Config:      Configuration{},
	}
}

func (t *Raw) Read(config string) {
	data, err := os.ReadFile(config)
	if err != nil {
		PanicF("error: %v", err)
	}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		PanicF("error: %v", err)
	}
}

func (t *Raw) ToParsed() *Parsed {
	parsed := Parsed{
		EventMigrations:             t.EventMigrations,
		ProjectionAdditionalMethods: t.ProjectionAdditionalMethods,
	}

	events := make(map[string]Event)

	for field, value := range t.TypeMapping {
		parsed.TypeMappings = append(parsed.TypeMappings, TypeMapping{
			field:   field,
			_type:   value["type"],
			_import: value["import"],
		})
	}

	for s, event := range t.Events {
		entity := s
		for name, fields := range event {
			e := Event{
				entity:   entity,
				name:     name,
				mappings: parsed.TypeMappings,
				fields:   make([]Field, 0),
			}
			for n, f := range fields {
				field := e.ParseField(n, f)
				e.fields = append(e.fields, field)
			}
			e.Validate()
			key := fmt.Sprintf("%s.%s", entity, name)
			events[key] = e
			parsed.Events = append(parsed.Events, e)
		}
	}

	for s, root := range t.Aggregates {
		entity := s
		for name, aggregate := range root {
			command := Command{
				entity:   entity,
				mappings: parsed.TypeMappings,
			}
			command.name = name
			command.fields = make([]Field, 0)
			for n, f := range aggregate.Fields {
				field := command.ParseField(n, f)
				command.fields = append(command.fields, field)
			}
			for _, event := range aggregate.Events {
				e, ok := events[event]
				if ok {
					command.events = append(command.events, &e)
				} else {
					PanicF("could not find event %s for command %s", event, name)
				}
			}
			command.Validate(&aggregate)
			parsed.Commands = append(parsed.Commands, command)
		}
	}

	for proj, s := range t.Projections {
		for _, name := range s {
			projection := Projection{
				name:   proj,
				events: make([]*Event, 0),
			}
			hasMatch := false
			for _, event := range parsed.Events {
				if event.Compare(name) {
					projection.events = append(projection.events, &event)
					parsed.Projections = append(parsed.Projections, projection)
					hasMatch = true
				}
			}
			if !hasMatch {
				PanicF("could not find event %s for projection %s", name, proj)
			}
		}
	}

	sort.Slice(parsed.Commands, func(i, j int) bool {
		return parsed.Commands[i].StructName() < parsed.Commands[j].StructName()
	})

	sort.Slice(parsed.Events, func(i, j int) bool {
		return parsed.Events[i].StructName() < parsed.Events[j].StructName()
	})

	return &parsed
}
