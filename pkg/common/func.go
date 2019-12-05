package common

import (
	"html/template"
)

func join(a interface{}, step string) interface{} {
	s := a.([]interface{})
	res := ""
	for i, item := range s {

		res += item.(string)
		if i < len(s)-1 {
			res += step
		}
	}
	return template.HTML(res)
}

func unescaped(x string) interface{} { return template.HTML(x) }
