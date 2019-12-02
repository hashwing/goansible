package playbook

import (
	"reflect"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

func Loop(loop interface{}, vars *model.Vars) []interface{} {
	switch reflect.TypeOf(loop).Kind() {
	case reflect.Slice:
		return loop.([]interface{})
	case reflect.String:
		v, res := common.GetVar(common.ParseTplWithPanic(loop.(string), vars), vars)
		if res {
			return v.([]interface{})
		}
	}
	return nil
}
