package model

import (
	"bufio"
	"context"
	"io"

	"golang.org/x/sync/errgroup"
)

type Session interface {
	Start(cmd string, logFunc ...func(scanner *bufio.Scanner)) error
	Stdin() io.Writer
	CloseStdin() error
	Wait() error
	Close() error
	Output() string
}

type ExecCallbackFunc func(Session) (error, *errgroup.Group)

type Connection interface {
	Close() error
	Exec(context.Context, bool, ExecCallbackFunc) (string, error)
	CopyFile(ctx context.Context, src io.Reader, size int64, dest, mode string) error
	IsSudo() bool
}
