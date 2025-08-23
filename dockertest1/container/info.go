package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	RUNNING       = "running"
	STOP          = "stopped"
	Exit          = "exited"
	InfoLoc       = "/var/lib/mydocker/containers/"
	InfoLocFormat = InfoLoc + "%s/"
	ConfigName    = "config.json"
	IDLength      = 10
	LogFile       = "%s-json.log"
)

type Info struct {
	Pid         string `json:"pid"`        // 容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内init运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
	Volume      string `json:"volume"`     // 容器的卷信息
}

func RecordContainerInfo(containerPID int, commandArray []string, containerName, containerId, volume string) error {
	// 如果未指定容器名，则使用随机生成的containerID
	if containerName == "" {
		containerName = containerId
	}
	command := strings.Join(commandArray, "")
	containerInfo := &Info{
		Id:          containerId,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:      RUNNING,
		Name:        containerName,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return errors.WithMessage(err, "container info marshal failed")
	}
	jsonStr := string(jsonBytes)
	// 拼接出存储容器信息文件的路径，如果目录不存在则级联创建
	dirPath := fmt.Sprintf(InfoLocFormat, containerId)
	if err = os.MkdirAll(dirPath, 0622); err != nil {
		return errors.WithMessagef(err, "mkdir %s failed", dirPath)
	}
	// 将容器信息写入文件
	fileName := path.Join(dirPath, ConfigName)
	file, err := os.Create(fileName)
	if err != nil {
		return errors.WithMessagef(err, "create file %s failed", fileName)
	}
	defer file.Close()
	if _, err = file.WriteString(jsonStr); err != nil {
		return errors.WithMessagef(err, "write container info to  file %s failed", fileName)
	}
	return nil
}

func DeleteContainerInfo(containerId string) {
	dirPath := fmt.Sprintf(InfoLocFormat, containerId)
	if err := os.RemoveAll(dirPath); err != nil {
		log.Errorf("Remove dir %s error. Error: %v", dirPath, err)
	}
}
func GetLogfile(containerId string) string {
	return fmt.Sprintf(LogFile, containerId)
}
