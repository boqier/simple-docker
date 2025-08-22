package main

import (
	"dockertest1/cgroups/subsystem"
	"dockertest1/container"
	"fmt"

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
		Run(tty, cmdArray, &resConf, volume, containerName)
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
		if len(context.Args().Slice()) < 1 {
			return fmt.Errorf("need at least 1 argument")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		log.Infof("commit come on")
		return nil
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
