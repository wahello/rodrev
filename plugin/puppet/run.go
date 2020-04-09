package puppet

import (
	"bufio"
	"os/exec"
	"sync"
	"time"
)

type  RunStatus struct {
	Busy bool
	Downtime bool
	Started bool
}

type RunOptions struct {
	Delay          time.Duration
	RandomizeDelay bool
}




func (p *Puppet)Run() RunStatus {
	if !p.runLock.TryAcquire(1) {
		return RunStatus{Busy: true}
	} else {
		go p.run(RunOptions{})
		return RunStatus{Started:true}
	}
}

func (p *Puppet)run(opt RunOptions) {
	defer p.runLock.Release(1)
	var err error
	if  opt.Delay > 0 {
		if opt.RandomizeDelay {
			opt.Delay = time.Duration(p.rng.Int63n(opt.Delay.Nanoseconds()))
		}
		p.l.Infof("sleeping %ds before run",int64(opt.Delay.Seconds()))
	}
	p.l.Info("running puppet")
	// TODO remove --test, dont need to log same shit twice
	cmd := exec.Command(p.puppetPath,"agent","--onetime","--no-daemonize","--test")
	stdout, err:= cmd.StdoutPipe()
	if err != nil {
		p.l.Errorf("error attaching stdin: %s", err)
		return
	}
	stderr, err := cmd.StderrPipe()
		if err != nil {
			p.l.Errorf("error attaching stdin: %s", err)
			return
		}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sout := bufio.NewScanner(stdout)
		for sout.Scan() {
			p.l.Infof("+ %s",sout.Text())
		}
	}()
		go func() {
		defer wg.Done()
		serr := bufio.NewScanner(stderr)
		for serr.Scan() {
			p.l.Infof("! %s",serr.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		p.l.Errorf("error starting puppet: %s", err)
		return
	}
	wg.Wait()
	err = 	cmd.Wait()
	if err != nil {
		p.l.Errorf("error after puppet run: %s", err)
		return
	}
	p.l.Infof("puppet run finished")
	return

}