// cgroups/manager.go
package cgroups

import (
	"dockertest1/cgroups/fs1" // v1实现
	"dockertest1/cgroups/fs2" // v2实现
	"dockertest1/cgroups/subsystem"
	"os"
)

type cgroupManager struct {
	Path       string
	Version    string                // "v1" 或 "v2"
	subsystems []subsystem.Subsystem // 通过接口解耦具体实现
}

func NewCgroupManager(path string) *cgroupManager {
	version := detectCgroupVersion()
	var subs []subsystem.Subsystem

	// 根据版本选择子系统实现
	if version == "v1" {
		subs = fs1.Subsystems
	} else {
		subs = fs2.Subsystems
	}

	return &cgroupManager{
		Path:       path,
		Version:    version,
		subsystems: subs,
	}
}

// 私有函数检测版本（避免导出）
func detectCgroupVersion() string {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		return "v2"
	}
	return "v1"
}

func (c *cgroupManager) Apply(pid int) error {
	for _, sub := range c.subsystems {
		if err := sub.Apply(c.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

func (c *cgroupManager) Set(res *subsystem.ResourceConfig) error {
	for _, sub := range c.subsystems {
		if err := sub.Set(c.Path, res); err != nil {
			return err
		}
	}
	return nil
}

func (c *cgroupManager) Remove() error {
	for _, sub := range c.subsystems {
		if err := sub.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}
