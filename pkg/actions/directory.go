package actions

import (
	"context"
	"fmt"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type DirectoryAction struct {
	Path  string `json:"path"`
	Owner string `yaml:"owner"`
	Group string `yaml:"group"`
	Mode  string `yaml:"mode"`
}

func (a *DirectoryAction) parse(vars *model.Vars) (*DirectoryAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	return &DirectoryAction{
		Path:  common.ParseTplWithPanic(a.Path, vars),
		Owner: common.ParseTplWithPanic(a.Owner, vars),
		Group: common.ParseTplWithPanic(a.Group, vars),
		Mode:  common.ParseTplWithPanic(a.Mode, vars),
	}, gerr
}

func (a *DirectoryAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	newAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	mko, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
		return sess.Start("mkdir -p " + newAction.Path), nil
	})
	if err != nil {
		return mko, err
	}
	if newAction.Mode != "" {
		chmodo, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			comm := fmt.Sprintf("chmod %s %s", newAction.Mode, newAction.Path)
			if err != nil {
				return err, nil
			}
			return sess.Start(comm), nil
		})
		if err != nil {
			return chmodo, err
		}
	}

	if newAction.Owner != "" && newAction.Group != "" {
		chowno, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			comm := fmt.Sprintf("chown %s:%s %s", newAction.Owner, newAction.Group, newAction.Path)
			if err != nil {
				return err, nil
			}
			return sess.Start(comm), nil
		})
		if err != nil {
			return chowno, err
		}
	}
	return "", nil
}
