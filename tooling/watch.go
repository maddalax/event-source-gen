package main

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
)

func main() {
	once := false
	if len(os.Args) > 1 {
		once = os.Args[1] == "--once"
	}

	command := ""
	for i, arg := range os.Args {
		if arg == "--command" {
			command = os.Args[i+1]
		}
	}

	if command == "" {
		panic("command is required")
	}

	if once {
		runCommand(command)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from fatal error:", r)
			// You can log the error here or take other corrective actions
		}
	}()

	runCommand(command)
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					success := runCommand(command)
					if success {
						log.Println(fmt.Sprintf("config.yml changed. code generation successful"))
					} else {
						log.Println(fmt.Sprintf("config.yml changed. code generation failed"))
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add("./config.yml")
	if err != nil {
		panic(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func runCommand(command string) bool {
	// Create a new command
	cmd := exec.Command("bash", "-c", command)

	// Capture stdout and stderr in buffers
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Println(fmt.Sprintf("error: %v", err))
		println(stderr.String())
		return false
	} else {
		println(out.String())
		return true
	}
}
