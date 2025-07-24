package fs1

import (
	"dockertest1/cgroups/subsystem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// MemorySubsystem implements the Subsystem interface for memory control groups.
func (s *MemorySubsystem) Name() string {
	return "memory"
}

type MemorySubsystem struct {
}

// 设置资源限制
func (s *MemorySubsystem) Set(cgroupPath string, res *subsystem.ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			if err := os.WriteFile(filepath.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set memory limit error: %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}
func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupspath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if err := os.WriteFile(filepath.Join(subsysCgroupspath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("apply memory cgroup error: %v", err)
		}
		return nil
	} else {
		return err
	}
}
func (s *MemorySubsystem) Remove(cgroupPath string) error {

	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}
