package actions

import (
	"context"
	"path/filepath"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/pkg/ftp"
)

type FTPAction struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	User   string `yaml:"username"`
	PWD    string `yaml:"password"`
	Dir    string `yaml:"dir"`
	Tmp    string `yaml:"tmp" `
	Action string `yaml:"action"`
	Remote string `yaml:"remote"`
	Local  string `yaml:"local"`
	Res    string `yaml:"res"`
}

func (a *FTPAction) parse(vars *model.Vars) (*FTPAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()
	ra := &FTPAction{
		Host:   common.ParseTplWithPanic(a.Host, vars),
		Port:   common.ParseTplWithPanic(a.Port, vars),
		User:   common.ParseTplWithPanic(a.User, vars),
		PWD:    common.ParseTplWithPanic(a.PWD, vars),
		Dir:    common.ParseTplWithPanic(a.Dir, vars),
		Tmp:    common.ParseTplWithPanic(a.Tmp, vars),
		Action: common.ParseTplWithPanic(a.Action, vars),
		Remote: common.ParseTplWithPanic(a.Remote, vars),
		Local:  common.ParseTplWithPanic(a.Local, vars),
		Res:    a.Res,
	}
	return ra, gerr
}

func (a *FTPAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	cli := ftp.NewClient(&ftp.Config{
		Host: parseAction.Host,
		Port: parseAction.Port,
		User: parseAction.User,
		PWD:  parseAction.PWD,
		Dir:  parseAction.Dir,
		Tmp:  parseAction.Tmp,
	})
	switch parseAction.Action {
	case "list":
		list, err := cli.FindList(parseAction.Dir)
		if err != nil {
			return "", err
		}
		if parseAction.Res != "" {
			common.SetVar(parseAction.Res, list, vars)
		}
	case "download":
		fpath := parseAction.Local
		if !filepath.IsAbs(parseAction.Local) {
			fpath = filepath.Join(conf.PlaybookFolder, parseAction.Local)
		}
		err := cli.FDownload(parseAction.Remote, fpath)
		if err != nil {
			return "", err
		}
	case "upload":
		fpath := parseAction.Local
		if !filepath.IsAbs(parseAction.Local) {
			fpath = filepath.Join(conf.PlaybookFolder, parseAction.Local)
		}
		err := cli.FUpload(fpath, parseAction.Remote)
		if err != nil {
			return "", err
		}
	case "delete":
		err := cli.Delete(parseAction.Remote)
		if err != nil {
			return "", err
		}
	case "tranfinish":
		res, err := cli.TranFinish(parseAction.Remote)
		if err != nil {
			return "", err
		}
		if parseAction.Res != "" {
			common.SetVar(parseAction.Res, res, vars)
		}
	}
	return "", nil
}
