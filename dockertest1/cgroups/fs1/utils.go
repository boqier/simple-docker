package fs1

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

func FindCgroupMountpoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	// If we reach here, no mount point was found for the subsystem.
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading mountinfo:", err)
		return ""
	}
	return ""
}

func GetCgroupPath(subsystem, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("failed to create cgroup path: %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	}
	return "", fmt.Errorf("cgroup path err: %s", cgroupPath)
}
