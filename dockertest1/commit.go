package main

import (
	"github.com/pkg/errors"
	"os/exec"
	"dockertest1/utils"
	log "github.com/sirupsen/logrus"
)
var ErrimageAleadyExists=errors.New("Image already exists")
func commitContainer(containerId string ,imageName string) error{
	mntPath:=utils.GetMerged(containerId)
	imagePath:=utils.GetImage(imageName)
	exists,err:=utils.PathExists(imagePath)
	if err!=nil{
	return errors.WithMessagef(err,"Check image path %s error",imagePath)
	}
	if exists{
		return ErrimageAleadyExists
	}
	log.Infof("Create image path %s",imagePath)
	if _,err:=exec.Command("tar","-czf",imagePath,"-C",mntPath,".").CombinedOutput();err!=nil{
		return errors.WithMessagef(err,"Tar image %s error",imageName)
	}
	return nil
}
