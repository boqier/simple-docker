package main

import (
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
			Name:  "mem",
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

		resConf := cgroups.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}
		Run(tty, cmdArray, resConf)
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
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}
