package main

import (
	"encoding/json"
	"github.com/codegangsta/cli"

	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func parseArgs(c *cli.Context) RunnerConfig {
	bArgs := normalizeArgs(c.Args())
	var binary string
	var iArgs []string

	if i := c.String("interpreter"); i != "" {
		binary = i
		iArgs = c.StringSlice("iargs")
	} else {
		binary = bArgs[0]
		bArgs = bArgs[1:]
	}

	bArgs = append(iArgs, bArgs...)

	config := RunnerConfig{
		Exec:    binary,
		Args:    bArgs,
		Retries: int64(c.Int("retries")),
		Sleep:   time.Duration(c.Int("sleep")),
		Timeout: int64(c.Int("timeout")),
	}
	return config
}

func onRunError(script string) {
	fp, err := filepath.Abs(script)
	check(err)

	ext := filepath.Ext(script)

	var config RunnerConfig
	if ext == ".json" {
		_config, err := loadConfigFromPath(fp)
		check(err)
		config = _config
	} else {
		config = RunnerConfig{
			Exec: script,
		}
	}
	runner := NewAsyncRunner(config)

	runner.RunSync(nil)

}

func loadConfigFromPath(path string) (RunnerConfig, error) {
	var config RunnerConfig
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func runRunner(c *cli.Context) {
	var config RunnerConfig
	var err error
	if configStr := c.String("config"); configStr != "" {
		config, err = loadConfigFromPath(configStr)
		check(err)

	} else {
		if len(c.Args()) == 0 {
			log.Println("You must specify what to run")
			os.Exit(1)
		}
		config = parseArgs(c)
	}

	if out := c.String("out"); out != "" {
		if out == "stdout" {
			config.Stdout = os.Stdout
		} else {
			st, err := os.Create(out)
			check(err)
			config.Stdout = st
		}
	}

	if error_out := c.String("error"); error_out != "" {
		if error_out == "stderr" {
			config.Stderr = os.Stderr
		} else {
			st, err := os.Create(error_out)
			check(err)
			config.Stderr = st
		}

	}

	runner := NewAsyncRunner(config)

	var i int64 = 0
	err = runner.RunSync(func(exitCode error) {
		i++
		log.Printf("Retrying .... %d attempt(s) remaining\n", config.Retries-i)
	})

	if onError := c.String("on-error"); onError != "" {
		onRunError(onError)
	}

	log.Printf("Error: %s", err.Error())
	//time.AfterFunc(time.Second*10, func() { quit <- true })
}
