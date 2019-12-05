package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/hashwing/goansible/model"
)

type CmdSession struct {
	sess            *exec.Cmd
	ctx             context.Context
	onceStdinCloser sync.Once
	stdin           io.WriteCloser
	output          *bytes.Buffer
}

// Start starts a remote process in the current session
func (s *CmdSession) Start(cmd string) error {
	var stdout bytes.Buffer
	sess := exec.CommandContext(s.ctx, "sh", "-c", cmd)
	sess.Stdout = &stdout
	sess.Stderr = &stdout
	s.output = &stdout
	s.sess = sess
	stdin, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	s.stdin = stdin
	return sess.Start()
}

// wait blocks until the remote process completes or is cancelled
func (s *CmdSession) Wait() error {
	return s.sess.Wait()
}

func (s *CmdSession) Output() string {
	return s.output.String()
}

// Stdin returns a pipe to the stdin of the remote process
func (s *CmdSession) Stdin() io.Writer {
	return s.stdin
}

// CloseStdin closes the stdin pipe of the remote process
func (s *CmdSession) CloseStdin() error {
	var err error
	s.onceStdinCloser.Do(func() {
		err = s.stdin.Close()
	})
	return err
}

// close closes the current session
func (s *CmdSession) Close() error {
	err := s.CloseStdin()
	if err != nil {
		return fmt.Errorf("failed to close stdin: %s", err)
	}
	return nil
}

func newCmdSession(ctx context.Context) *CmdSession {
	return &CmdSession{ctx: ctx}
}

type LocalCmd struct {
}

func ConnectCmd() model.Connection {
	return &LocalCmd{}
}

func (conn *LocalCmd) Close() error {
	return nil
}

func (conn *LocalCmd) Exec(ctx context.Context, withTerminal bool, fn model.ExecCallbackFunc) (string, error) {
	sess := newCmdSession(ctx)
	defer sess.Close()

	err, errGroup := fn(sess)
	if err != nil {
		return sess.Output(), fmt.Errorf("failed to start the command: %s", err)
	}

	// Wait for the session to finish running
	err = sess.Wait()
	if err != nil {
		// Check the async operation (if there is any) for the error
		// cause before returning
		err = fmt.Errorf("failed command: %s", err)
	}

	if errGroup != nil {
		asyncErr := errGroup.Wait()
		if asyncErr != nil {
			err = fmt.Errorf("%s: failed async operation: %s", err, asyncErr)
		}
	}

	if err != nil {
		return sess.Output(), err
	}

	// Make sure we always return some error when the command is cancelled
	return sess.Output(), ctx.Err()
}

func (conn *LocalCmd) CopyFile(ctx context.Context, src io.Reader, size int64, dest, mode string) error {
	modeInt, err := strconv.ParseInt(mode, 8, 32)
	if err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, os.FileMode(modeInt))
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, src)
	return err
}
