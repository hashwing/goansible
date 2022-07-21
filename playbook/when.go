package playbook

import (
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"

	"github.com/hashwing/goansible/pkg/expr"
)

func When(s string, vars *model.Vars) bool {
	// res, err := expr.Run(s, vars)
	// if err != nil {
	// 	panic(err)
	// }
	// return res.Bool()
	res, err := expr.Eval(s, vars)
	if err != nil {
		panic(err)
	}
	return res.(bool)
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
