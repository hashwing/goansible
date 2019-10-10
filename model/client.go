package model

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"
)

type Session interface {
	Start(cmd string) error
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
}
