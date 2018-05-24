package main

import (
	"log"
	"os"
	"github.com/urfave/cli"
)

var config ServerConfig
var err error
var tick int
// var emitter EventEmitter = GetEmmiter()

func main() {
	emitter := GetEmitter()

	defineEmitter(emitter)

	app := cli.NewApp()

	app.Name = "RabbitMQ Monitor tool"
	app.Usage = "cli tool to monitor a RabbitMQ Server"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host",
			Usage:       "server host",
			Destination: &config.Host,
			Value:       "localhost",
		},
		cli.StringFlag{
			Name:        "user",
			Usage:       "user name that will access the API",
			Destination: &config.User,
		},
		cli.StringFlag{
			Name:        "password",
			Usage:       "user password",
			Destination: &config.Password,
		},
		cli.StringFlag{
			Name:        "vhost",
			Usage:       "vhost to monitor",
			Destination: &config.Vhost,
		},
		cli.IntFlag{
			Name:        "port",
			Usage:       "API port (default 15672)",
			Destination: &config.Port,
			Value:       15672,
		},
		cli.IntFlag{
			Name:        "tick",
			Usage:       "ticker time lag",
			Destination: &tick,
			Value:       1000,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "monitor",
			Aliases: []string{"l"},
			Flags: append(app.Flags, []cli.Flag{}...),
			Usage: "Wait tasks by listening to queue server",
			Action: func(c *cli.Context) error {
				log.Println("[INFO] Checking server config...", config)
				if err = config.Validate(); err != nil {
					return err
				}

				monitor := Monitor(config)

				monitor.Tick(tick)

				return err
			},
		},
	}

	app.Action = app.Commands[0].Action

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}