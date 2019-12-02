package playbook

import (
	"fmt"
	"io/ioutil"

	"github.com/hashwing/goansible/pkg/actions"
	"gopkg.in/yaml.v2"
)

type Playbook struct {
	Name           string                 `yaml:"name"`
	Hosts          string                 `yaml:"hosts"`
	Vars           map[string]interface{} `yaml:"vars"`
	ImportPlaybook string                 `yaml:"import_playbook"`
	IncludeValues  []string               `yaml:"include_values"`
	Tasks          []Task                 `yaml:"tasks"`
}

type Task struct {
	Name        string                   `yaml:"name"`
	FileAction  *actions.FileAction      `yaml:"file"`
	Template    *actions.TemplateAction  `yaml:"template"`
	ShellAction *actions.ShellAction     `yaml:"shell"`
	StdOut      string                   `yaml:"stdout"`
	Regexp      *actions.RegexpAction    `yaml:"regexp"`
	Debug       string                   `yaml:"debug"`
	Loop        interface{}              `yaml:"loop"`
	Setface     *actions.SetfaceAction   `yaml:"setface"`
	Until       *actions.UntilAction     `yaml:"until"`
	Directory   *actions.DirectoryAction `yaml:"directory"`
	When        string                   `yaml:"when"`
	Include     string                   `yaml:"include"`
	IgnoreError bool                     `yaml:"ignore_error"`
}

func UnmarshalFromFile(playbookFile string) ([]Playbook, error) {
	fileContents, err := ioutil.ReadFile(playbookFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open playbook file: %s", err)
	}

	var playbooks []Playbook
	err = yaml.Unmarshal(fileContents, &playbooks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal playbook contents: %s", err)
	}

	return playbooks, nil
}
