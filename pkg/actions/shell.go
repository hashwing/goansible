package actions

import (
	"context"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type ShellAction string

func (a *ShellAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	return conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
		comm, err := common.ParseTpl(string(*a), vars)
		if err != nil {
			return err, nil
		}
		return sess.Start(comm), nil
	})
}
