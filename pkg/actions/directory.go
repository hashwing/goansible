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

func (a *DirectoryAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	dest, err := common.ParseTpl(a.Path, vars)
	if err != nil {
		return "", err
	}
	mko, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
		return sess.Start("mkdir " + dest), nil
	})
	if err != nil {
		return mko, err
	}
	if a.Mode != "" {
		chmodo, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			comm := fmt.Sprintf("chmod %s %s", a.Mode, dest)
			if err != nil {
				return err, nil
			}
			return sess.Start(comm), nil
		})
		if err != nil {
			return chmodo, err
		}
	}

	if a.Owner != "" && a.Group != "" {
		chowno, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			comm := fmt.Sprintf("chown %s:%s %s", a.Owner, a.Group, dest)
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
