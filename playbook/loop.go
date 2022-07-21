package playbook

import (
	"reflect"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

func Loop(loop interface{}, vars *model.Vars) map[interface{}]interface{} {
	switch reflect.TypeOf(loop).Kind() {
	case reflect.Slice:
		res := make(map[interface{}]interface{})
		for i, item := range loop.([]interface{}) {
			res[i] = common.ParseTplWithPanic(item.(string), vars)
		}
		return res
	case reflect.String:
		v, res := common.GetVar(loop.(string), vars)
		if res {
			return Loop(v, vars)
		}
	case reflect.Map:
		res := make(map[interface{}]interface{})
		v := reflect.ValueOf(loop)
		for _, k := range v.MapKeys() {
			res[k.Interface()] = v.MapIndex(k).Interface()
		}
		return res
		// if v, ok := loop.(map[string]interface{}); ok {
		// 	for k, vv := range v {
		// 		res[k] = vv
		// 	}
		// 	return res
		// }
		// if v, ok := loop.(map[string]map[string]interface{}); ok {
		// 	for k, vv := range v {
		// 		res[k] = vv
		// 	}
		// 	return res
		// }
		// return loop.(map[interface{}]interface{})
	}
	return nil
}
