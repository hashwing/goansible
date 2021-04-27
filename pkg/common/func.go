package common

import (
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"join":           join,
	"plus":           plus,
	"minus":          minus,
	"join_groupvars": join_groupvars,
	"trimSuffix":     trimSuffix,
	"trimPrefix":     trimPrefix,
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
	return res
}

func plus(a, b interface{}) interface{} {
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

func minus(a, b interface{}) interface{} {
	if aInt, ok := a.(int); ok {
		if bInt, ok := b.(int); ok {
			return aInt - bInt
		}
		if bInt, ok := b.(int64); ok {
			return int64(aInt) - bInt
		}
	}
	return 0
}

func join_groupvars(groupVars map[string]map[string]interface{}, key, step string) interface{} {
	i := 0
	res := ""
	for _, hostvars := range groupVars {
		res += hostvars[key].(string)
		if i < len(groupVars)-1 {
			res += step
		}
		i++
	}
	return res
}

func trimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

func trimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}
