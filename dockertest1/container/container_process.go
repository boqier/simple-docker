package container

import (
	"os"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	// 创建匿名管道用于传递参数，将readPipe作为子进程的ExtraFiles，子进程从readPipe中读取参数
	// 父进程中则通过writePipe将参数写入管道
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	//指定 cmd 的工作目录为我们前面准备好的用于存放 busybox rootfs的目录，暂时固定为 /root/busybox
	mntURL := "/root/merged/"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mntURL)
	cmd.Dir = mntURL
	return cmd, writePipe
}
func NewWorkSpace(rootPath string, mntURL string) {
	createLower(rootPath)
	createDirs(rootPath)
	mountOverlayFS(rootPath, mntURL)
}
func createLower(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Errorf("Fail to judge whether dir %s exists. Error: %v", busyboxURL, err)
	}
	if !exist {
		{
			if err := os.MkdirAll(busyboxURL, 0755); err != nil {
				log.Errorf("Mkdir dir %s error. Error: %v", busyboxURL, err)
			}
			if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
				log.Errorf("Untar dir %s error. Error: %v", busyboxURL, err)
			}
		}

	}
}
func createDirs(rootURL string) {
	upperURL := rootURL + "upper/"
	if err := os.Mkdir(upperURL, 0755); err != nil {
		log.Errorf("Mkdir dir %s error. Error: %v", upperURL, err)
	}
	workURL := rootURL + "work/"
	if err := os.Mkdir(workURL, 0755); err != nil {
		log.Errorf("Mkdir dir %s error. Error: %v", workURL, err)
	}
}
func mountOverlayFS(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0755); err != nil {
		log.Errorf("Mkdir dir %s error. Error: %v", mntURL, err)
	}
	dirs := "lowerdir=" + rootURL + "busybox,upperdir=" + rootURL + "upper,workdir=" + rootURL + "work"
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount overlayfs error. Error: %v", err)
	}
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
func DeleteWorkSpace(rootURL string, mntURL string) {
	umountOverlayFS(mntURL)
	deleteDirs(rootURL)
}
func umountOverlayFS(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount overlayfs error. Error: %v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error. Error: %v", mntURL, err)
	}
}
func deleteDirs(rootURL string) {
	wirteURL := rootURL + "upper/"
	if err := os.RemoveAll(wirteURL); err != nil {
		log.Errorf("Remove dir %s error. Error: %v", wirteURL, err)
	}
	workURL := rootURL + "work"
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("Remove dir %s error. Error: %v", workURL, err)
	}
}
