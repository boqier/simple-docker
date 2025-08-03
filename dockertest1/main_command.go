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
			Usage: "volume mount, e.g: -v /host/path:/container/path",
		},
		&cli.StringFlag{
			Name: "d"
			Usage: "detach container"
		},
		&cli.BoolFlag{
			Name:  "name",
			Usage: "name of the container"
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("need at least 1 argument")
		}
		var cmdArray []string
		for _, arg := range context.Args().Slice() {
			cmdArray = append(cmdArray, arg)
		}
		Createtty := context.Bool("it")
		detach:= context.Bool("d")
		if detach&&Createtty {
			return fmt.Errorf("cannot use -it and -d at the same time")
		}
		volume := context.String("v")
		resConf := subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}
		containerName := context.String("name")
		Run(Createtty, cmdArray, &resConf, volume,containerName)
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
var commitCommand = &cli.Command{
	Name:  "commit",
	usage: "commit a container into image",
	Action: func(context *cli.Context) error {
	if len(context.Args())<1 {
		return fmt.Errorf("need at least 1 argument")
	}
	imageName:= context.Args().Get(0)
	log.Infof("commit image: %s", imageName)
	commitContainer(imageName)
	return nil
},
}

