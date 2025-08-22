package main

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func commitContainer(imageName string) {
	mntPath := "/root/merged/"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		log.Errorf("Commit container error %v", err)
		return
	}
}
