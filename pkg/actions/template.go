package actions

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type TemplateAction struct {
	Src   string `yaml:"src"`
	Dest  string `yaml:"dest"`
	Owner string `yaml:"owner"`
	Group string `yaml:"group"`
	Mode  string `yaml:"mode"`
}

func (a *TemplateAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	output, err := conn.Exec(ctx, false, func(sess model.Session) (error, *errgroup.Group) {
		tpl, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, a.Src))
		if err != nil {
			return fmt.Errorf("failed to open source file: %s", err), nil
		}
		data, err := common.ParseTpl(string(tpl), vars)
		if err != nil {
			return fmt.Errorf("failed to parse source file: %s", err), nil
		}
		buf := bytes.NewBufferString(data)

		dir, _ := filepath.Split(a.Dest)

		// Start scp receiver on the remote host
		err = sess.Start("scp -qt " + dir)
		if err != nil {
			return fmt.Errorf("failed to start scp receiver: %s", err), nil
		}

		mode := a.Mode
		if mode == "" {
			mode = "0644"
		}
		var g errgroup.Group
		g.Go(func() error {
			err := copyFile(
				sess,
				buf,
				int64(len(data)),
				a.Dest,
				mode,
			)
			return err
		})
		return nil, &g
	})
	if err != nil {
		return output, fmt.Errorf("failed to copy file %q: %s", a.Src, err)
	}

	if a.Owner != "" && a.Group != "" {
		output, err = conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(
				fmt.Sprintf("chown %s:%s %s", a.Owner, a.Group, a.Dest),
			), nil
		})
		if err != nil {
			return output, fmt.Errorf(
				"failed to set the file owner on %q to %s:%s: %s",
				a.Dest, a.Owner, a.Group, err,
			)
		}
	}

	return output, nil
}
