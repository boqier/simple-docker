package fs2

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetCgroupPathV2 获取或创建 cgroups v2 的控制组路径
// cgroupPath: 相对于 cgroup2 根目录的子路径（如 "my_container"）
// autoCreate: 如果路径不存在，是否自动创建
func GetCgroupPathV2(cgroupPath string, autoCreate bool) (string, error) {
	// cgroups v2 的挂载点通常是 /sys/fs/cgroup
	cgroupRoot := "/sys/fs/cgroup"

	// 检查 cgroup2 是否已挂载
	if _, err := os.Stat(filepath.Join(cgroupRoot, "cgroup.controllers")); err != nil {
		return "", fmt.Errorf("cgroups v2 not mounted at %s: %v", cgroupRoot, err)
	}

	fullPath := filepath.Join(cgroupRoot, cgroupPath)

	// 检查路径是否存在
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check cgroup path: %v", err)
	}

	// 自动创建（如果需要）
	if autoCreate {
		if err := os.Mkdir(fullPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create cgroup path: %v", err)
		}
		return fullPath, nil
	}

	return "", fmt.Errorf("cgroup path does not exist: %s", fullPath)
}
