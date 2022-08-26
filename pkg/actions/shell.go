package actions

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type ShellAction string

func (a *ShellAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	comm, err := common.ParseTpl(string(*a), vars)
	if err != nil {
		return "", err
	}
	fmt.Println(comm)
	if conn.IsSudo() {
		fdata := []byte(comm)
		shellPath := "/tmp/goansible_shell_" + uuid.NewString()
		err = conn.CopyFile(ctx, bytes.NewReader(fdata), int64(len(fdata)), shellPath, "0644")
		if err != nil {
			return "", err
		}
		defer func() {
			conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
				return sess.Start("rm -f " + shellPath), nil
			})
		}()
		comm = "sh " + shellPath
	}

	return conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
		// comm, err := common.ParseTpl(string(*a), vars)
		// if err != nil {
		// 	return err, nil
		// }
		return sess.Start(comm), nil
	})
}
