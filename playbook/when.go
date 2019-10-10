package playbook

import (
	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

func When(s string, vars *model.Vars) bool {
	if v, ok := common.GetVar(s, vars); ok {
		return v.(bool)
	}
	return false
}
