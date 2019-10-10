package actions

import (
	"context"
	"errors"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

type SetfaceAction string

func (a *SetfaceAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	params := strings.Split(string(*a), "=")
	if len(params) > 1 {
		data, err := common.ParseTpl(strings.Join(params[1:], "="), vars)
		if err != nil {
			return "", err
		}
		common.SetVar(params[0], data, vars)
		return "", nil
	}

	return "", errors.New("set_face not enough parameters")
}
