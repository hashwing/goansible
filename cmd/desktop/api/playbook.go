package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/pkg/termutil"
	"github.com/hashwing/goansible/playbook"
)

var (
	defaultValues = `
groups:
  all:
    localhost:
        ansible_ssh_host: ""
        ansible_ssh_port: ""
        ansible_ssh_user: ""
        ansible_ssh_pass: ""
        ansible_ssh_key: ""
  test:
    localhost: {}
vars: {}
`

	defaultIndex = `
- name: test
  hosts: test
  vars:
    test: 
    - dddd
    - ffff
  tag: dd
  tasks:
  - name: Exec Command
    shell: echo hello world
    stdout: hostvars.stdout
    debug: "{{ .HostVars.stdout }}"
`
)

func (a *API) CreatePlaybook(name string) string {
	dir := a.cfg.PlaybookDir + "/" + name
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err.Error()
	}
	ioutil.WriteFile(dir+"/values.yaml", []byte(defaultValues), 0644)
	ioutil.WriteFile(dir+"/index.yaml", []byte(defaultIndex), 0644)
	return ""
}

func (a *API) ListPlaybook() []map[string]string {
	res := make([]map[string]string, 0)
	dirs, err := ioutil.ReadDir(a.cfg.PlaybookDir)
	if err != nil {
		return res
	}
	for _, fs := range dirs {
		if fs.IsDir() {
			res = append(res, map[string]string{"name": fs.Name()})
		}
	}
	return res
}

func (a *API) DeletePlaybook(name string) string {
	dir := a.cfg.PlaybookDir + "/" + name
	err := os.RemoveAll(dir)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *API) EditPlaybook(name string) string {
	dir := a.cfg.PlaybookDir + "/" + name
	err := exec.Command("code", dir).Run()
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *API) RunPlaybook(name, tag string) string {
	dir := a.cfg.PlaybookDir + "/" + name
	a.log = make([]string, 0)
	cfg := model.Config{
		InvFile:        "values.yaml",
		PlaybookFolder: dir,
		Tag:            tag,
	}
	termutil.With = 80
	termutil.Echo = func(s ...interface{}) (int, error) {
		ss := fmt.Sprint(s...)
		a.mu.Lock()
		defer a.mu.Unlock()
		a.log = append(a.log, ss)
		return 0, nil
	}
	inv, err := inventory.NewYaml(cfg.PlaybookFolder + "/" + cfg.InvFile)
	if err != nil {
		return err.Error()
	}
	ps, err := playbook.UnmarshalFromFile(cfg.PlaybookFolder + "/index.yaml")
	if err != nil {
		return err.Error()
	}
	go func() {
		playbook.Run(cfg, ps, inv)
	}()
	return ""
}

func (a *API) GetLog(i int) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if i >= len(a.log)-1 || len(a.log) == 0 {
		return []string{}
	}
	return a.log[i:]
}
