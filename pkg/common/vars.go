package common

import (
	"bytes"
	"strings"
	"sync"

	"html/template"

	"github.com/hashwing/goansible/model"
)

var mutex *sync.Mutex

func init() {
	mutex = new(sync.Mutex)
}

func GetVar(s string, vars *model.Vars) (interface{}, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	rs := strings.Split(s, ".")
	if len(rs) > 1 {
		var p map[string]interface{}
		switch rs[0] {
		case "hostvars":
			p = vars.HostVars
		case "values":
			p = vars.Values
		default:
			return nil, false
		}
		for i := 1; i < len(rs); i++ {
			if i == len(rs)-1 {
				if _, ok := p[rs[i]]; ok {
					return p[rs[i]], true
				}
			}
			if _, ok := p[rs[i]]; !ok {
				return nil, false
			}
			if _, ok := p[rs[i]].(map[string]interface{}); !ok {
				return nil, false
			}
		}
	}
	return nil, false
}

func SetVar(s string, value interface{}, vars *model.Vars) {
	mutex.Lock()
	defer mutex.Unlock()
	rs := strings.Split(s, ".")

	if len(rs) > 1 {
		var p map[string]interface{}
		switch rs[0] {
		case "hostvars":
			p = vars.HostVars
		case "values":
			p = vars.Values
		default:
			return
		}
		for i := 1; i < len(rs); i++ {
			if i == len(rs)-1 {
				p[rs[i]] = value
				return
			}
			if _, ok := p[rs[i]]; !ok {
				p[rs[i]] = make(map[string]interface{})
			}
			if _, ok := p[rs[i]].(map[string]interface{}); !ok {
				p[rs[i]] = make(map[string]interface{})
			}
			p = p[rs[i]].(map[string]interface{})
		}
	}
}

func ParseTpl(tpl string, vars *model.Vars) (string, error) {
	tmpl, err := template.New("tpl").Parse(tpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, vars)
	return b.String(), err
}

// MergeValues Merges source and destination map, preferring values from the source map
func MergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = MergeValues(destMap, nextMap)
	}
	return dest
}
