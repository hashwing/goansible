package transport

import (
	"context"
	"testing"

	"github.com/hashwing/goansible/model"
	"golang.org/x/sync/errgroup"
)

var conn model.Connection

func TestConnect(t *testing.T) {
	var err error
	conn, err = Connect("root", "sunrunvas", "", "188.8.2.130:22")
	if err != nil {
		t.Error(err)
		return
	}
	output, err := conn.Exec(context.Background(), true, func(s model.Session) (error, *errgroup.Group) {
		return s.Start("ddddd"), nil
	})
	if err != nil {
		t.Error(err, output)
		return
	}

}

func TestSess(t *testing.T) {
	output, err := conn.Exec(context.Background(), true, func(s model.Session) (error, *errgroup.Group) {
		return s.Start("ls"), nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(output)
}
