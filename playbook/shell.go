package playbook

import (
	"bufio"
	"context"
	"fmt"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/termutil"
	"golang.org/x/sync/errgroup"
)

//RunShell run shell command
func RunShell(cfg model.Config, inv model.Inventory, name, shell string) error {
	gs, err := inv.Groups()
	if err != nil {
		return err
	}
	//fmt.Println(gs)
	for _, h := range gs[name].Hosts {
		conn, err := connect(h)
		if err != nil {
			return err
		}
		termutil.Printf("%s:", h.Name)
		_, err = conn.Exec(context.TODO(), true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(shell, func(scanner *bufio.Scanner) {
				for scanner.Scan() {
					fmt.Println(scanner.Text())
				}
			}), nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}
