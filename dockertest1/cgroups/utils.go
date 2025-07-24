package cgroups

import (
	"os"
	"path/filepath"
)

const (
	CgroupV1 = "v1"
	CgroupV2 = "v2"
)

// DetectCgroupVersion 检测当前系统的 cgroup 版本
func DetectCgroupVersion() string {
	// 检查 cgroups v2 的标记文件
	if _, err := os.Stat(filepath.Join("/sys/fs/cgroup", "cgroup.controllers")); err == nil {
		return CgroupV2
	}
	// 默认返回 v1（兼容旧内核）
	return CgroupV1
}
