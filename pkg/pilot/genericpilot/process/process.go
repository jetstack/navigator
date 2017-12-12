package process

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/golang/glog"
)

type Interface interface {
	// Start starts the underlying process.
	Start() error
	// Stop should request the process stop. It should not return until the
	// the process has exited.
	Stop() error
	// Terminate should terminate the currently running process immediately,
	// usually by sending a SIGKILL signal.
	Terminate() error
	// Reload should request the process reload it's configuration. This
	// should not trigger the process itself to exit or interrupt a Run() call.
	Reload() error
	// Wait will call Wait on the underlying process
	Wait() error
	// Running returns true if the process is running
	Running() bool
	// State returns the state of an exited process
	State() *os.ProcessState
	// String returns a textual represntation of the process
	String() string
}

type Adapter struct {
	Signals Signals
	Cmd     *exec.Cmd
	wg      sync.WaitGroup
}

var _ Interface = &Adapter{}

func (p *Adapter) startCommandOutputLoggers() error {
	stdout, err := p.Cmd.StdoutPipe()
	if err != nil {
		return err
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			glog.Infoln(in.Text())
		}
		err := in.Err()
		if err != nil {
			glog.Error(err)
		}
	}()

	stderr, err := p.Cmd.StderrPipe()
	if err != nil {
		return err
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			glog.Infoln(in.Text())
		}
		err := in.Err()
		if err != nil {
			glog.Error(err)
		}
	}()
	return nil
}

func (p *Adapter) Start() error {
	glog.V(2).Infof("Starting process: %v", p.Cmd.Args)

	err := p.startCommandOutputLoggers()
	if err != nil {
		return err
	}

	if err = p.Cmd.Start(); err != nil {
		return fmt.Errorf("error starting process: %s", err.Error())
	}
	return nil
}

func (p *Adapter) Wait() error {
	defer p.wg.Wait()
	return p.Cmd.Wait()
}

func (p *Adapter) Stop() error {
	if p.Cmd == nil {
		return fmt.Errorf("must call Run() before Stop()")
	}
	p.Cmd.Process.Signal(p.Signals.Stop)
	return p.Cmd.Wait()
}

func (p *Adapter) Terminate() error {
	if p.Cmd == nil {
		return fmt.Errorf("must call Run() before Terminate()")
	}
	p.Cmd.Process.Signal(p.Signals.Terminate)
	return p.Cmd.Wait()
}

func (p *Adapter) Reload() error {
	if p.Cmd == nil {
		return fmt.Errorf("must call Run() before Reload()")
	}
	p.Cmd.Process.Signal(p.Signals.Reload)
	return p.Cmd.Wait()
}

func (p *Adapter) Running() bool {
	if p.Cmd == nil || p.Cmd.Process == nil || p.Cmd.Process.Pid == 0 || p.State() != nil {
		return false
	}
	return true
}

func (p *Adapter) State() *os.ProcessState {
	if p.Cmd == nil {
		return nil
	}
	return p.Cmd.ProcessState
}

func (p *Adapter) String() string {
	if p.Cmd == nil || p.Cmd.Process == nil {
		return fmt.Sprintf("inactive")
	}
	return fmt.Sprintf("%d", p.Cmd.Process.Pid)
}
