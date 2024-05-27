package main

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"syscall"
)

type Shell struct {
	Proc   *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
	Cancel context.CancelFunc
}

// from "github.com/kylefeng28/go-shell"
func NewShell(command string) (Shell, error) {
	var err error

	shell := Shell{}

	ctx, cancel := context.WithCancel(context.Background())
	shell.Cancel = cancel

	shell.Proc = exec.CommandContext(ctx, command)
	shell.Proc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	if shell.Stdin, err = shell.Proc.StdinPipe(); err != nil {
		return shell, errors.New("could not get a pipe to stdin")
	}
	if shell.Stdout, err = shell.Proc.StdoutPipe(); err != nil {
		return shell, errors.New("could not get a pipe to stdout")
	}
	if shell.Stderr, err = shell.Proc.StderrPipe(); err != nil {
		return shell, errors.New("could not get a pipe to stderr")
	}

	if err = shell.Proc.Start(); err != nil {
		return shell, errors.New("could not start process")
	}

	return shell, nil
}

func (shell Shell) Close() error {
	shell.Stdout.Close()
	shell.Stderr.Close()
	shell.Cancel()
	pgid, _ := syscall.Getpgid(shell.Proc.Process.Pid)
	err := syscall.Kill(-pgid, syscall.SIGKILL)
	if err != nil {
		return err
	}
	// shell.Proc.Process.Kill()
	return shell.Proc.Wait()
}
