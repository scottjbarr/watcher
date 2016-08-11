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

func printCommand(cmd *exec.Cmd) {
	log.Printf("%s", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("%s\n", string(outs))
	}
}

func main() {
	var watcher *fsnotify.Watcher
	var err error

	if watcher, err = fsnotify.NewWatcher(); err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		panic("No directory supplied")
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

					cmdOutput := &bytes.Buffer{}

					// Attach buffer to command
					cmd.Stdout = cmdOutput

					// Execute command
					printCommand(cmd)
					err := cmd.Run()
					printError(err)
					printOutput(cmdOutput.Bytes())

					log.Printf("Watching...")
				}
			}
		}
	}()

	<-doneChan
}

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
