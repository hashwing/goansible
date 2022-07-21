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

func (a *FileAction) parse(vars *model.Vars) (*FileAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	return &FileAction{
		Src:   common.ParseTplWithPanic(a.Src, vars),
		Dest:  common.ParseTplWithPanic(a.Dest, vars),
		Owner: common.ParseTplWithPanic(a.Owner, vars),
		Group: common.ParseTplWithPanic(a.Group, vars),
		Mode:  common.ParseTplWithPanic(a.Mode, vars),
	}, gerr
}

func (a *FileAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	fpath := parseAction.Src
	if !filepath.IsAbs(fpath) {
		fpath = filepath.Join(conf.PlaybookFolder, parseAction.Src)
	}
	f, err := os.Open(fpath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %s", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get source file info: %s", err)
	}

	mode := parseAction.Mode
	if mode == "" {
		mode = "0644"
	}
	err = conn.CopyFile(ctx, f, stat.Size(), parseAction.Dest, mode)
	if err != nil {
		return "", fmt.Errorf("failed to copy file %s to %s error: %v", parseAction.Src, parseAction.Dest, err)
	}

	if parseAction.Owner != "" && parseAction.Group != "" {
		output, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(
				fmt.Sprintf("chown %s:%s %s", parseAction.Owner, parseAction.Group, parseAction.Dest),
			), nil
		})
		if err != nil {
			return output, fmt.Errorf(
				"failed to set the file owner on %q to %s:%s: %s",
				parseAction.Dest, parseAction.Owner, parseAction.Group, err,
			)
		}
	}

	return "", nil
}
