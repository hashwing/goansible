package common

import (
	"testing"

	"github.com/hashwing/goansible/model"
)

func TestGetVar(t *testing.T) {
	vars := &model.Vars{
		Values: map[string]interface{}{
			"a": true,
			"test": map[string]interface{}{
				"a":  true,
				"bb": "dddd",
			},
		},
	}
	res1, ok1 := GetVar("values.a", vars)
	t.Log(res1, ok1)
	res2, ok2 := GetVar("values.test.a", vars)
	t.Log(res2, ok2)
}
