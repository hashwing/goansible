package playbook

import (
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

func When(s string, vars *model.Vars) bool {
	not := false
	if strings.HasPrefix(s, "!") {
		not = true
		s = strings.TrimPrefix(s, "!")
	}

	if v, ok := common.GetVar(s, vars); ok {
		if not {
			return !v.(bool)
		}
		return v.(bool)
	}
	if not {
		return true
	}
	return false
}
