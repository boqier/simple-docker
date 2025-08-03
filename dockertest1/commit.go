package main
import (
	log "github.com/sirupsen/logrus"
	"fmt"
	"os/exec"
)
func commitContainer(imageName string){
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("commit image %s to %s\n", imageName, imageTar)
	if _, err := exec.Command("tar", "-cvf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("commit image %s failed: %v", imageName, err)
		return
	}
}


