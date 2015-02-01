package main

import (
	"fmt"

	"github.com/codegangsta/cli"

	"os"
)

func main() {

	app := cli.NewApp()
	app.Name = "Keepon"
	app.Version = "0.0.1"
	app.Author = "Rasmus Kildevæld"
	app.Email = "rasmuskildevaeld@gmail.com"

	app.Flags = Flags()
	app.Action = runRunner
	err := app.Run(os.Args)

	if err != nil {
		fmt.Printf("Error %s\n", err.Error())
	}

}

func Flags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  "retries, r",
			Value: 10,
		},
		cli.StringFlag{
			Name: "interpreter, i",
		},
		cli.StringSliceFlag{
			Name:  "iargs",
			Usage: "Interpreter arguments",
			Value: &cli.StringSlice{},
		},
		cli.IntFlag{
			Name:  "sleep, s",
			Usage: "Sleep seconds before re-executing",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "timeout, t",
			Value: 60,
		},
		cli.StringFlag{
			Name: "config, c",
		},
		cli.StringFlag{
			Name: "on-error",
		},
		cli.StringFlag{
			Name: "on-retry",
		},
		cli.StringFlag{
			Name: "out, o",
		},
		cli.StringFlag{
			Name: "error, e",
		},
	}
}

func normalizeArgs(args []string) []string {
	out := []string{}
	for _, v := range args {
		if v != "--" {
			out = append(out, v)
		}
	}
	return out
}
