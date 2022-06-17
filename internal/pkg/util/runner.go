package util

import (
	"bufio"
	"fmt"
	"k8s.io/klog/v2"
	"os/exec"
	"runtime"
)

type RealRunner struct {
	cmd          string // the tool name
	wd           string // the working directory
	addShellFlag bool
	history      []string
	debug        bool
}

func NewRealRunner(cmd string, wd string, debug bool) *RealRunner {
	return &RealRunner{cmd: cmd, wd: wd, debug: debug}
}

func (r *RealRunner) SetWorkingDir(wd string) {
	r.wd = wd
}

func (r *RealRunner) AddShellFlag() {
	r.addShellFlag = true
}

func (r *RealRunner) History() []string {
	return r.history
}

func (r *RealRunner) NoShellFlag() {
	r.addShellFlag = false
}

// run a docker command, return the stdout and stderr
func (r RealRunner) Run(subcmd []string) (string, string, error) {
	var outStr = ""
	// on windows we need to add the run command
	if runtime.GOOS == "windows" && r.addShellFlag {
		subcmd = append(subcmd, "--shell")
		subcmd = append(subcmd, "cmd")
	}

	cmd := exec.Command(r.cmd, subcmd...)

	cmdString := r.cmd
	for _, v := range subcmd {
		cmdString = cmdString + " " + v
	}
	if r.debug {
		fmt.Printf("--> Running: %s\n", cmdString)
	}
	r.history = append(r.history, cmdString)
	//cmd.Dir = filepath.Join(r.wd)

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		klog.Fatal("Couldn't connect to command's stderr")
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		klog.Fatal("Couldn't connect to command's stdout")
	}
	_ = bufio.NewReader(errPipe)
	outReader := bufio.NewReader(outPipe)

	if r.debug {
		fmt.Printf("--> starting command: %s\n", cmdString)
	}

	// start the command and filter the output
	if err = cmd.Start(); err != nil {
		return "", "", err
	}
	outScanner := bufio.NewScanner(outReader)
	for outScanner.Scan() {
		outStr += outScanner.Text() + "\n"
		if r.debug {
			fmt.Println(outScanner.Text())
		}
	}
	err = cmd.Wait()
	return outStr, "", err

}
