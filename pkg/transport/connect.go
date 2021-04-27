package transport

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path/filepath"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type connection struct {
	client *ssh.Client
}

func Connect(user, passwd, pkFile, addr string) (model.Connection, error) {
	auth := []ssh.AuthMethod{}
	if passwd != "" {
		keyboardInteractiveChallenge := func(
			user,
			instruction string,
			questions []string,
			echos []bool,
		) (answers []string, err error) {
			if len(questions) == 0 {
				return []string{}, nil
			}
			return []string{passwd}, nil
		}
		auth = append(auth, ssh.Password(passwd))
		auth = append(auth, ssh.KeyboardInteractive(keyboardInteractiveChallenge))
	} else {
		pkData, err := ioutil.ReadFile(pkFile)
		if err != nil {
			return nil, err
		}
		pk, err := ssh.ParsePrivateKey([]byte(pkData))
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(pk))
	}

	config := ssh.Config{
		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
	}

	clientConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: time.Duration(1) * time.Minute,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}
	return &connection{client}, nil
}

func (conn *connection) Close() error {
	return conn.client.Close()
}

func (conn *connection) Exec(ctx context.Context, withTerminal bool, fn model.ExecCallbackFunc) (string, error) {
	sess, err := newSession(ctx, conn.client, withTerminal, fn)
	if err != nil {
		return "", fmt.Errorf("failed to create new session: %s", err)
	}
	// TODO: Log error
	defer sess.Close()

	err, errGroup := fn(sess)
	if err != nil {
		return sess.Output(), fmt.Errorf("failed to start the ssh command: %s", err)
	}

	// Wait for the session to finish running
	err = sess.Wait()
	if err != nil {
		// Check the async operation (if there is any) for the error
		// cause before returning
		err = fmt.Errorf("failed ssh command: %s", err)
	}

	if errGroup != nil {
		asyncErr := errGroup.Wait()
		if asyncErr != nil {
			err = fmt.Errorf("%s: failed async ssh operation: %s", err, asyncErr)
		}
	}

	if err != nil {
		return sess.Output(), err
	}

	// Make sure we always return some error when the command is cancelled
	return sess.Output(), ctx.Err()
}

func (conn *connection) CopyFile(ctx context.Context, src io.Reader, size int64, dest, mode string) error {
	sftpClient, err := sftp.NewClient(conn.client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	err = sftpClient.MkdirAll(filepath.Dir(dest))
	if err != nil {
		return err
	}
	fd, err := sftpClient.Create(dest)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = io.Copy(fd, src)
	// buf := make([]byte, 1024)
	// for {
	// 	n, _ := src.Read(buf)
	// 	if n == 0 {
	// 		break
	// 	}
	// 	fd.Write(buf[:n])
	// }
	return err
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
