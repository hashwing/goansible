package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hashwing/goansible/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Session is a wrapper around ssh.Session
type Session struct {
	sshSess               *ssh.Session
	onceStdinCloser       sync.Once
	stdin                 io.WriteCloser
	output                *bytes.Buffer
	sigintHandlerQuitChan chan struct{}
}

// Start starts a remote process in the current session
func (s *Session) Start(cmd string) error {
	err := s.sshSess.Start(cmd)
	return err
}

// wait blocks until the remote process completes or is cancelled
func (s *Session) Wait() error {
	return s.sshSess.Wait()
}

func (s *Session) Output() string {
	return s.output.String()
}

// Stdin returns a pipe to the stdin of the remote process
func (s *Session) Stdin() io.Writer {
	return s.stdin
}

// CloseStdin closes the stdin pipe of the remote process
func (s *Session) CloseStdin() error {
	var err error
	s.onceStdinCloser.Do(func() {
		err = s.stdin.Close()
	})
	return err
}

// close closes the current session
func (s *Session) Close() error {
	if s.sigintHandlerQuitChan != nil {
		close(s.sigintHandlerQuitChan)
	}

	err := s.CloseStdin()
	if err != nil {
		return fmt.Errorf("failed to close stdin: %s", err)
	}

	err = s.sshSess.Close()
	if err != nil {
		return fmt.Errorf("failed to close session: %s", err)
	}

	return nil
}

// newSession creates a new session
func newSession(ctx context.Context, client *ssh.Client, withTerminal bool) (model.Session, error) {
	sshSess, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to initialise session: %s", err)
	}

	var b bytes.Buffer
	sshSess.Stdout = &b
	stdin, err := sshSess.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get the session stdin pipe: %s", err)
	}

	if withTerminal {
		err = sshSess.RequestPty("xterm", 80, 40,
			ssh.TerminalModes{
				ssh.ECHO:          0,     // disable echoing
				ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
				ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to setup the pseudo terminal: %s", err)
		}
	}

	// If requested, send SIGINT to the remote process and close the session
	quitChan := make(chan struct{})
	sess := Session{sshSess: sshSess, stdin: stdin, output: &b, sigintHandlerQuitChan: quitChan}
	go func() {
		select {
		case <-ctx.Done():
			if withTerminal {
				_, err := stdin.Write([]byte("\x03"))
				if err != nil && err != io.EOF {
					log.Warnf("Failed to send SIGINT to the remote process: %s", err)
				}
			}
			err := sess.CloseStdin()
			if err != nil {
				log.Warnf("Failed to close session stdin: %s", err)
			}
			err = ctx.Err()
			if err == context.DeadlineExceeded {
				log.Warnf("Context deadline exceeeded on server: %s", client.RemoteAddr())
			}
		case <-quitChan:
			// Stop the signal handler when the task completes
		}
	}()

	return &sess, nil
}
