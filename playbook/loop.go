package playbook

import (
	"reflect"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

func Loop(loop interface{}, vars *model.Vars) []interface{} {
	switch reflect.TypeOf(loop).Kind() {
	case reflect.Slice:
		res:=make([]interface{},0)
		for _,item:=range loop.([]interface{}){
			res=append(res,common.ParseTplWithPanic(item.(string)))
		}
		return res
	case reflect.String:
		v, res := common.GetVar(loop.(string), vars)
		if res {
			return v.([]interface{})
		}
	}
	return nil
}
