package playbook

import (
	"context"
	"reflect"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/actions"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/mitchellh/mapstructure"
)

type SwitchAction struct {
	Con   string               `yaml:"con"`
	Tasks map[string][]TaskMap `yaml:"tasks"`
}

type TaskMap map[string]interface{}

func (t TaskMap) Get() Task {
	var task Task
	md, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "yaml",
		Result:  &task,
	})
	md.Decode(t)
	for k, ctx := range t {
		v, ok := actions.CustomActions.Load(k)
		if ok {
			cf := &CustomFunc{
				Action: v.(model.Action),
				Ctx:    ctx,
			}
			task.CustomAction = cf
		}
	}
	return task
}

func ConvMapToTasks(ms []TaskMap) []Task {
	var its []Task
	for _, t := range ms {
		its = append(its, t.Get())
	}
	return its
}

type SetFunc struct {
	Name  string    `yaml:"name"`
	Tasks []TaskMap `yaml:"tasks"`
}

func (a *SetFunc) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	act := &MutiActions{
		Tasks: a.Tasks,
	}
	actions.CustomActions.Store(a.Name, act)
	return "", nil
}

type MutiActions struct {
	Tasks []TaskMap
}

func (a *MutiActions) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (s string, err error) {
	return "", nil
}

type CustomFunc struct {
	Action model.Action
	Ctx    interface{}
}

func (a *CustomFunc) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	vars.Ctx = a.parseCtx(vars)
	// defer func() {
	// 	vars.Ctx = nil
	// }()
	return a.Action.Run(ctx, conn, conf, vars)
}

func (a *CustomFunc) parseCtx(vars *model.Vars) map[string]interface{} {
	return parseCtx(a.Ctx, vars)
}

func parseCtx(ctxv interface{}, vars *model.Vars) map[string]interface{} {
	switch reflect.TypeOf(ctxv).Kind() {
	// case reflect.Slice:
	// 	var res []string
	// 	for _, v := range ctxv.([]interface{}) {
	// 		if s, ok := v.(string); ok {
	// 			res = append(res, common.ParseTplWithPanic(s, vars))
	// 		} else {
	// 			return ctxv
	// 		}
	// 	}
	// 	return res
	// case reflect.String:
	// 	return common.ParseTplWithPanic(ctxv.(string), vars)
	case reflect.Map:
		v := reflect.ValueOf(ctxv)
		res := make(map[string]interface{})
		for _, k := range v.MapKeys() {
			if s, ok := v.MapIndex(k).Interface().(string); ok {
				if strings.HasPrefix(s, "$.") {
					ss, _ := common.GetVar(strings.TrimPrefix(s, "$."), vars)
					res[k.Interface().(string)] = ss
					continue
				}
				res[k.Interface().(string)] = common.ParseTplWithPanic(s, vars)
			} else {
				res[k.Interface().(string)] = v.MapIndex(k).Interface()
			}
		}
		return res
		// case reflect.Float64:
		// 	return ctxv.(float64)
	}
	return make(map[string]interface{})
}

func copyCtx(res map[string]interface{}, ctxv interface{}) {
	switch reflect.TypeOf(ctxv).Kind() {
	case reflect.Map:
		v := reflect.ValueOf(ctxv)
		for _, k := range v.MapKeys() {
			res[k.Interface().(string)] = v.MapIndex(k).Interface()
		}
	}
}

type DoFuncActions struct {
	F func(vars *model.Vars) (s string, err error)
}

func (a *DoFuncActions) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (s string, err error) {
	return a.F(vars)
}
