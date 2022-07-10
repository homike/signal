package signal

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
)

type Signal struct {
	signalTerm func()
	signalHUP  func()

	dup      bool
	execDir  string
	execFile string
}

func NewSignal(opts ...Option) *Signal {
	arg0, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err)
	}
	absExecFile, err := filepath.Abs(arg0)
	if err != nil {
		panic(err)
	}
	execDir, execFile := filepath.Split(absExecFile)
	sig := &Signal{
		execDir:  execDir,
		execFile: execFile,
	}
	for _, opt := range opts {
		opt(sig)
	}
	return sig
}

func (this *Signal) Handle() {
	fpid := this.pidfile()
	err := ioutil.WriteFile(fpid, []byte(fmt.Sprintf("%d", os.Getpid())), 0600)
	if err == nil {
		defer os.Remove(fpid)
	}

	// 当进程因未捕获的异常导致崩溃时, 将标准错误堆栈记录下来
	if this.dup {
		fout := fmt.Sprintf("%s/%s.%d", this.execDir, this.execFile, os.Getpid())
		logFile, err := os.OpenFile(fout, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err == nil {
			defer os.Remove(fout)
			//syscall.Dup2(int(logFile.Fd()), int(os.Stdout.Fd()))
			_ = syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
			defer logFile.Close()
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s, ok := <-c
		if !ok {
			return
		}
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			this.safetyCall(this.signalTerm)
			return
		case syscall.SIGHUP:
			this.safetyCall(this.signalHUP)
		case syscall.SIGABRT:
			fmt.Println("syscall.SIGABRT")
		}
	}
}

func (this *Signal) pidfile() string {
	return path.Join(this.execDir, this.execFile+".pid")
}

func (this *Signal) safetyCall(call func()) {
	if call == nil {
		return
	}
	//defer

	call()
}

func Kill(s syscall.Signal) error {
	sig := NewSignal()
	pid, err := ioutil.ReadFile(sig.pidfile())
	if err != nil {
		return err
	}
	return exec.Command("kill", fmt.Sprintf("-%d", s), string(pid)).Run()
}
