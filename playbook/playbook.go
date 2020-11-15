package playbook

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/actions"
	"github.com/hashwing/goansible/pkg/termutil"
	"gopkg.in/yaml.v2"
)

type Playbook struct {
	Name           string                 `yaml:"name"`
	Hosts          string                 `yaml:"hosts"`
	Vars           map[string]interface{} `yaml:"vars"`
	ImportPlaybook string                 `yaml:"import_playbook"`
	IncludeValues  []string               `yaml:"include_values"`
	Tasks          []Task                 `yaml:"tasks"`
	Tag            string                 `yaml:"tag"`
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
	Tag         string                   `yaml:"tag"`
}

func (t *Task) Action() model.Action {
	var action model.Action
	if t.FileAction != nil {
		action = t.FileAction
	}
	if t.Template != nil {
		action = t.Template
	}
	if t.ShellAction != nil {
		action = t.ShellAction
	}
	if t.Regexp != nil {
		action = t.Regexp
	}
	if t.Until != nil {
		action = t.Until
	}
	if t.Setface != nil {
		action = t.Setface
	}
	if t.Directory != nil {
		action = t.Directory
	}
	return action
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

func Run(cfg model.Config, ps []Playbook, inv model.Inventory) error {
	gs, err := inv.Groups()
	if err != nil {
		return err
	}
	customVars, err := inv.Vars()
	if err != nil {
		return err
	}
	defer func() {
		if err := recover(); err != nil {
			termutil.Errorf("erorr: %v", err)
		}
	}()
	start := time.Now()
	vars := make(map[string]interface{})
	termutil.FullInfo("Start playbooks ", "=")
	for _, p := range ps {
		err := p.Run(gs, customVars, vars, cfg)
		if err != nil {
			return err
		}
		vars = p.Vars
	}
	end := time.Now()
	cost := end.Unix() - start.Unix()
	var m int64
	var s = cost
	if cost > 120 {
		s = cost % 60
		m = cost / 60
	}
	termutil.FullInfo("Finish playbooks ", "=")
	termutil.Printf("start: %v", start)
	termutil.Printf("end: %v", end)
	termutil.Printf("cost: %dm%ds\n", m, s)
	return nil
}
