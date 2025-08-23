package container

import (
	fmt "fmt"
	"os"
	"os/exec"
	"syscall"
	"dockertest1/utils"
	log "github.com/sirupsen/logrus"
)

func NewParentProcess(tty bool, volume string, containerId string,imageName string) (*exec.Cmd, *os.File) {
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
	} else {
		dirPath := fmt.Sprintf(InfoLocFormat, containerId)
		if err := os.Mkdir(dirPath, 0622); err != nil {
			log.Errorf("Mkdir %s failed: %v", dirPath, err)
			return nil, nil
		}
		stdLohFilePath := dirPath + GetLogfile(containerId)
		stdLogFile, err := os.Create(stdLohFilePath)
		if err != nil {
			log.Errorf("Create log file %s failed: %v", stdLohFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	
	NewWorkSpace(containerId, imageName, volume)
	cmd.Dir = utils.GetMerged(containerId)
	return cmd, writePipe
}
