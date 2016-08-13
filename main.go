package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-fsnotify/fsnotify"
)

func main() {
	var watcher *fsnotify.Watcher
	var err error

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		message := fmt.Sprintf("Usage : %v directory\n", os.Args[0])
		os.Stderr.WriteString(message)
		os.Exit(1)
	}

	for _, f := range os.Args[1:] {
		log.Printf("Watching directory : %v", f)
		if err := watcher.Add(f); err != nil {
			panic(err)
		}
	}

	doneChan := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if isMatch(event) {
					cmd := exec.Command("go", "test")

					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}

					// write stdout to buffer
					cmd.Stdout = stdout
					cmd.Stderr = stderr

					// execute the command
					log.Printf("%s", strings.Join(cmd.Args, " "))
					err := cmd.Run()

					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("%s\n", err.Error()))
					}

					// write output
					writeOutput(stdout)
					writeOutput(stderr)

					log.Printf("Watching...")
				}
			}
		}
	}()

	<-doneChan
}

// isMatch returns true if a Go test file is created or changed, false
// otherwise.
func isMatch(event fsnotify.Event) bool {
	op := event.Op
	n := event.Name

	if !(op&fsnotify.Create == fsnotify.Create ||
		op&fsnotify.Write == fsnotify.Write) {

		return false
	}

	if !strings.HasPrefix(n, ".") && strings.HasSuffix(n, ".go") {
		return true
	}

	return false
}

func writeOutput(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	if len(buf.Bytes()) > 0 {
		fmt.Printf("%s\n", string(buf.Bytes()))
	}
}
