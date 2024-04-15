package main

import (
	"fmt"
	"time"

	"github.com/draganm/gophergazer/build"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{

		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:    "debounce-time",
				Usage:   "debounce time",
				EnvVars: []string{"DEBOUNCE_TIME"},
				Value:   300 * time.Millisecond,
			},
		},

		Action: func(c *cli.Context) error {

			if c.Args().Len() == 0 {
				return fmt.Errorf("no module directory provided")
			}

			moduleDir := c.Args().First()

			binary, err := build.BuildBinary(c.Context, moduleDir)
			if err != nil {
				return err
			}

			fmt.Println(binary, c.Args().Slice())
			return nil
		},
	}
	app.RunAndExitOnError()
}
