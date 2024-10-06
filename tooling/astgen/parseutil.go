package main

import (
	"fmt"
	"strings"
)

func ParseValidation(reciever string, fieldName string, validation string) *Validation {
	fieldName = FormatFieldName(fieldName)
	fieldNameQuotes := fmt.Sprintf(`"%s"`, fieldName)
	value := fmt.Sprintf("%s.%s", reciever, fieldName)

	if validation == "required" {
		return &Validation{
			MethodCall: fmt.Sprintf("validators.IsEmpty(%s, %s)", fieldNameQuotes, value),
		}
	}

	if strings.HasPrefix(validation, "prefix") {
		start := strings.Index(validation, "'") + 1
		end := strings.LastIndex(validation, "'")
		prefix := ""

		if start >= 0 && end >= 0 && start < end {
			extracted := validation[start:end]
			prefix = extracted
		}

		return &Validation{
			MethodCall: fmt.Sprintf("validators.HasPrefix(%s, %s, %s)", fieldNameQuotes, value, prefix),
		}
	}

	return nil
}

func ParseField(importArr *[]string, mappings []TypeMapping, reciever string, name string, value string) Field {
	split := strings.Split(value, " ")
	fieldType := split[0]
	isIdField := fieldType == "id"
	if isIdField {
		fieldType = "string"
	}

	validations := make([]Validation, 0)

	if len(split) > 1 {
		validationStr := strings.TrimLeft(split[1], "[")
		validationStr = strings.TrimRight(validationStr, "]")
		valSplit := strings.Split(validationStr, ",")
		for _, s := range valSplit {
			trim := strings.TrimSpace(s)
			if trim == "" {
				continue
			}
			result := ParseValidation(reciever, name, trim)
			if result != nil {
				*importArr = append(*importArr, Config.ValidationsDir)
				validations = append(validations, *result)
			}
		}
	}

	f := &Field{
		Name:        FormatFieldName(name),
		IsIdField:   isIdField,
		Type:        fieldType,
		Validations: validations,
	}

	for _, mapping := range mappings {
		mapping.ApplyMapping(f, importArr)
	}

	return *f
}

func FormatFieldName(name string) string {
	split := strings.Split(name, "_")
	if strings.Contains(name, "-") {
		split = strings.Split(name, "-")
	}
	parts := make([]string, 0)
	for _, s := range split {
		parts = append(parts, PascalCase(s))
	}
	return strings.Join(parts, "")
}

func PascalCase(s string) string {
	if s == "" {
		return s
	}
	// Convert the first rune (character) to uppercase and concatenate with the rest of the string
	return strings.ToUpper(string(s[0])) + s[1:]
}
