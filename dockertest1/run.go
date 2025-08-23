package main

import (
	"dockertest1/cgroups"
	"dockertest1/cgroups/subsystem"
	"dockertest1/container"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Run(tty bool, comArray []string, res *subsystem.ResourceConfig, volume string, containerName string, imageName string) {
	containerId := container.GenerateContainerID()
	parent, writePipe := container.NewParentProcess(tty, volume, containerId, imageName)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Start parent process error: %v", err)
	}
	err := container.RecordContainerInfo(parent.Process.Pid, comArray, containerName, containerId, volume)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}
	cgroupmanager := cgroups.NewCgroupManager("my-docker")

	cgroupmanager.Set(res)
	cgroupmanager.Apply(os.Getpid())
	sendInitCommand(comArray, writePipe)
	if tty {
		_ = parent.Wait()
		container.DeleteWorkSpace(containerId, volume)
		container.DeleteContainerInfo(containerId)
	}
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is: %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()

}
