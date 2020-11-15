package inventory

import (
	"errors"
	"io/ioutil"

	"github.com/hashwing/goansible/model"
	"gopkg.in/yaml.v2"
)

type valuesYaml struct {
	Groups map[string]map[string]map[string]interface{} `yaml:"groups"`
	Vars   map[string]interface{}                       `yaml:"vars"`
}

type yamlInv struct {
	path   string
	values valuesYaml
}

//NewYaml yaml inv
func NewYaml(path string) (model.Inventory, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var values valuesYaml
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return nil, err
	}
	return &yamlInv{path, values}, nil
}

func (g *yamlInv) Vars() (map[string]interface{}, error) {
	if g.values.Vars == nil {
		return make(map[string]interface{}), nil
	}
	return g.values.Vars, nil
}

// GetGroup get group struct by inventory file
func (g *yamlInv) Groups() (map[string]*model.Group, error) {
	res := make(map[string]*model.Group)
	allGroup, ok := g.values.Groups["all"]
	if !ok {
		return res, errors.New("all group not found")
	}
	for groupName, vars := range g.values.Groups {
		group := &model.Group{
			Name:  groupName,
			Hosts: make(map[string]*model.Host),
		}
		for hostName, hostVars := range vars {
			vars := make(map[string]interface{})
			if groupName != "all" {
				vars = allGroup[hostName]
			}
			for k, v := range hostVars {
				vars[k] = v
			}
			vars["ansible_hostname"] = hostName
			host := &model.Host{
				Name:     hostName,
				HostVars: vars,
			}
			group.Hosts[hostName] = host
			if groupName == "all" {
				hgroup := &model.Group{
					Name: hostName,
					Hosts: map[string]*model.Host{
						hostName: host,
					},
				}
				res[hostName] = hgroup
			}
		}
		res[groupName] = group
	}
	return res, nil
}
