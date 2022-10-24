package playbook

import (
	"io/ioutil"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/actions"
)

type ModulesAction struct {
}

func LoadModules(dir string, conf model.Config) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), ".yaml") {
			tasks, err := FileToTaskMaps(dir+"/"+f.Name(), conf)
			if err != nil {
				return err
			}
			act := &MutiActions{
				Tasks: tasks,
			}
			actions.CustomActions.Store(strings.TrimSuffix(f.Name(), ".yaml"), act)
		}
		if strings.HasSuffix(f.Name(), ".js") {
			act := actions.JsFileAction(dir + "/" + f.Name())
			actions.CustomActions.Store(strings.TrimSuffix(f.Name(), ".js"), &act)
		}
	}
	return nil
}

func LoadLibs(dir string, conf model.Config) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	libData := ""
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(dir + "/" + f.Name())
		if err != nil {
			return err
		}
		libData += string(data) + "\n"

	}
	actions.LibStr = libData
	return nil
}
