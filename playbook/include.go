package playbook

import (
	"fmt"
	"io/ioutil"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"gopkg.in/yaml.v2"
)

func FileToTaskMaps(s string, conf model.Config) ([]TaskMap, error) {
	fileContents, err := ioutil.ReadFile(conf.PlaybookFolder + "/" + s)
	if err != nil {
		return nil, fmt.Errorf("failed to open playbook file: %s", err)
	}

	var tasks []TaskMap
	err = yaml.Unmarshal(fileContents, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal playbook contents: %s", err)
	}
	return tasks, nil
}

func FileToTasks(s string, conf model.Config) ([]Task, error) {
	tasks, err := FileToTaskMaps(s, conf)
	if err != nil {
		return nil, err
	}
	return ConvMapToTasks(tasks), nil
}

func FilesToValues(fs []string, conf model.Config) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	for _, f := range fs {
		fileContents, err := ioutil.ReadFile(conf.PlaybookFolder + "/" + f)
		if err != nil {
			return nil, fmt.Errorf("failed to open values file: %s", err)
		}

		var values map[string]interface{}
		err = yaml.Unmarshal(fileContents, &values)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal values contents: %s", err)
		}
		common.MergeValues(vals, values)
	}
	return vals, nil
}

func PlaybookToTasks(s string, conf model.Config) ([]Task, error) {
	fileContents, err := ioutil.ReadFile(conf.PlaybookFolder + "/" + s)
	if err != nil {
		return nil, fmt.Errorf("failed to open playbook file: %s", err)
	}

	var ps []Playbook
	err = yaml.Unmarshal(fileContents, &ps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal playbook contents: %s", err)
	}
	tasks := make([]Task, 0)
	for _, p := range ps {
		for _, t := range p.Tasks {
			tasks = append(tasks, t.Get())
		}
	}
	return tasks, nil
}
