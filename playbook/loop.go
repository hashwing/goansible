package playbook

import (
	"reflect"
	"strconv"

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
				v := reflect.ValueOf(loopv)
				for i := 0; i < v.Len(); i++ {
					res = append(res, LoopRes{
						Item:    v.Index(i).Interface(),
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

func getConcurrency(v interface{}, vars *model.Vars) int {
	if v == nil {
		return 1
	}
	if n, ok := v.(int); ok {
		return n
	}
	if s, ok := v.(string); ok {
		s = common.ParseTplWithPanic(s, vars)
		n, err := strconv.Atoi(s)
		if err != nil {
			return 1
		}
		return n
	}
	return 1
}
