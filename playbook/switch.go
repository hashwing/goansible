package playbook

import (
	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/expr"
)

func DoSwitch(s *SwitchAction, vars *model.Vars) []Task {
	res, err := expr.Eval(s.Con, vars)
	if err != nil {
		panic(err)
	}
	for key, tasks := range s.Tasks {
		if key == res.(string) {
			return ConvMapToTasks(tasks)
		}
	}
	return nil
}
