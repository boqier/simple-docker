package container

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewWorkSpace(rootURL string, mntURL string, volume string) {

	CreateReadOnlyLayer(rootURL)
	CreateReadWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		length := len(volumeUrls)
		if length == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			MountVolume(rootURL, mntURL, volumeUrls)
			log.Info("Volume mount success")
		} else {
			log.Errorf("volume format error, should be -v /host/path:/container/path")
		}
	}
}
func volumeUrlExtract(volume string) []string {
	var volumeUrls []string
	volumeUrls = strings.Split(volume, ":")
	return volumeUrls
}
func MountVolume(rootURL string, mntURL string, volumeUrls []string) {
	parentURL := volumeUrls[0]
	if err := os.MkdirAll(parentURL, 0777); err != nil {
		log.Errorf("create mount path %s error: %v", parentURL, err)
	}
	containerURL := volumeUrls[1]
	containerVolimeURL := rootURL + containerURL
	if err := os.MkdirAll(containerVolimeURL, 0777); err != nil {
		log.Errorf("create mount path %s error: %v", mntURL+containerURL, err)
	}
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolimeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume %s error: %v", containerVolimeURL, err)
		return
	}
	log.Infof("mount volume %s success", containerVolimeURL)
}
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "/busybox"
	busyboxTarURl := rootURL + "/busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("check path %s error: %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("create busybox path %s error: %v", busyboxURL, err)
			return
		}
		if _, err = exec.Command("tar", "-xvf", busyboxTarURl, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("extract busybox tar %s error: %v", busyboxTarURl, err)
			return
		}
	}
}
func CreateReadWriteLayer(rootURL string) {
	readWriteURL := rootURL + "/readwrite"
	if err := os.Mkdir(readWriteURL, 0777); err != nil {
		log.Errorf("create readwrite path %s error: %v", readWriteURL, err)
		return
	}
}
func CreateMountPoint(rootURL string, mntURL string) {
	mountURL := rootURL + "/mnt"
	if err := os.Mkdir(mountURL, 0777); err != nil {
		log.Errorf("create mount path %s error: %v", mountURL, err)
		return
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount aufs error: %v", err)
		return
	}
	log.Infof("mount aufs success")
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		vlumneUrls := volumeUrlExtract(volume)
		if len(vlumneUrls) == 2 && vlumneUrls[0] != "" && vlumneUrls[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, vlumneUrls)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
		DeleteWriteLayer(rootURL)
	}
}
func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount %s error: %v", mntURL, err)
		return
	}
	log.Infof("umount %s success", mntURL)
}
func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeUrls []string) {
	containerUrl := mntURL + volumeUrls[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount %s error: %v", mntURL, err)
		return
	}
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount %s error: %v", mntURL, err)
		return
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("remove %s error: %v", mntURL, err)
		return
	}
	log.Infof("umount %s success", mntURL)
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLater/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove write layer %s error: %v", writeURL, err)
		return
	}
	log.Infof("remove write layer %s success", writeURL)
}
