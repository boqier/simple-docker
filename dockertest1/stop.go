package main

import (
	"dockertest1/container"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func stopContainer(containerID string) {
	containerInfo, err := GetContainerInfobyId(containerID)
	if err != nil {
		log.Errorf("Get container info error %v", err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("Convert pid to int error %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGKILL); err != nil {
		log.Errorf("Kill container process error %v", err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	newContainerInfo, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Marshal container info error %v", err)
		return
	}
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerID)
	configFilePath := path.Join(dirPath, container.ConfigName)
	if err := os.WriteFile(configFilePath, newContainerInfo, 0644); err != nil {
		log.Errorf("Write container info error %v", err)
		return
	}
}

func GetContainerInfobyId(containerId string) (*container.Info, error) {
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var containerInfo container.Info
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}
