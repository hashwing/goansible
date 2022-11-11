package actions

import (
	"context"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	//"github.com/hashwing/goansible/pkg/ssh"
)

type SshAction struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	User   string `yaml:"username"`
	PWD    string `yaml:"password"`
	Prompt string `yaml:"prompt"`
	Stdout string `yaml:"stdout"`
}

func (a *SshAction) parse(vars *model.Vars) (*SshAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()
	ra := &SshAction{
		Host:   common.ParseTplWithPanic(a.Host, vars),
		Port:   common.ParseTplWithPanic(a.Port, vars),
		User:   common.ParseTplWithPanic(a.User, vars),
		PWD:    common.ParseTplWithPanic(a.PWD, vars),
		Stdout: a.Stdout,
		Prompt: a.Prompt,
	}
	return ra, gerr
}

func (a *SshAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	_ = parseAction
	//ssh.Connect(parseAction.Host)
	return "", nil
}
