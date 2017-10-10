package process

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

type Interface interface {
	// Run should start the process. It should block until the process exits.
	Run() error
	// Stop should request the process stop. It should not return until the
	// the process has exited.
	Stop() error
	// Terminate should terminate the currently running process immediately,
	// usually by sending a SIGKILL signal.
	Terminate() error
	// Reload should request the process reload it's configuration. This
	// should not trigger the process itself to exit or interrupt a Run() call.
	Reload() error
	// State should return the current state of the process.
	State() *os.ProcessState
}

type Adapter struct {
	Signals Signals
	Stdout  *os.File
	Stderr  *os.File
	Cmd     *exec.Cmd
}

var _ Interface = &Adapter{}

func (p *Adapter) Run() error {
	glog.V(2).Infof("Starting process: %v", p.Cmd.Args)

	if err := p.Cmd.Start(); err != nil {
		return fmt.Errorf("error starting process: %s", err.Error())
	}

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

func (p *Adapter) State() *os.ProcessState {
	if p.Cmd == nil {
		return nil
	}
	return p.Cmd.ProcessState
}
