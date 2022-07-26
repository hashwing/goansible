package model

// Host 主机属性
type Host struct {
	Name     string
	HostVars map[string]interface{}
}

// Group 组
type Group struct {
	Name  string
	Hosts map[string]*Host
}

type Inventory interface {
	Groups() (map[string]*Group, error)
	Vars() (map[string]interface{}, error)
}

type DefaultInventory struct {
}

func (i *DefaultInventory) Groups() (map[string]*Group, error) {
	return map[string]*Group{
		"all": &Group{
			Name: "all",
			Hosts: map[string]*Host{
				"localhost": &Host{
					Name: "localhost",
				},
			},
		},
	}, nil
}

func (i *DefaultInventory) Vars() (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}
