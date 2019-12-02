package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type FileAction struct {
	Src   string `yaml:"src"`
	Dest  string `yaml:"dest"`
	Owner string `yaml:"owner"`
	Group string `yaml:"group"`
	Mode  string `yaml:"mode"`
}

func (a *FileAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	src, err := common.ParseTpl(a.Src, vars)
	if err != nil {
		return "", err
	}
	dest, err := common.ParseTpl(a.Dest, vars)
	if err != nil {
		return "", err
	}
	f, err := os.Open(filepath.Join(conf.PlaybookFolder, src))
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %s", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get source file info: %s", err)
	}

	mode := a.Mode
	if mode == "" {
		mode = "0644"
	}
	err = conn.CopyFile(ctx, f, stat.Size(), dest, mode)
	if err != nil {
		return "", fmt.Errorf("failed to copy file %q: %s", src, err)
	}

	if a.Owner != "" && a.Group != "" {
		output, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(
				fmt.Sprintf("chown %s:%s %s", a.Owner, a.Group, dest),
			), nil
		})
		if err != nil {
			return output, fmt.Errorf(
				"failed to set the file owner on %q to %s:%s: %s",
				dest, a.Owner, a.Group, err,
			)
		}
	}

	return "", nil
}
