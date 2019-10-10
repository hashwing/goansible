package playbook

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/pkg/transport"
	log "github.com/sirupsen/logrus"
)

const (
	sshUser = "ansible_ssh_user"
	sshPwd  = "ansible_ssh_pass"
	sshKey  = "ansible_ssh_key"
	sshHost = "ansible_ssh_host"
	sshPort = "ansible_ssh_port"
)

func (p *Playbook) Run(gs map[string]*model.Group, conf model.Config) error {
	log.Infof("Start %s playbook ================================", p.Name)
	if p.ImportPlaybook != "" {
		pls, err := UnmarshalFromFile(conf.PlaybookFolder + "/" + p.ImportPlaybook)
		if err != nil {
			return err
		}
		for _, pl := range pls {
			err := pl.Run(gs, conf)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if p.Hosts == "" {
		p.Hosts = "all"
	}

	err := initConn(gs)
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
	common.MergeValues(p.Vars, values)

	for _, t := range p.Tasks {
		if t.Include != "" {
			itasks, err := FileToTasks(t.Include, conf)
			if err != nil {
				return err
			}
			for _, itask := range itasks {
				err := p.runTask(itask, groupVars, g, conf)
				if err != nil {
					if itask.IgnoreError {
						log.Info("ignore error ...")
						continue
					}
					return err
				}
			}
			continue
		}
		err := p.runTask(t, groupVars, g, conf)
		if err != nil {
			if t.IgnoreError {
				log.Info("ignore error ...")
				continue
			}
			return err
		}
	}

	return nil
}

func (p *Playbook) runTask(t Task, groupVars map[string]map[string]interface{}, group *model.Group, conf model.Config) error {
	log.Infof("Start %s task ******************************", t.Name)
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
	if action == nil {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(group.Hosts))
	var globalErr error

	for _, h := range group.Hosts {
		go func(h *model.Host) {
			defer wg.Done()
			vars := &model.Vars{
				Values:    p.Vars,
				GroupVars: groupVars,
				HostVars:  groupVars[h.Name],
			}
			if t.When != "" {
				if !When(t.When, vars) {
					log.Infof("slip:%s", h.Name)
					return
				}
			}
			loops := []interface{}{""}
			if t.Loop != nil {
				loops = Loop(t.Loop, vars)
			}
			for _, item := range loops {
				vars.Item = item
				conn, err := getConn(h.Name)
				if err != nil {
					log.Errorf("error:%s, msg: %v", h.Name, err)
					globalErr = err
					return
				}
				stdout, err := action.Run(context.Background(), conn, conf, vars)
				if err != nil {
					log.Errorf("error:%s, msg: %v", h.Name, err)
					globalErr = err
					return
				}
				if t.StdOut != "" {
					rs := strings.Split(t.StdOut, ".")
					if len(rs) == 2 {
						switch rs[0] {
						case "hostvars":
							groupVars[h.Name][rs[1]] = strings.TrimSuffix(stdout, "\r\n")
						case "values":
							p.Vars[rs[1]] = strings.TrimSuffix(stdout, "\r\n")
						}
					}
				}
				if t.Debug != "" {
					info, err := common.ParseTpl(t.Debug, vars)
					if err != nil {
						log.Errorf("error:%s, msg: %v", h.Name, err)
						globalErr = err
						return
					}
					log.Infof("debug:%s", info)
				}
				if v, ok := item.(string); ok && v != "" {
					log.Infof("success:%s itme=>%s", h.Name, v)
				}
			}
			log.Infof("success:%s", h.Name)
		}(h)
	}
	wg.Wait()
	return globalErr
}

var globalConns map[string]model.Connection = make(map[string]model.Connection)

func initConn(gs map[string]*model.Group) error {
	for _, h := range gs["all"].Hosts {
		host, ok := h.HostVars[sshHost]
		if !ok {
			return errors.New(h.Name + " ssh host undefine")
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
		conn, err := transport.Connect(user.(string), pwd.(string), key.(string), host.(string)+":"+port.(string))
		if err != nil {
			return err
		}
		globalConns[h.Name] = conn
	}
	return nil
}

func getConn(name string) (model.Connection, error) {
	return globalConns[name], nil
}
