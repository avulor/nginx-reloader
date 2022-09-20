package utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type NginxRunner struct {
	nginxProc    *os.Process
	signalsChan  chan os.Signal
	changeChan   chan bool
	logFullChan  chan bool
	logFile      string
	NginxCommand []string
}

func MakeNginxRunner(changeChan chan bool, logFullChan chan bool, logFile string, nginxCommand []string) NginxRunner {
	nginxRunner := NginxRunner{
		signalsChan:  make(chan os.Signal),
		changeChan:   changeChan,
		logFullChan:  logFullChan,
		logFile:      logFile,
		NginxCommand: nginxCommand,
	}
	return nginxRunner
}

func (r *NginxRunner) StartNginx() *exec.Cmd {
	r.checkNginxConfig()
	r.listenSignals()
	cmd := exec.Command(r.NginxCommand[0], r.NginxCommand[1:]...)
	err := cmd.Start()
	if err != nil {
		Panicf("nginx exited with error:\n%v\n", err)
	}
	r.nginxProc = cmd.Process
	r.forwardSignals()
	r.reloadOnChange()
	r.rotateLogsOnFull()
	return cmd
}

func (r *NginxRunner) checkNginxConfig() *exec.Cmd {
	cmd := exec.Command("nginx", "-t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Panicf("nginx config validation has failed:\n%v\n%v\n", err, string(out))
	}
	fmt.Println("nginx config validated successfully")
	return cmd
}

func (r *NginxRunner) listenSignals() {
	signal.Notify(
		r.signalsChan,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT,
	)
}

func (r *NginxRunner) forwardSignals() {
	// Forward signals to nginx process
	go func() {
		for sig := range r.signalsChan {
			err := r.nginxProc.Signal(sig)
			if err != nil {
				Panicf("couldn't send signal to nginx process:\n%v\n", err)
			}
		}
	}()
}

func (r *NginxRunner) reloadOnChange() {
	go func() {
		for {
			<-r.changeChan
			r.reloadNginx()
		}
	}()
}

func (r *NginxRunner) rotateLogsOnFull() {
	go func() {
		for {
			<-r.logFullChan
			r.rotateLogsNginx()
		}
	}()
}

func (r *NginxRunner) reloadNginx() {
	r.checkNginxConfig()
	Stdoutf("conf directories change detected, sending reload signal to nginx")
	err := r.nginxProc.Signal(syscall.SIGHUP)
	if err != nil {
		Panicf("couldn't send SIGHUP (reload signal) to nginx process:\n%v\n", err)
	}
	Stdoutf("reload signal has been successfully sent to nginx")
}

func (r *NginxRunner) rotateLogsNginx() {
	Stdoutf("rotating nginx log")
	unixTime := time.Now().Unix()
	bakFile := fmt.Sprintf("%s.%d.bak", r.logFile, unixTime)
	err := os.Rename(r.logFile, bakFile)
	if err != nil {
		Panicf("couldn't move nginx log file:\n%v\n", err)
	}
	Stdoutf("sending reopen logs signal to nginx (stdout or stderr fd open() errors are ok)")
	err = r.nginxProc.Signal(syscall.SIGUSR1)
	if err != nil {
		Panicf("couldn't send SIGUSR1 (reopen logs signal) to nginx process:\n%v\n", err)
	}
	Stdoutf("reopen logs signal has been successfully sent to nginx")
}
