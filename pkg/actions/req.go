package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/imroc/req/v3"
)

type ReqAction struct {
	URL     string `yaml:"url"`
	Method  string `yaml:"method"`
	Timeout int64  `yaml:"timeout"`
	Body    string `yaml:"body"`
}

func (a *ReqAction) parse(vars *model.Vars) (*ReqAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	ra := &ReqAction{
		URL:     common.ParseTplWithPanic(a.URL, vars),
		Method:  common.ParseTplWithPanic(a.Method, vars),
		Timeout: a.Timeout,
	}
	if a.Timeout == 0 {
		ra.Timeout = 30
	}
	return ra, gerr
}

func (a *ReqAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	client := req.C().
		SetUserAgent("goansible").
		SetTimeout(time.Duration(parseAction.Timeout) * time.Second).
		DevMode()
	r, err := setURL(client, parseAction.Method, parseAction.URL)
	if err != nil {
		return "", err
	}
	resp := r.SetBody(parseAction.Body).Do()
	if resp.Err != nil {
		return "", resp.Err
	}
	if resp.IsError() {
		return "", errors.New(resp.String())
	}
	return resp.String(), nil
}

func setURL(client *req.Client, method, url string) (*req.Request, error) {
	switch method {
	case "post":
		return client.Post(url), nil
	case "get":
		return client.Get(url), nil
	case "put":
		return client.Put(url), nil
	case "patch":
		return client.Patch(url), nil
	case "delete":
		return client.Delete(url), nil
	}
	return nil, fmt.Errorf("not support method %s", method)
}
