package actions

import (
	"context"
	"regexp"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

type RegexpAction struct {
	Src string `yaml:"src"`
	Exp string `yaml:"exp"`
	Dst string `yaml:"dst"`
}

func (a *RegexpAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	vRegexp := regexp.MustCompile(a.Exp)
	vParams := vRegexp.FindStringSubmatch(a.Src)
	common.SetVar(a.Dst, vParams[1:], vars)
	return "", nil
}
