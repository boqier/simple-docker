package main

import (
	"dockertest1/container"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func removeContainer(containerID string, force bool) {
	containerInfo, err := GetContainerInfobyId(containerID)
	if err != nil {
		log.Errorf("Get container info error %v", err)
		return
	}
	switch containerInfo.Status {
	case container.STOP:
		dirPath := fmt.Sprintf(container.InfoLocFormat, containerID)
		if err := os.RemoveAll(dirPath); err != nil {
			log.Errorf("Remove container dir %s error %v", dirPath, err)
			return
		}
		container.DeleteWorkSpace(containerID, containerInfo.Volume)
	case container.RUNNING:
		if !force {
			log.Errorf("Container %s is still running, use --force to remove it", containerID)
			return
		}
		stopContainer(containerID)
		removeContainer(containerID, force)
	default:
		log.Infof("Container %s stopped and removed successfully", containerID)

	}
}
