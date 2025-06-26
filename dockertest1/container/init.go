package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const fdIndex = 3

// 挂载proc文件系统视图
func mountProc() {
	log.Infof("mounting proc fd: %d", fdIndex)
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
}
func RunContainerInitProcess(command string, args []string) error {
	//mount /proc文件系统
	mountProc()
	cmdArray := readUserCommamd()
	if len(cmdArray) == 0 {
		return fmt.Errorf("No command to execute")
	}
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Infof("Failed loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", path)
	if err = syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf("RunContainerInitProcess exec :" + err.Error())
	}
	return nil
}
func readUserCommamd() []string {
	prpe := os.NewFile(uintptr(fdIndex), "pipe")
	msg, err := ioutil.ReadAll(prpe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("new pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	//将readpipe作为ExtraFiles，这样cmd执行时就会带着这个文件句柄创建子线程
	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe
}
