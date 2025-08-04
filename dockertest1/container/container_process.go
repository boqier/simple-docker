package container

import (
	"fmt"
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
	containerVolumeURL := mntURL + containerURL
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Errorf("create mount path %s error: %v", containerVolumeURL, err)
	}

	// 使用bind mount挂载volume
	cmd := exec.Command("mount", "--bind", parentURL, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume %s error: %v", containerVolumeURL, err)
		return
	}
	log.Infof("mount volume %s success", containerVolumeURL)
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "/busybox"
	busyboxTarURL := rootURL + "/busybox.tar"

	// 检查busybox.tar是否存在
	if _, err := os.Stat(busyboxTarURL); os.IsNotExist(err) {
		log.Errorf("busybox.tar not found at %s", busyboxTarURL)
		// 尝试从当前目录查找
		if _, err := os.Stat("./busybox.tar"); err == nil {
			busyboxTarURL = "./busybox.tar"
			log.Infof("found busybox.tar in current directory")
		} else {
			log.Errorf("busybox.tar not found, please ensure it exists")
			return
		}
	}

	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("check path %s error: %v", busyboxURL, err)
	}
	if !exist {
		if err := os.MkdirAll(busyboxURL, 0777); err != nil {
			log.Errorf("create busybox path %s error: %v", busyboxURL, err)
			return
		}
		if output, err := exec.Command("tar", "-xf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("extract busybox tar %s error: %v, output: %s", busyboxTarURL, err, string(output))
			return
		}
		log.Infof("extract busybox tar success")
	}
}

func CreateReadWriteLayer(rootURL string) {
	writeLayerURL := rootURL + "/writeLayer"
	if err := os.MkdirAll(writeLayerURL, 0777); err != nil {
		log.Errorf("create writeLayer path %s error: %v", writeLayerURL, err)
		return
	}

	// 创建work目录，overlayfs需要
	workURL := rootURL + "/work"
	if err := os.MkdirAll(workURL, 0777); err != nil {
		log.Errorf("create work path %s error: %v", workURL, err)
		return
	}
}

func CreateMountPoint(rootURL string, mntURL string) {
	// 先清理可能存在的挂载点
	if isMounted(mntURL) {
		log.Infof("unmounting existing mount at %s", mntURL)
		exec.Command("umount", "-l", mntURL).Run()
	}

	// 删除并重新创建挂载点目录
	os.RemoveAll(mntURL)
	if err := os.MkdirAll(mntURL, 0755); err != nil {
		log.Errorf("create mount point %s error: %v", mntURL, err)
		return
	}

	// overlayfs目录路径
	lowerdir := rootURL + "/busybox"
	upperdir := rootURL + "/writeLayer"
	workdir := rootURL + "/work"

	log.Infof("Overlay directories: lower=%s, upper=%s, work=%s", lowerdir, upperdir, workdir)

	// 检查 lowerdir 是否存在且有内容
	if files, err := os.ReadDir(lowerdir); err != nil {
		log.Errorf("lowerdir %s not accessible: %v", lowerdir, err)
		return
	} else if len(files) == 0 {
		log.Errorf("lowerdir %s is empty", lowerdir)
		return
	} else {
		log.Infof("lowerdir %s contains %d files", lowerdir, len(files))
		// 显示前几个文件
		for i, file := range files {
			if i < 5 {
				log.Infof("  %s", file.Name())
			}
		}
	}

	// 使用overlayfs
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerdir, upperdir, workdir)
	log.Infof("Mounting overlay with options: %s", opts)

	cmd := exec.Command("mount", "-t", "overlay", "-o", opts, "overlay", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount overlay error: %v", err)
		// 如果overlay失败，尝试bind mount
		fallbackToBindMount(lowerdir, mntURL)
		return
	}
	log.Infof("mount overlay success")

	// 验证挂载后的内容
	if files, err := os.ReadDir(mntURL); err == nil {
		log.Infof("Mount point %s now contains %d files:", mntURL, len(files))
		for i, file := range files {
			if i < 10 {
				log.Infof("  %s", file.Name())
			}
		}
	}
}

// 检查路径是否已经挂载
func isMounted(path string) bool {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == path {
			return true
		}
	}
	return false
}
func isOverlaySupported() bool {
	// 检查/proc/filesystems中是否支持overlay
	data, err := os.ReadFile("/proc/filesystems")
	if err != nil {
		log.Errorf("failed to read /proc/filesystems: %v", err)
		return false
	}

	filesystems := string(data)
	return strings.Contains(filesystems, "overlay") || strings.Contains(filesystems, "overlayfs")
}

func fallbackToBindMount(busyboxURL, mountURL string) {
	log.Infof("falling back to bind mount from %s to %s", busyboxURL, mountURL)
	cmd := exec.Command("mount", "--bind", busyboxURL, mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("bind mount error: %v", err)
		return
	}
	log.Infof("bind mount success")
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeUrls)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
	} else {
		DeleteMountPoint(rootURL, mntURL)
	}
	DeleteWriteLayer(rootURL)
}

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount %s error: %v", mntURL, err)
	} else {
		log.Infof("umount %s success", mntURL)
	}

	// 删除挂载点目录
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("remove mount point %s error: %v", mntURL, err)
	}
}

func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeUrls []string) {
	containerUrl := mntURL + volumeUrls[1]

	// 先卸载volume
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount volume %s error: %v", containerUrl, err)
	}

	// 再卸载主挂载点
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount %s error: %v", mntURL, err)
	} else {
		log.Infof("umount %s success", mntURL)
	}

	// 删除挂载点目录
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("remove mount point %s error: %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "/writeLayer"
	workURL := rootURL + "/work"

	// 删除writeLayer目录
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove write layer %s error: %v", writeURL, err)
	} else {
		log.Infof("remove write layer %s success", writeURL)
	}

	// 删除work目录
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("remove work dir %s error: %v", workURL, err)
	} else {
		log.Infof("remove work dir %s success", workURL)
	}
}
