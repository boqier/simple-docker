package main

import (
	"dockertest1/cgroups"
	"dockertest1/cgroups/subsystem"
	"dockertest1/container"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Run(tty bool, comArray []string, res *subsystem.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Start parent process error: %v", err)
	}
	cgroupmanager := cgroups.NewCgroupManager("my-docker")
	defer cgroupmanager.Remove()
	cgroupmanager.Set(res)
	cgroupmanager.Apply(os.Getpid())
	sendInitCommand(comArray, writePipe)
	if tty {
	_ = parent.Wait()
	mntURL := "/root/mnt"
	rootURL := "/root"
	container.DeleteWorkSpace(mntURL, rootURL, volume)
	}
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is: %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
