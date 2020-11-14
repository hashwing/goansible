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
