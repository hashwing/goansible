package actions

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type UntilAction struct {
	Port     int    `yaml:"port"`
	Shell    string `yaml:"shell"`
	Match    string `yaml:"match"`
	Timeout  int64  `yaml:"timeout"`
	Interval int64  `yaml:"interval"`
}

func (a *UntilAction) parse(vars *model.Vars) (*UntilAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()
	port, err := strconv.Atoi(common.ParseTplWithPanic(strconv.Itoa(a.Port), vars))
	if err != nil {
		return nil, err
	}
	timeout, err := strconv.Atoi(common.ParseTplWithPanic(strconv.Itoa(int(a.Timeout)), vars))
	if err != nil {
		return nil, err
	}
	inv, err := strconv.Atoi(common.ParseTplWithPanic(strconv.Itoa(int(a.Interval)), vars))
	if err != nil {
		return nil, err
	}
	return &UntilAction{
		Port:     port,
		Shell:    common.ParseTplWithPanic(a.Shell, vars),
		Match:    common.ParseTplWithPanic(a.Match, vars),
		Timeout:  int64(timeout),
		Interval: int64(inv),
	}, gerr
}

var grepPortShell = "ss -lnp|awk '{print $5}'|grep  -n :%d$"
var grepPortNetShell = "netstat -lnp|awk '{print $5}'|grep  -n :%d$"

func (a *UntilAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	newa, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	if newa.Port != 0 {
		newa.Shell = fmt.Sprintf(grepPortShell, newa.Port)
		newa.Match = ".+"
	}
	if newa.Timeout == 0 {
		newa.Timeout = 300
	}
	if newa.Interval == 0 {
		newa.Interval = 5
	}
	startTime := time.Now().Unix()
	for {
		currentTime := time.Now().Unix()
		if currentTime-startTime > a.Timeout {
			break
		}
		stdout, err := conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
			return sess.Start(newa.Shell), nil
		})
		if err != nil && newa.Port != 0 {
			stdout, err = conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
				return sess.Start(grepPortNetShell), nil
			})
		}
		if err != nil {
			//return stdout, err
		}
		isMatch, err := regexp.MatchString(newa.Match, stdout)
		if err != nil {
			return stdout, err
		}
		if isMatch {
			return stdout, nil
		}
		time.Sleep(time.Second * time.Duration(newa.Interval))
	}
	return "", errors.New("timeout")
}
