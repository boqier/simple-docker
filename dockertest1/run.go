package main

import (
	"dockertest1/container"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Run(tty bool, comArray []string) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Start parent process error: %v", err)
	}
	sendInitCommand(comArray, writePipe)
	_ = parent.Wait()

}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is: %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
