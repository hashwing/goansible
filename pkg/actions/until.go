package actions

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashwing/goansible/model"
	"golang.org/x/sync/errgroup"
)

type UntilAction struct {
	Port     int    `yaml:"port"`
	Shell    string `yaml:"shell"`
	Match    string `yaml:"match"`
	Timeout  int64  `yaml:"timeout"`
	Interval int64  `yaml:"interval"`
}

var grepPortShell = "ss -lnp|awk '{print $5}'|grep  -n :%d$"

func (a *UntilAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	if a.Port != 0 {
		a.Shell = fmt.Sprintf(grepPortShell, a.Port)
		a.Match = ".+"
	}
	if a.Timeout == 0 {
		a.Timeout = 300
	}
	if a.Interval == 0 {
		a.Interval = 5
	}
	startTime := time.Now().Unix()
	for {
		currentTime := time.Now().Unix()
		if currentTime-startTime > a.Timeout {
			break
		}
		stdout, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(a.Shell), nil
		})
		isMatch, err := regexp.MatchString(a.Match, stdout)
		if err != nil {
			return stdout, err
		}
		if isMatch {
			return stdout, nil
		}
		time.Sleep(time.Second * time.Duration(a.Interval))
	}
	return "", errors.New("timeout")
}
