package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type FileAction struct {
	Kind  string `yaml:"kind"`
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
		Kind:  a.Kind,
	}, gerr
}

func walkDir(src, dest string, data *[]FileAction) error {
	fs, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.IsDir() {
			walkDir(src+"/"+f.Name(), dest+"/"+f.Name(), data)
		} else {
			*data = append(*data, FileAction{
				Src:  src + "/" + f.Name(),
				Dest: dest + "/" + f.Name(),
			})
		}
	}
	return nil
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
	var fpaths []FileAction
	if parseAction.Kind == "dir" {
		walkDir(parseAction.Src, parseAction.Dest, &fpaths)
	} else {
		fpaths = []FileAction{FileAction{
			Dest: parseAction.Dest,
			Src:  parseAction.Src,
		}}
	}
	for _, dfpath := range fpaths {
		f, err := os.Open(dfpath.Src)
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
		dest := dfpath.Dest
		if conn.IsSudo() {
			dest = "/tmp/goansible_file_" + uuid.NewString()
		}
		err = conn.CopyFile(ctx, f, stat.Size(), dest, mode)
		if err != nil {
			return "", fmt.Errorf("failed to copy file %s to %s error: %v", dfpath.Src, dfpath.Dest, err)
		}
		if conn.IsSudo() {
			output, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
				return sess.Start(
					fmt.Sprintf("mv %s %s", dest, dfpath.Dest),
				), nil
			})
			if err != nil {
				return output, fmt.Errorf(
					"failed to move the file to %s -> %s, %v",
					dest, dfpath.Dest, err,
				)
			}
		}

		if parseAction.Owner != "" && parseAction.Group != "" {
			output, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
				return sess.Start(
					fmt.Sprintf("chown %s:%s %s", parseAction.Owner, parseAction.Group, dfpath.Dest),
				), nil
			})
			if err != nil {
				return output, fmt.Errorf(
					"failed to set the file owner on %q to %s:%s: %s",
					dfpath.Dest, parseAction.Owner, parseAction.Group, err,
				)
			}
		}
	}

	return "", nil
}
