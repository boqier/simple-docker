package container

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const fdIndex = 3

// 挂载proc文件系统视图
func mountProc() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)

	// 执行pivot_root
	if err := pivotRoot(pwd); err != nil {
		log.Errorf("pivot root error: %v", err)
		return
	}

	log.Infof("pivot root success, now mounting proc")

	// 确保proc目录存在
	if err := os.MkdirAll("/proc", 0755); err != nil {
		log.Errorf("create /proc dir error: %v", err)
	}

	log.Infof("mounting proc fd: %d", fdIndex)
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("mount proc error: %v", err)
	} else {
		log.Infof("mount proc success")
	}
}

func pivotRoot(root string) error {
	log.Infof("Starting pivot_root from %s", root)

	// 检查root目录内容
	if files, err := os.ReadDir(root); err == nil {
		log.Infof("Files in %s before pivot_root:", root)
		for _, file := range files {
			log.Infof("  %s", file.Name())
		}
	}

	// 检查root是否是挂载点
	if !isMountPoint(root) {
		log.Errorf("%s is not a mount point", root)
		return fmt.Errorf("%s is not a mount point", root)
	}

	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	log.Infof("Bind mount %s to itself success", root)

	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("Create pivot dir %s error: %v", pivotDir, err)
	}
	log.Infof("Created pivot dir %s", pivotDir)

	// pivot_root 到新的rootfs, 现在老的 old_root 是挂载在rootfs/.pivot_root
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	log.Infof("PivotRoot success")

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	log.Infof("Changed to root directory")

	// 验证切换是否成功
	if pwd, err := os.Getwd(); err == nil {
		log.Infof("Current working directory after chdir: %s", pwd)
	}

	// 检查新根目录的内容
	if files, err := os.ReadDir("/"); err == nil {
		log.Infof("Files in new root directory:")
		for _, file := range files {
			log.Infof("  /%s", file.Name())
		}
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		log.Warnf("unmount pivot_root dir %v", err)
	} else {
		log.Infof("Unmounted old root")
	}

	// 删除临时文件夹
	if err := os.Remove(pivotDir); err != nil {
		log.Warnf("remove pivot_root dir %v", err)
	} else {
		log.Infof("Removed pivot dir")
	}

	log.Infof("pivot_root completed successfully")
	return nil
}

// 检查路径是否是挂载点
func isMountPoint(path string) bool {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		log.Errorf("failed to read /proc/mounts: %v", err)
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == path {
			log.Infof("Found mount point: %s", path)
			return true
		}
	}
	log.Errorf("%s is not a mount point", path)
	return false
}

func RunContainerInitProcess(command string, args []string) error {
	log.Infof("init come on")
	log.Infof("command: %s", command)

	// 在pivot_root之前检查当前目录
	if pwd, err := os.Getwd(); err == nil {
		log.Infof("Before pivot_root, current dir: %s", pwd)
		if files, err := os.ReadDir("."); err == nil {
			log.Infof("Files in current dir before pivot_root:")
			for _, file := range files {
				log.Infof("  %s", file.Name())
			}
		}
	}

	//mount /proc文件系统
	mountProc()

	// 在pivot_root之后检查当前目录
	if pwd, err := os.Getwd(); err == nil {
		log.Infof("After pivot_root, current dir: %s", pwd)
		if files, err := os.ReadDir("."); err == nil {
			log.Infof("Files in current dir after pivot_root:")
			for _, file := range files {
				log.Infof("  %s", file.Name())
			}
		}
	}

	// 检查根目录
	if files, err := os.ReadDir("/"); err == nil {
		log.Infof("Files in root directory after pivot_root:")
		for _, file := range files {
			log.Infof("  /%s", file.Name())
		}
	}

	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return fmt.Errorf("No command to execute")
	}

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Infof("Failed to find path error %v", err)
		return err
	}
	log.Infof("Find path %s", path)

	if err = syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		log.Errorf("RunContainerInitProcess exec: %v", err)
	}
	return nil
}

func readUserCommand() []string {
	// 从文件描述符3创建文件对象
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	if pipe == nil {
		log.Errorf("failed to create pipe file descriptor")
		return nil
	}
	defer pipe.Close()

	// 直接从管道读取数据，而不是读取文件名
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}

	msgStr := strings.TrimSpace(string(msg))
	if msgStr == "" {
		log.Errorf("received empty command from pipe")
		return nil
	}

	log.Infof("received command from pipe: %s", msgStr)
	return strings.Fields(msgStr) // 使用Fields而不是Split，更好地处理空格
}

func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("new pipe error %v", err)
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
	//将readpipe作为ExtraFiles，这样cmd执行时就会带着这个文件句柄创建子线程
	cmd.ExtraFiles = []*os.File{readPipe}
	mntURL := "/root/mnt"
	rootURL := "/root"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}
