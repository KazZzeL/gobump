package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unsafe"
)

const (
	ERR_READ  = 2
	ERR_WRITE = 3
	ERR_PARSE = 4
	ERR_CMD   = 5
	ERR_GIT   = 6
)

var ErrCmd = fmt.Errorf("command error")

// cmd runs a subprocess; when verbose, echoes the command and streams stdout/stderr to out
// (intended for go get and -exec inside a preformatted block).
func cmd(name string, args ...string) error {
	return runCmd(name, args, true)
}

// cmdQuiet runs a subprocess without writing to out (e.g. go mod tidy before git commit).
func cmdQuiet(name string, args ...string) error {
	return runCmd(name, args, false)
}

func runCmd(name string, args []string, logOutput bool) error {
	if logOutput && config.Verbose {
		out.Println(name, strings.Join(args, " "))
	}
	c := exec.Command(name, args...)
	c.Env = os.Environ()
	if logOutput && config.Verbose {
		c.Stdout = out
		c.Stderr = out
	}
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

// cmdOutput runs a command and returns its stdout output.
func cmdOutput(cmd string, envs []string, args ...string) (string, error) {
	if config.Verbose {
		out.Println(cmd, strings.Join(args, " "))
	}
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	if len(envs) > 0 {
		c.Env = append(c.Env, envs...)
	}
	c.Stderr = out
	data, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(unsafe.String(unsafe.SliceData(data), len(data))), nil
}

func cmds(str string) error {
	parts := strings.Fields(str)
	if len(parts) == 0 {
		return fmt.Errorf("%w: no command", ErrCmd)
	}

	if len(parts) == 1 {
		return cmd(parts[0])
	}

	p1 := parts[0]
	p2 := parts[1:]
	return cmd(p1, p2...)
}
