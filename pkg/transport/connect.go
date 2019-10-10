package transport

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"context"

	"github.com/hashwing/goansible/model"
	"golang.org/x/crypto/ssh"
)

type connection struct {
	client *ssh.Client
}

func Connect(user, passwd, pkFile, addr string) (model.Connection, error) {
	var am ssh.AuthMethod
	if passwd != "" {
		am = ssh.Password(passwd)
	} else {
		pkData, err := ioutil.ReadFile(pkFile)
		if err != nil {
			return nil, err
		}
		pk, err := ssh.ParsePrivateKey([]byte(pkData))
		if err != nil {
			return nil, err
		}
		am = ssh.PublicKeys(pk)
	}
	auth := []ssh.AuthMethod{am}
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
	sess, err := newSession(ctx, conn.client, withTerminal)
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
