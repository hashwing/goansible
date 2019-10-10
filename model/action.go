package model

import "context"

type Action interface {
	Run(context.Context, Connection, Config, *Vars) (string, error)
}

type Vars struct {
	Values    map[string]interface{}
	HostVars  map[string]interface{}
	GroupVars map[string]map[string]interface{}
	Item      interface{}
}
