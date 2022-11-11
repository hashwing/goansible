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
	URL             string            `yaml:"url"`
	Method          string            `yaml:"method"`
	Timeout         int64             `yaml:"timeout"`
	Body            string            `yaml:"body"`
	Headers         map[string]string `yaml:"headers"`
	BaseAuth        ReqBaseAuth       `yaml:"baseAuth"`
	BearerAuthToken string            `yaml:"bearerToken"`
	DownloadFile    string            `yaml:"downloadFile"`
	UploadFiles     map[string]string `yaml:"uploadFile"`
}

type ReqBaseAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (a *ReqAction) parse(vars *model.Vars) (*ReqAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	ra := &ReqAction{
		URL:             common.ParseTplWithPanic(a.URL, vars),
		Method:          common.ParseTplWithPanic(a.Method, vars),
		Timeout:         a.Timeout,
		Body:            common.ParseTplWithPanic(a.Body, vars),
		Headers:         make(map[string]string),
		BearerAuthToken: common.ParseTplWithPanic(a.BearerAuthToken, vars),
		DownloadFile:    common.ParseTplWithPanic(a.DownloadFile, vars),
		UploadFiles:     make(map[string]string),
	}
	if a.Timeout == 0 {
		ra.Timeout = 30
	}
	if a.Headers != nil {
		for k, v := range a.Headers {
			ra.Headers[k] = common.ParseTplWithPanic(v, vars)
		}
	}
	if a.BaseAuth.Username != "" || a.BaseAuth.Password != "" {
		ra.BaseAuth = ReqBaseAuth{
			Username: common.ParseTplWithPanic(a.BaseAuth.Username, vars),
			Password: common.ParseTplWithPanic(a.BaseAuth.Password, vars),
		}
	}
	if a.UploadFiles != nil {
		for k, v := range a.Headers {
			ra.UploadFiles[k] = common.ParseTplWithPanic(v, vars)
		}
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
	if parseAction.BaseAuth.Username != "" {
		r = r.SetBasicAuth(parseAction.BaseAuth.Username, parseAction.BaseAuth.Password)
	}
	if parseAction.BearerAuthToken != "" {
		r.SetBearerAuthToken(parseAction.BearerAuthToken)
	}
	if len(parseAction.UploadFiles) > 0 {
		r.SetFiles(parseAction.UploadFiles)
	}
	resp := r.SetHeaders(parseAction.Headers).
		SetBody(parseAction.Body).Do()
	if resp.Err != nil {
		return "", resp.Err
	}
	if resp.IsError() {
		return "", errors.New(resp.String())
	}
	return resp.String(), nil
}

func setURL(client *req.Client, method, url string) (*req.Request, error) {
	if method == "" {
		method = "GET"
	}
	switch method {
	case "POST":
		return client.Post(url), nil
	case "GET":
		return client.Get(url), nil
	case "PUT":
		return client.Put(url), nil
	case "PATCH":
		return client.Patch(url), nil
	case "DELETE":
		return client.Delete(url), nil
	}
	return nil, fmt.Errorf("not support method %s", method)
}
