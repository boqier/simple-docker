package main

import (
	"dockertest1/container"
	"fmt"
	"io"
	"os"
)

func logContainer(containerID string) {
	logFileLocation := fmt.Sprintf(container.InfoLocFormat, containerID)
	logFileLocation += container.GetLogfile(containerID)
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading log file:", err)
		return
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		fmt.Println("Error printing log content:", err)
		return
	}
}
