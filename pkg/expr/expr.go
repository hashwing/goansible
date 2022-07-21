package expr

import (
	"reflect"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"

	"github.com/antonmedv/expr"
)

func Eval(s string, vars *model.Vars) (interface{}, error) {
	envs := common.Vars(vars)
	envs["getOne"] = getOne
	return expr.Eval(s, envs)
}

func getOne(p interface{}) interface{} {
	v := reflect.ValueOf(p)
	for _, k := range v.MapKeys() {
		return v.MapIndex(k).Interface()
	}
	return ""
}
