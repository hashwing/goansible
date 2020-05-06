package common

import (
	"html/template"
)

var funcMap = template.FuncMap{
	"join":          join,
	"unescaped":     unescaped,
	"add":           add,
	"join_hostvars": join_hostvars,
}

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

func add(a, b interface{}) interface{} {
	if aInt, ok := a.(int); ok {
		if bInt, ok := b.(int); ok {
			return aInt + bInt
		}
		if bInt, ok := b.(int64); ok {
			return int64(aInt) + bInt
		}
	}
	return 0
}

func join_hostvars(groupVars map[string]map[string]interface{}, key, step string) interface{} {
	i := 0
	res := ""
	for _, hostvars := range groupVars {
		res += hostvars[key].(string)
		if i < len(groupVars)-1 {
			res += step
		}
		i++
	}
	return template.HTML(res)
}
