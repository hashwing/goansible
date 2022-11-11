package playbook

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/pkg/termutil"
	"github.com/hashwing/goansible/pkg/transport"
)

const (
	sshUser    = "ansible_ssh_user"
	sshPwd     = "ansible_ssh_pass"
	sshKey     = "ansible_ssh_key"
	sshHost    = "ansible_ssh_host"
	sshPort    = "ansible_ssh_port"
	sshSudoPwd = "ansible_ssh_sudopass"
	sshLocal   = "ansible_ssh_local"
)

func (p *Playbook) Run(gs map[string]*model.Group, customVars map[string]interface{}, vars map[string]interface{}, conf model.Config) error {
	termutil.FullInfo("Playbook [%s] ", "=", p.Name)
	if (conf.IsUntag && TagFilter(conf.Tags, p.Tags)) || (!conf.IsUntag && !TagFilter(conf.Tags, p.Tags)) {
		termutil.Printf("slip: tag filter\n")
		return nil
	}
	groups := make(map[string]interface{})
	for _, g := range gs {
		groupVars := make(map[string]map[string]interface{})
		for _, h := range g.Hosts {
			groupVars[h.Name] = h.HostVars
		}
		groups[g.Name] = groupVars
	}
	tvars := &model.Vars{
		Values: vars,
		Groups: groups,
	}
	if p.When != "" {
		if !When(p.When, tvars) {
			termutil.Printf("slip: [%s]", p.Name)
			return nil
		}
	}
	if p.ImportPlaybook != "" {
		pls, err := UnmarshalFromFile(conf.PlaybookFolder + "/" + p.ImportPlaybook)
		if err != nil {
			return err
		}
		for _, pl := range pls {
			if len(pl.Tags) == 0 {
				pl.Tags = p.Tags
			}
			tvars := &model.Vars{
				Values: vars,
				Groups: groups,
			}
			if p.When != "" {
				if !When(p.When, tvars) {
					termutil.Printf("slip: [%s]", p.Name)
					return nil
				}
			}
			err := pl.Run(gs, customVars, vars, conf)
			if err != nil {
				return err
			}
			vars = pl.Vars
			p.Vars = pl.Vars
		}
		return nil
	}
	if p.SubPlaybook != nil {
		subConf := model.Config{
			PlaybookFolder: conf.PlaybookFolder + "/" + p.SubPlaybook.WorkDir,
			Tag:            conf.Tag,
			Tags:           conf.Tags,
			Untag:          conf.Untag,
			IsUntag:        conf.IsUntag,
			PlaybookFile:   "index.yaml",
			InvFile:        "values.yaml",
		}
		if p.SubPlaybook.PlaybookFile != "" {
			subConf.PlaybookFile = p.SubPlaybook.PlaybookFile
		}
		if p.SubPlaybook.InvFile != "" {
			subConf.InvFile = p.SubPlaybook.InvFile
		}
		_, err := Run(subConf, customVars, gs)
		return err
	}
	if p.Hosts == "" {
		p.Hosts = "all"
	}

	err := initConn(gs, "all")
	if err != nil {
		return err
	}

	g, ok := gs[p.Hosts]
	if !ok {
		return errors.New(p.Hosts + " hosts group undefine")
	}

	groupVars := make(map[string]map[string]interface{})
	for _, h := range g.Hosts {
		groupVars[h.Name] = h.HostVars
	}

	values, err := FilesToValues(p.IncludeValues, conf)
	if err != nil {
		return err
	}
	if p.Vars == nil {
		p.Vars = make(map[string]interface{})
	}
	if vars != nil {
		common.MergeValues(p.Vars, vars)
	}
	if values != nil {
		common.MergeValues(p.Vars, values)
	}
	if customVars != nil {
		common.MergeValues(p.Vars, customVars)
	}

	for _, tt := range p.Tasks {
		t := tt.Get()
		_, err := p.preTask(t, groups, groupVars, g, conf, nil)
		if err != nil {
			return err
		}
		// if len(t.Tags) == 0 {
		// 	t.Tags = p.Tags
		// }
		// pctxvars := make(map[string]interface{})
		// if cf, ok := t.Action().(*CustomFunc); ok {
		// 	if mas, ok := cf.Action.(*MutiActions); ok {
		// 		var f = func(vars *model.Vars) (string, error) {
		// 			ctxvars := parseCtx(cf.Ctx, vars)
		// 			for _, mtask := range mas.Tasks {
		// 				itask := mtask.Get()
		// 				ctxvars, err = p.runTask(itask, groups, groupVars, g, conf, ctxvars)
		// 				if err != nil {
		// 					if itask.IgnoreError {
		// 						termutil.Printf("ignore error ...\n")
		// 						continue
		// 					}
		// 					return "", err
		// 				}
		// 			}
		// 			copyCtx(pctxvars, ctxvars)
		// 			return "", nil
		// 		}
		// 		t.CustomAction = &DoFuncActions{F: f}
		// 	}
		// }
		// if t.Include != "" || len(t.Tasks) > 0 || t.Playbook != "" {
		// 	var f = func(vars *model.Vars) (string, error) {
		// 		itasks := ConvMapToTasks(t.Tasks)
		// 		if t.Include != "" {
		// 			itasks, err = FileToTasks(t.Include, conf)
		// 			if err != nil {
		// 				return "", err
		// 			}
		// 		} else if t.Playbook != "" {
		// 			itasks, err = PlaybookToTasks(t.Playbook, conf)
		// 			if err != nil {
		// 				return "", err
		// 			}
		// 		} else if t.Switch != nil {
		// 			itasks = DoSwitch(t.Switch, vars)
		// 		}
		// 		ctxvars := LoopRes{
		// 			Item:    vars.Item,
		// 			ItemKey: vars.ItemKey,
		// 		}
		// 		for _, itask := range itasks {
		// 			_, err := p.runTask(itask, groups, groupVars, g, conf, ctxvars)
		// 			if err != nil {
		// 				if itask.IgnoreError {
		// 					termutil.Printf("ignore error ...\n")
		// 					continue
		// 				}
		// 				return "", err
		// 			}
		// 		}
		// 		return "", nil
		// 	}
		// 	t.CustomAction = &DoFuncActions{F: f}
		// }
		// if t.While != "" {
		// 	var f = func(vars *model.Vars) (string, error) {
		// 		t.CustomAction = nil
		// 		i := 0
		// 		ctxvars := make(map[string]interface{})
		// 		for {
		// 			if !When(t.While, vars) {
		// 				break
		// 			}
		// 			ctxvars["ItemKey"] = i
		// 			_, err := p.runTask(t, groups, groupVars, g, conf, ctxvars)
		// 			if err != nil {
		// 				if t.IgnoreError {
		// 					termutil.Printf("ignore error ...\n")
		// 					continue
		// 				}
		// 				return "", err
		// 			}
		// 			i++
		// 		}
		// 		return "", err
		// 	}
		// 	t.CustomAction = &DoFuncActions{F: f}
		// }
		// if t.Async != "" {
		// 	go func() {
		// 		AddAwait(t.Async)
		// 		_, err := p.runTask(t, groups, groupVars, g, conf, pctxvars)
		// 		if err != nil {
		// 			if t.IgnoreError {
		// 				termutil.Printf("ignore error ...\n")
		// 				DoneAwait(t.Async, nil)
		// 				return
		// 			}
		// 		}
		// 		DoneAwait(t.Async, err)

		// 	}()
		// 	continue
		// }
		// if len(t.Await) > 0 {
		// 	termutil.Printf("await ...\n")
		// 	err := Await(t.Await, t.IgnoreError)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// _, err := p.runTask(t, groups, groupVars, g, conf, pctxvars)
		// if err != nil {
		// 	if t.IgnoreError {
		// 		termutil.Printf("ignore error ...\n")
		// 		continue
		// 	}
		// 	return err
		// }
		// fmt.Println("")
	}

	return nil
}

func (p *Playbook) preTask(t Task, groups map[string]interface{}, groupVars map[string]map[string]interface{}, group *model.Group, conf model.Config, pctxvars map[string]interface{}) (map[string]interface{}, error) {
	if len(t.Tags) == 0 {
		t.Tags = p.Tags
	}
	var err error
	if pctxvars == nil {
		pctxvars = make(map[string]interface{})
	}

	if cf, ok := t.Action().(*CustomFunc); ok {
		if mas, ok := cf.Action.(*MutiActions); ok {
			var f = func(vars *model.Vars) (string, error) {
				ctxvars := parseCtx(cf.Ctx, vars)
				for _, mtask := range mas.Tasks {
					itask := mtask.Get()
					ctxvars, err = p.preTask(itask, groups, groupVars, group, conf, ctxvars)
					if err != nil {
						if itask.IgnoreError {
							termutil.Printf("ignore error ...\n")
							continue
						}
						return "", err
					}
				}
				copyCtx(pctxvars, ctxvars)
				return "", nil
			}
			t.CustomAction = &DoFuncActions{F: f}
		}
	}

	if t.Include != "" || len(t.Tasks) > 0 || t.Playbook != "" {
		var f = func(vars *model.Vars) (string, error) {
			itasks := ConvMapToTasks(t.Tasks)
			if t.Include != "" {
				itasks, err = FileToTasks(t.Include, conf)
				if err != nil {
					return "", err
				}
			} else if t.Playbook != "" {
				itasks, err = PlaybookToTasks(t.Playbook, conf)
				if err != nil {
					return "", err
				}
			} else if t.Switch != nil {
				itasks = DoSwitch(t.Switch, vars)
			}

			ctxvars := map[string]interface{}{
				"Item":    vars.Item,
				"ItemKey": vars.ItemKey,
			}
			for _, itask := range itasks {
				_, err := p.preTask(itask, groups, groupVars, group, conf, ctxvars)
				if err != nil {
					if itask.IgnoreError {
						termutil.Printf("ignore error ...\n")
						continue
					}
					return "", err
				}
			}
			copyCtx(pctxvars, ctxvars)
			return "", nil
		}
		t.CustomAction = &DoFuncActions{F: f}
	}
	if t.While != "" {
		customAction := t.CustomAction
		var f = func(vars *model.Vars) (string, error) {
			t.CustomAction = customAction
			i := 0
			ctxvars := make(map[string]interface{})
			for {
				if !When(t.While, vars) {
					break
				}
				ctxvars["ItemKey"] = i
				_, err := p.runTask(t, groups, groupVars, group, conf, ctxvars)
				if err != nil && !t.IgnoreError {
					return "", err
				}
				i++
				fmt.Println("")
				if t.Sleep > 0 {
					time.Sleep(time.Duration(t.Sleep) * time.Second)
				}
			}
			return "", err
		}
		t.CustomAction = &DoFuncActions{F: f}
	}
	if t.Async != "" {
		go func() {
			AddAwait(t.Async)
			_, err := p.runTask(t, groups, groupVars, group, conf, pctxvars)
			if err != nil {
				if t.IgnoreError {
					termutil.Printf("ignore error ...\n")
					DoneAwait(t.Async, nil)
					return
				}
			}
			DoneAwait(t.Async, err)

		}()
		return pctxvars, nil
	}
	if len(t.Await) > 0 {
		termutil.Printf("await ...\n")
		err := Await(t.Await, t.IgnoreError)
		if err != nil {
			return pctxvars, err
		}
	}
	_, err = p.runTask(t, groups, groupVars, group, conf, pctxvars)
	if err != nil {
		if t.IgnoreError {
			termutil.Printf("ignore error ...\n")
			return pctxvars, nil
		}
		return pctxvars, err
	}
	return pctxvars, nil
}

func (p *Playbook) runTask(t Task, groups map[string]interface{}, groupVars map[string]map[string]interface{}, group *model.Group, conf model.Config, ctxvars interface{}) (interface{}, error) {
	termutil.FullPrintf("Task [%s] ", "*", t.Name)
	if (conf.IsUntag && TagFilter(conf.Tags, t.Tags)) || (!conf.IsUntag && !TagFilter(conf.Tags, t.Tags)) {
		termutil.Printf("slip: tag filter\n")
		return ctxvars, nil
	}
	action := t.Action()
	if action == nil {
		return ctxvars, nil
	}

	var wg sync.WaitGroup
	if t.Once {
		wg.Add(1)
	} else {
		wg.Add(len(group.Hosts))
	}
	var globalErr error
	start := time.Now()
	for _, h := range group.Hosts {
		go func(h *model.Host) {
			defer wg.Done()
			vars := &model.Vars{
				Values:    p.Vars,
				GroupVars: groupVars,
				HostVars:  groupVars[h.Name],
				Groups:    groups,
				Ctx:       ctxvars,
			}
			if t.When != "" {
				if !When(t.When, vars) {
					termutil.Printf("slip: [%s]", h.Name)
					return
				}
			}
			//loops := make([]LoopRes,0)
			loops := []LoopRes{
				LoopRes{
					Item:    "",
					ItemKey: "0",
				},
			}
			if t.Loop != nil {
				loops = Loop(t.Loop, vars)
			}
			if t.While != "" {
				k, _ := common.GetVar("ctx.ItemKey", vars)
				loops = []LoopRes{
					LoopRes{
						Item:    "",
						ItemKey: k,
					},
				}
			}
			var cwg sync.WaitGroup
			cwg.Add(len(loops))
			n := getConcurrency(t.Concurrency, vars)
			conc := make(chan struct{}, n)
			for _, item := range loops {
				vars.Item = item.Item
				vars.ItemKey = item.ItemKey

				conn, err := getConn(h.Name)
				if err != nil {
					termutil.Errorf("error: [%s], msg: %v", h.Name, err)
					globalErr = err
					return
				}
				if n > 1 {
					conc <- struct{}{}
					go func(t Task, action model.Action, conn model.Connection, h *model.Host, item LoopRes, conf model.Config, vars *model.Vars) {
						defer func() {
							cwg.Done()
							<-conc
						}()
						err = p.runAction(t, action, conn, h, item, conf, vars)
						if err != nil {
							globalErr = err
							return
						}
					}(t, action, conn, h, item, conf, vars)
					continue
				}
				cwg.Done()
				err = p.runAction(t, action, conn, h, item, conf, vars)
				if err != nil {
					globalErr = err
					return
				}
			}
			cwg.Wait()
			termutil.Successf("success: [%s]", h.Name)
		}(h)
		if t.Once {
			break
		}
	}
	wg.Wait()
	end := time.Now()
	termutil.Debugf("cost: %ds start: %v end: %v", end.Unix()-start.Unix(), start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))
	return ctxvars, globalErr
}

func (p *Playbook) runAction(t Task, action model.Action, conn model.Connection, h *model.Host, item LoopRes, conf model.Config, vars *model.Vars) error {
	stdout, err := action.Run(context.Background(), conn, conf, vars)
	if err != nil {
		termutil.Errorf("error: [%s], msg: %v, %s", h.Name, err, stdout)
		return err
	}
	if t.StdOut != "" {
		common.SetVar(t.StdOut, strings.TrimSuffix(stdout, "\r\n"), vars)
	}
	if t.Setface != nil {
		t.Setface.Run(context.Background(), conn, conf, vars)
	}
	if t.Debug != "" {
		info := stdout
		if t.Debug != "stdout" {
			infov, err := common.ParseTpl(t.Debug, vars)
			if err != nil {
				termutil.Errorf("error: [%s], msg: %v, %s", h.Name, err, info)
				return err
			}
			if infov != "__stdout__" {
				info = infov
			}
		}
		if info != "__false__" {
			termutil.Changedf("debug:\n%s", info)
		}
	}
	if t.Loop != nil {
		termutil.Successf("success: [%s] item => %+v", h.Name, item.Item)
	}
	return nil
}

var globalConns map[string]model.Connection = make(map[string]model.Connection)
var isInitConn = false

func initConn(gs map[string]*model.Group, name string) error {
	if isInitConn {
		return nil
	}
	var wg sync.WaitGroup
	var gerr error
	mu := new(sync.Mutex)
	wg.Add(len(gs[name].Hosts))
	for _, h := range gs[name].Hosts {
		go func(h *model.Host) {
			defer wg.Done()
			conn, err := connect(h)
			if err != nil {
				gerr = err
				termutil.Errorf(err.Error())
				return
			}
			mu.Lock()
			globalConns[h.Name] = conn
			mu.Unlock()
		}(h)
	}
	wg.Wait()
	isInitConn = true
	return gerr
}

func connect(h *model.Host) (model.Connection, error) {
	if h.Name == "localhost" {
		return transport.ConnectCmd(), nil
	}
	if isLocal, ok := h.HostVars[sshLocal]; ok {
		if isLocal.(bool) {
			return transport.ConnectCmd(), nil
		}
	}
	host, ok := h.HostVars[sshHost]
	if !ok {
		return nil, errors.New(h.Name + " ssh host undefine")
	}
	user, ok := h.HostVars[sshUser]
	if !ok {
		user = "root"
	}

	port, ok := h.HostVars[sshPort]
	if !ok {
		port = "22"
	}

	pwd, ok := h.HostVars[sshPwd]
	if !ok {
		pwd = ""
	}
	key, ok := h.HostVars[sshKey]
	if !ok {
		key = ""
	}
	sudoPwds := []string{}
	sudoPwd, ok := h.HostVars[sshSudoPwd]
	if ok {
		sudoPwds = append(sudoPwds, sudoPwd.(string))
	}
	conn, err := transport.Connect(user.(string), pwd.(string), key.(string), host.(string)+":"+port.(string), sudoPwds...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getConn(name string) (model.Connection, error) {
	// if name == "localhost" {
	// 	return transport.ConnectCmd(), nil
	// }
	return globalConns[name], nil
}
