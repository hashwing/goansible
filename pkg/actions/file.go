package actions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashwing/goansible/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type FileAction struct {
	Src   string `yaml:"src"`
	Dest  string `yaml:"dest"`
	Owner string `yaml:"owner"`
	Group string `yaml:"group"`
	Mode  string `yaml:"mode"`
}

// Copies the contents of src to dest on a remote host
func copyFile(sess model.Session, src io.Reader, size int64, dest, mode string) error {
	// Instruct the remote scp process that we want to bail out immediately
	defer func() {
		err := sess.CloseStdin()
		if err != nil {
			log.Warnf("Failed to close session stdin: %s", err)
		}
	}()

	_, err := fmt.Fprintln(sess.Stdin(), "C"+mode, size, filepath.Base(dest))
	if err != nil {
		return fmt.Errorf("failed to create remote file: %s", err)
	}

	_, err = io.Copy(sess.Stdin(), src)
	if err != nil {
		return fmt.Errorf("failed to write remote file contents: %s", err)
	}

	_, err = fmt.Fprint(sess.Stdin(), "\x00")
	if err != nil {
		return fmt.Errorf("failed to close remote file: %s", err)
	}

	return nil
}

func (a *FileAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	output, err := conn.Exec(ctx, false, func(sess model.Session) (error, *errgroup.Group) {
		f, err := os.Open(filepath.Join(conf.PlaybookFolder, a.Src))
		if err != nil {
			return fmt.Errorf("failed to open source file: %s", err), nil
		}
		stat, err := f.Stat()
		if err != nil {
			return fmt.Errorf("failed to get source file info: %s", err), nil
		}

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
				f,
				stat.Size(),
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
