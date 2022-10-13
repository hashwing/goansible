package actions

import (
	"context"
	"errors"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/pkg/expr"
)

type SetfaceAction string

func (a *SetfaceAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	params := strings.Split(string(*a), "=")
	if len(params) > 1 {
		d, err := common.ParseTpl(strings.Join(params[1:], "="), vars)
		if err != nil {
			return "", err
		}
		data, err := expr.Eval(d, vars)
		if err != nil {
			return "", err
		}
		common.SetVar(params[0], data, vars)
		return "", nil
	}

	return "", errors.New("set_face not enough parameters")
}
