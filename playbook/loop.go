package playbook

import (
	"reflect"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

func Loop(loop interface{}, vars *model.Vars) []LoopRes {
	switch reflect.TypeOf(loop).Kind() {
	case reflect.Slice:
		res := make([]LoopRes, 0)
		for i, item := range loop.([]interface{}) {
			res = append(res, LoopRes{
				Item:    common.ParseTplWithPanic(item.(string), vars),
				ItemKey: i,
			})
		}
		return res
	case reflect.String:
		loopv, res := common.GetVar(loop.(string), vars)
		if res {
			res := make([]LoopRes, 0)
			switch reflect.TypeOf(loopv).Kind() {
			case reflect.Slice:
				for i, item := range loopv.([]interface{}) {
					res = append(res, LoopRes{
						Item:    item,
						ItemKey: i,
					})
				}
				return res
			case reflect.Map:
				v := reflect.ValueOf(loopv)
				for _, k := range v.MapKeys() {
					res = append(res, LoopRes{
						Item:    v.MapIndex(k).Interface(),
						ItemKey: k.Interface(),
					})
				}
				return res
			}
		}
	case reflect.Map:
		res := make([]LoopRes, 0)
		v := reflect.ValueOf(loop)
		for _, k := range v.MapKeys() {
			res = append(res, LoopRes{
				Item:    v.MapIndex(k).Interface(),
				ItemKey: k,
			})
		}
		return res
	}
	return nil
}

type LoopRes struct {
	ItemKey interface{}
	Item    interface{}
}
