package main

import (
	"dockertest1/cgroups/subsystem"
	"dockertest1/container"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var runCommand = &cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroup limit mydocker run -it [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		&cli.StringFlag{
			Name:  "m",
			Usage: "memory limit,e.g: -men 100m",
		},
		&cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu limit,e.g: -cpu 100",
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g: -setcpu 100",
		},
		&cli.StringFlag{
			Name:  "v",
			Usage: "volume: -v /etc/conf:/mydocker/conf",
		},
		&cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args().Slice()) < 1 {
			return fmt.Errorf("need at least 1 argument")
		}
		var cmdArray []string
		for _, arg := range context.Args().Slice() {
			cmdArray = append(cmdArray, arg)
		}
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
		tty := context.Bool("it")
		deatch := context.Bool("d")
		if tty && deatch {
			return fmt.Errorf("it and d paramter can not both provided")
		}
		volume := context.String("v")
		containerName := context.String("name")
		resConf := subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}
		

		Run(tty, cmdArray, &resConf, volume, containerName,imageName)
		return nil
	},
}
var initCommand = &cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("command: %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}
var commitCommand = &cli.Command{
	Name:  "commit",
	Usage: "Commit a container to an image",
	Action: func(context *cli.Context) error {
		if len(context.Args().Slice()) < 2 {
			return fmt.Errorf("miss ing container name or image name")
		}
		containerId:= context.Args().Get(0)
		imageName:= context.Args().Get(1)
		return commitContainer(containerId,imageName)
	},
}

var listCommand = &cli.Command{
	Name:  "ps",
	Usage: "List all the containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}
var logCommand = &cli.Command{
	Name:  "logs",
	Usage: "Show logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args().Slice()) < 1 {
			return fmt.Errorf("need at least 1 argument")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}
var stopCommand = &cli.Command{
	Name:  "stop",
	Usage: "Stop a running container",
	Action: func(context *cli.Context) error {
		if len(context.Args().Slice()) < 1 {
			return fmt.Errorf("need at least 1 argument")
		}
		containerName := context.Args().Get(0)
		stopContainer(containerName)
		log.Infof("stop come on")
		return nil
	},
}
var removeCommand = &cli.Command{
	Name:  "rm",
	Usage: "Remove a container",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "f",
			Usage: "Force remove a running container",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args().Slice()) < 1 {
			return fmt.Errorf("need at least 1 argument")
		}
		containerId := context.Args().Get(0)
		force := context.Bool("f")
		removeContainer(containerId, force)
		log.Infof("remove come on")
		return nil
	},
}
var execCommand = &cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		// 如果环境变量存在，说明C代码已经运行过了，即setns系统调用已经执行了，这里就直接返回，避免重复执行
		if os.Getenv(EnvExecPid) != "" {
			log.Infof("pid callback pid %v", os.Getgid())
			return nil
		}
		// 格式：mydocker exec 容器名字 命令，因此至少会有两个参数
		if len(context.Args().Slice()) < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := context.Args().Get(0)
		// 将除了容器名之外的参数作为命令部分
		commandArray := context.Args().Tail()
		ExecContainer(containerName, commandArray)
		return nil
	},
}