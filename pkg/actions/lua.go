package actions

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/bitfield/script"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

type LuaAction string

func (a *LuaAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	luavm := lua.NewState()
	defer luavm.Close()
	for k, v := range common.Vars(vars) {
		luavm.SetGlobal(k, luar.New(luavm, v))
	}
	newPipe := func() *script.Pipe {
		return script.NewPipe()
	}
	luavm.SetGlobal("newPipe", luar.New(luavm, newPipe))
	luavm.SetGlobal("mustCompile", luar.New(luavm, regexp.MustCompile))
	err := luavm.DoString(string(*a))
	return "", err
}

type LuaFileAction string

func (a *LuaFileAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, string(*a)))
	if err != nil {
		return "", err
	}
	ja := JsAction(string(data))
	return ja.Run(ctx, conn, conf, vars)
}
