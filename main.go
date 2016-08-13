package main

import (
	"bytes"
	"flag"
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

	var dir = flag.String("d", ".", "directory to watch")
	var command = flag.String("c", "go test", "command to execute on change")

	flag.Parse()

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		panic(err)
	}

	log.Printf("Adding directory : %v", *dir)

	if err := watcher.Add(*dir); err != nil {
		panic(err)
	}

	parts := strings.Fields(*command)
	doneChan := make(chan bool)

	log.Printf("Command : %v", *command)
	log.Printf("Watching for changes...")

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if isMatch(event) {
					cmd := exec.Command(parts[0], parts[1:]...)

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

					log.Printf("Watching for changes...")
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
