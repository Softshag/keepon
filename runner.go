package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

type RunnerConfig struct {
	Exec    string
	Args    []string
	Workdir string
	Retries int64
	Timeout int64
	Sleep   time.Duration
	Stdout  io.Writer
	Stderr  io.Writer
}

type Runner struct {
	RunnerConfig
	running bool
	mu      sync.Mutex
	cmd     *exec.Cmd
}

func (r *Runner) run(quit chan bool, binary string, env []string) error {
	r.cmd = nil
	cmd := &exec.Cmd{
		Path:   binary,
		Args:   append([]string{r.Exec}, r.Args...),
		Env:    env,
		Stdout: r.Stdout,
		Stderr: r.Stderr,
	}

	if r.Workdir != "" {
		cmd.Dir = r.Workdir
	}

	stop := make(chan bool)

	err := cmd.Start()

	if err != nil {
		return err
	}

	go func() {
	listen:
		for {
			select {
			case <-quit:
				cmd.Process.Kill()
				r.cmd = nil
				break listen
			case <-stop:
				break listen
			}
		}
	}()

	r.cmd = cmd

	err = cmd.Wait()

	if err == nil {
		err = errors.New("Finished to soon")
	}

	stop <- true

	r.cmd = nil

	return err
}

func (r *Runner) IsRunning() (x bool) {
	r.mu.Lock()
	x = r.running
	r.mu.Unlock()
	return
}

func (r *Runner) Run() (chan<- bool, chan error, chan error) {
	if r.IsRunning() {
		panic(errors.New("Already running"))
	}

	r.mu.Lock()
	r.running = true
	r.mu.Unlock()

	quit := make(chan bool)
	errchan := make(chan error)
	retrychan := make(chan error)

	env := os.Environ()

	go func() {
		p, _ := exec.LookPath(r.Exec)

		retries := r.Retries
		if retries == 0 {
			retries = 1
		}

		var exitCode error
		for i := 0; int64(i) < retries; i++ {
			exitCode = r.run(quit, p, env)

			if i < int(retries-1) {
				retrychan <- exitCode
			}

			if r.Sleep > 0 {
				time.Sleep(r.Sleep * time.Second)
			}
		}
		errchan <- exitCode

		r.mu.Lock()
		r.running = false
		r.mu.Unlock()

	}()

	return quit, errchan, retrychan
}

func (r *Runner) RunSync(fn func(error)) error {
	_, errchan, retrychan := r.Run()
	var out_error error

loop:
	for {
		select {
		case e, ok := <-errchan:
			if ok {
				out_error = e
			}
			break loop
		case e, ok := <-retrychan:
			if ok && fn != nil {
				fn(e)
			}
		}
	}

	return out_error

}

func (r *Runner) RunSyncAtMost(duration time.Duration, fn func(error)) error {
	quit, errchan, retrychan := r.Run()
	var out_error error
	time.AfterFunc(duration, func() {
		quit <- true
	})
loop:
	for {
		select {
		case e, ok := <-errchan:
			if ok {
				out_error = e
			}
			break loop
		case e, ok := <-retrychan:
			if ok && fn != nil {
				fn(e)
			}
		}
	}

	return out_error
}

func NewAsyncRunner(config RunnerConfig) *Runner {
	return &Runner{
		RunnerConfig: config,
		running:      false,
	}
}
