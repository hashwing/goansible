package playbook

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/actions"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/pkg/termutil"
	"gopkg.in/yaml.v2"
)

type Playbook struct {
	Name           string                 `yaml:"name"`
	Hosts          string                 `yaml:"hosts"`
	Vars           map[string]interface{} `yaml:"vars"`
	ImportPlaybook string                 `yaml:"import_playbook"`
	SubPlaybook    *SubPlaybookOption     `yaml:"sub_playbook"`
	IncludeValues  []string               `yaml:"include_values"`
	Tasks          []TaskMap              `yaml:"tasks"`
	Tag            string                 `yaml:"tag"`
	Tags           []string               `yaml:"tags"`
	When           string                 `yaml:"when"`
}

type SubPlaybookOption struct {
	WorkDir      string `yaml:"workdir"`
	PlaybookFile string `yaml:"playbook"`
	InvFile      string `yaml:"values"`
}

type Task struct {
	Name         string                   `yaml:"name"`
	StdOut       string                   `yaml:"stdout"`
	Debug        string                   `yaml:"debug"`
	Loop         interface{}              `yaml:"loop"`
	When         string                   `yaml:"when"`
	Sleep        int64                    `yaml:"sleep"`
	While        string                   `yaml:"while"`
	Async        string                   `yaml:"async"`
	Await        []string                 `yaml:"await"`
	Concurrency  interface{}              `yaml:"concurrency"`
	Include      string                   `yaml:"include"`
	Tasks        []TaskMap                `yaml:"tasks"`
	Playbook     string                   `yaml:"playbook"`
	IgnoreError  bool                     `yaml:"ignore_error"`
	Tag          string                   `yaml:"tag"`
	Tags         []string                 `yaml:"tags"`
	Once         bool                     `yaml:"once"`
	Switch       *SwitchAction            `yaml:"switch"`
	FileAction   *actions.FileAction      `yaml:"file"`
	Template     *actions.TemplateAction  `yaml:"template"`
	ShellAction  *actions.ShellAction     `yaml:"shell"`
	Regexp       *actions.RegexpAction    `yaml:"regexp"`
	Setface      *actions.SetfaceAction   `yaml:"setface"`
	Until        *actions.UntilAction     `yaml:"until"`
	Directory    *actions.DirectoryAction `yaml:"directory"`
	Cert         *actions.CertAction      `yaml:"cert"`
	Curl         *actions.CurlAction      `yaml:"curl"`
	Js           *actions.JsAction        `yaml:"js"`
	JsFile       *actions.JsFileAction    `yaml:"jsfile"`
	Lua          *actions.LuaAction       `yaml:"lua"`
	LuaFile      *actions.LuaFileAction   `yaml:"luafile"`
	Req          *actions.ReqAction       `yaml:"req"`
	DBInitAction *actions.DBInitAction    `yaml:"dbinit"`
	DBAction     *actions.DBAction        `yaml:"db"`
	FtpAction    *actions.FTPAction       `yaml:"ftp"`
	SetFunc      *SetFunc                 `yaml:"setfunc"`
	CustomAction model.Action             `yaml:"-"`
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
	// if t.Setface != nil {
	// 	action = t.Setface
	// }
	if t.Directory != nil {
		action = t.Directory
	}
	if t.Cert != nil {
		action = t.Cert
	}
	if t.Curl != nil {
		action = t.Curl
	}
	if t.Js != nil {
		action = t.Js
	}
	if t.JsFile != nil {
		action = t.JsFile
	}
	if t.Lua != nil {
		action = t.Lua
	}
	if t.LuaFile != nil {
		action = t.LuaFile
	}
	if t.Req != nil {
		action = t.Req
	}
	if t.DBInitAction != nil {
		action = t.DBInitAction
	}
	if t.DBAction != nil {
		action = t.DBAction
	}
	if t.FtpAction != nil {
		action = t.FtpAction
	}
	if t.SetFunc != nil {
		action = t.SetFunc
	}
	if t.CustomAction != nil {
		action = t.CustomAction
	}
	if action == nil {
		action = new(actions.NoneAction)
	}
	return action
}

func UnmarshalFromFile(playbookFile string) ([]*Playbook, error) {
	fileContents, err := ioutil.ReadFile(playbookFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open playbook file: %s", err)
	}

	var playbooks []*Playbook
	err = yaml.Unmarshal(fileContents, &playbooks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal playbook contents: %s", err)
	}
	return playbooks, nil
}

func Run(cfg model.Config, customVars map[string]interface{}, gs map[string]*model.Group) (map[string]interface{}, error) {
	var err error
	var subinv model.Inventory
	if cfg.InvFile != "" {
		subinv, err = inventory.NewYaml(cfg.PlaybookFolder + "/" + cfg.InvFile)
		if err != nil {
			if !os.IsNotExist(err) {
				termutil.Errorf(err.Error())
				return nil, err
			}
			termutil.Changedf("inventory file '%s' not found, use default inventory", cfg.InvFile)
			subinv = &model.DefaultInventory{}
		}
	} else {
		subinv = &model.DefaultInventory{}
	}

	subCustomVars, err := subinv.Vars()
	if err != nil {
		return nil, err
	}
	if customVars != nil {
		common.MergeValues(subCustomVars, customVars)
	}

	ps, err := UnmarshalFromFile(cfg.PlaybookFolder + "/" + cfg.PlaybookFile)
	if err != nil {
		termutil.Errorf(err.Error())
		return nil, err
	}
	if gs == nil {
		gs, err = subinv.Groups()
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		if err := recover(); err != nil {
			termutil.Errorf("erorr: %v", err)
		}
	}()
	start := time.Now()
	vars := make(map[string]interface{})
	LoadModules("modules", cfg)
	LoadLibs("libs", cfg)
	termutil.FullInfo(fmt.Sprintf("Start playbooks [%s] ", cfg.PlaybookFolder), "=")
	for _, p := range ps {
		err := p.Run(gs, subCustomVars, vars, cfg)
		if err != nil {
			return nil, err
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
	termutil.FullInfo(fmt.Sprintf("Finish playbooks [%s] ", cfg.PlaybookFolder), "=")
	termutil.Printf("start: %v", start)
	termutil.Printf("end: %v", end)
	termutil.Printf("cost: %dm%ds\n", m, s)
	return vars, nil
}
