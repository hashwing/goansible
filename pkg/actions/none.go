package actions

import (
	"context"

	"github.com/hashwing/goansible/model"
)

type NoneAction string

func (a *NoneAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	return "", nil
}
