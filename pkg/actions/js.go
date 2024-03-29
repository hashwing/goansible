package actions

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/dop251/goja"

	"github.com/bitfield/script"
	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

var LibStr = ""

type JsAction string

func (a *JsAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	jsvm := goja.New()
	for k, v := range common.Vars(vars) {
		jsvm.Set(k, v)
	}
	newPipe := func() *script.Pipe {
		return script.NewPipe()
	}
	jsvm.Set("newPipe", newPipe)
	jsvm.Set("mustCompile", regexp.MustCompile)
	jsvm.Set("setVars", func(s string, v interface{}) {
		common.SetVar(s, v, vars)
	})
	jsvm.Set("getVars", func(s string) interface{} {
		v, _ := common.GetVar(s, vars)
		return v
	})
	jsvm.Set("parseTpl", func(v string) string {
		return common.ParseTplWithPanic(v, vars)
	})
	jsvm.Set("md5", MD5)
	jsvm.Set("FileMd5", FileMd5)
	jsvm.Set("readFile", func(filename string) string {
		data, _ := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, filename))
		return string(data)
	})

	//_, err := jsvm.RunString(fmt.Sprintf("function run(){%s}\n run()", *a))
	_, err := jsvm.RunString(LibStr + string(*a))

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
