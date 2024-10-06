package main

import (
	"fmt"
	"go/ast"
	"log"
	"path"
	"time"
)

var SourcingPath = "/internal/sourcing"
var AggregatesPath = path.Join(SourcingPath, "aggregates")
var EventsPath = path.Join(SourcingPath, "events")
var ProjectionsPath = path.Join(SourcingPath, "projections")
var CommandsPath = path.Join(SourcingPath, "commands")

func main() {
	now := time.Now()
	parsed := NewConfigParser("config.yml")

	WriteFile(path.Join(CommandsPath, "generated.go"), func(content *ast.File) string {
		return GenerateCommandsFile(parsed)
	})

	WriteFile(path.Join(AggregatesPath, "generated.go"), func(content *ast.File) string {
		return GenerateAggregatesHandleFile(parsed)
	})

	WriteFile(path.Join(EventsPath, "generated.go"), func(content *ast.File) string {
		return GenerateEventsFile(parsed)
	})

	WriteFile(path.Join(ProjectionsPath, "generated.go"), func(content *ast.File) string {
		return GenerateProjectionsHandlerInterface(parsed)
	})

	WriteProjectionHandlers(parsed)
	WriteAggregateFile(parsed)

	log.Println(fmt.Sprintf("code generation successful. took %s", time.Since(now)))
}
