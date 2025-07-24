package fs2

import (
	"dockertest1/cgroups/subsystem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type MemorySubsystem struct {
}

func (s *MemorySubsystem) Name() string {
	return "memory"
}
func (s *MemorySubsystem) Set(cgroupPath string, res *subsystem.ResourceConfig) error {
	log.Info("set memory limit is: %v", res.MemoryLimit)
	if subsysCgroupPath, err := GetCgroupPathV2(cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			if err := os.WriteFile(filepath.Join(subsysCgroupPath, "memory.max"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set memory limit error: %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}
func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupspath, err := GetCgroupPathV2(cgroupPath, true); err == nil {
		if err := os.WriteFile(filepath.Join(subsysCgroupspath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("apply memory cgroup error: %v", err)
		}
		return nil
	} else {
		return err
	}
}
func (s *MemorySubsystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := GetCgroupPathV2(cgroupPath, false)
	if err != nil {
		return err
	}

	// 1. 读取当前 cgroup 中的所有进程
	procsFile := filepath.Join(subsysCgroupPath, "cgroup.procs")
	pids, err := os.ReadFile(procsFile)
	if err != nil {
		return fmt.Errorf("read cgroup.procs failed: %v", err)
	}

	// 2. 将每个进程移出到根 cgroup
	for _, pid := range strings.Split(strings.TrimSpace(string(pids)), "\n") {
		if pid == "" {
			continue
		}
		if err := os.WriteFile("/sys/fs/cgroup/cgroup.procs", []byte(pid), 0644); err != nil {
			return fmt.Errorf("move pid %s failed: %v", pid, err)
		}
	}

	// 3. 确认 cgroup.procs 已空
	if remaining, _ := os.ReadFile(procsFile); len(remaining) > 0 {
		return fmt.Errorf("cgroup still has processes: %s", string(remaining))
	}

	// 4. 删除目录
	return os.Remove(subsysCgroupPath)
}
