package main

import (
	"dockertest1/container"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
)

func ListContainers() {
	files, err := os.ReadDir(container.InfoLoc)
	if err != nil {
		log.Fatalf("Read dir %s error %v", container.InfoLoc, err)
	}
	containers := make([]*container.Info, 0, len(files))
	for _, file := range files {
		tmpContaner, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContaner)
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		log.Errorf("Fprint error %v", err)
	}
	for _, item := range containers {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			log.Errorf("Fprint error %v", err)
		}
	}
	if err = w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
	}
}
func getContainerInfo(file os.DirEntry) (*container.Info, error) {
	configFileDir := fmt.Sprintf(container.InfoLocFormat, file.Name())
	configFilePath := path.Join(configFileDir, container.ConfigName)
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	info := new(container.Info)
	if err := json.Unmarshal(content, info); err != nil {
		log.Errorf("Unmarshal json error %v", err)
		return nil, err
	}
	return info, nil
}
