package actions

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/dop251/goja"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

type JsAction string

func (a *JsAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	jsvm := goja.New()
	for k, v := range common.Vars(vars) {
		jsvm.Set(k, v)
	}
	_, err := jsvm.RunString(fmt.Sprintf("function run(){%s}\n run()", *a))

	return "", err
}

type JsFileAction string

func (a *JsFileAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, string(*a)))
	if err != nil {
		return "", err
	}
	ja := JsAction(string(data))
	return ja.Run(ctx, conn, conf, vars)
}
