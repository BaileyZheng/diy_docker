package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/BaileyZheng/diy_docker/container"
	"github.com/BaileyZheng/diy_docker/cgroups/subsystems"
	"github.com/BaileyZheng/diy_docker/network"
	"os"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
		code-5.1 run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name: "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name: "m",
			Usage: "memory limie",
		},
		cli.StringFlag{
			Name: "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name: "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name: "name",
			Usage: "container name",
		},
		cli.StringFlag{
			Name: "v",
			Usage: "volume",
		},
		cli.StringFlag{
			Name: "image",
			Usage: "image",
		},
		cli.StringSliceFlag{
			Name: "e",
			Usage: "set environment",
		},
		cli.StringFlag{
			Name: "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name: "p",
			Usage: "port mapping",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args())<1{
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args(){
			cmdArray=append(cmdArray,arg)
		}
		createTty := context.Bool("ti")
		detach := context.Bool("d")
		
		if createTty && detach {
			return fmt.Errorf("ti and d parameter can not both provided")
		}
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet: context.String("cpuset"),
			CpuShare: context.String("cpushare"),
		}
		log.Infof("createTty %v", createTty)
		containerName:=context.String("name")
		imageName:=context.String("image")
		volume:=context.String("v")
		network:=context.String("net")
		envSlice:=context.StringSlice("e")
		portmapping:=context.StringSlice("p")
		Run(createTty, cmdArray, resConf,containerName,volume,imageName,envSlice,network,portmapping)
		return nil
	},
}

var initCommand = cli.Command{
	Name: "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error{
		log.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}

var listCommand = cli.Command{
	Name: "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error{
		ListContainers()
		return nil
	},
	
}

var logCommand = cli.Command{
	Name: "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error{
		if len(context.Args())<1{
			return fmt.Errorf("Please input your container name")
		}
		containerName:=context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand=cli.Command{
	Name: "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error{
		if os.Getenv(ENV_EXEC_PID) != ""{
			log.Infof("pid callback pid %s",os.Getpid())
			return nil
		}
		if len(context.Args())<2{
			return fmt.Errorf("Missing container name or command")
		}
		containerName:=context.Args().Get(0)
		var commandArray []string
		for _,arg:=range context.Args().Tail(){
			commandArray=append(commandArray,arg)
		}
		ExecContainer(containerName,commandArray)
		return nil
	},
}

var stopCommand=cli.Command{
	Name: "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error{
		if len(context.Args())<1{
			return fmt.Errorf("Missing container name")
		}
		containerName:=context.Args().Get(0)
		stopContainer(containerName)
		return nil
	},
}

var removeCommand=cli.Command{
	Name: "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error{
		if len(context.Args())<1{
			return fmt.Errorf("missing container name")
		}
		containerName:=context.Args().Get(0)
		removeContainer(containerName)
		return nil
	},
}

var commitCommand=cli.Command{
	Name: "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error{
		if len(context.Args())<2{
			return fmt.Errorf("missing container name and image name")
		}
		containerName:=context.Args().Get(0)
		imageName:=context.Args().Get(1)
		commitContainer(containerName,imageName)
		return nil
	},
}

var networkCommand=cli.Command{
	Name: "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{	
			Name: "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "subnet",
					Usage: "subnet cide",
				},
				cli.StringFlag{
					Name: "driver",
					Usage: "network driver",
				},
			},
			Action: func(context *cli.Context) error{
				if len(context.Args())<1{
					return fmt.Errorf("Missing network name")
				}
				network.Init()
				driver:=context.String("driver")
				subnet:=context.String("subnet")
				name:=context.Args()[0]
				log.Infof("driver:%s, subnet:%s, name:%s",driver,subnet,name)
				if err:=network.CreateNetwork(driver,subnet,name);err!=nil{
					return fmt.Errorf("create network error %+v",err)
				}
				return nil
			},
		},
	},
}
