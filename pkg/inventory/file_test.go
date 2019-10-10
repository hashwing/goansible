package inventory

import (
	"encoding/json"
	"testing"
)

func TestFileGroup(t *testing.T) {
	inv, _ := NewFile("testdata/hosts")
	gs, err := inv.Groups()
	if err != nil {
		t.Error(err)
		return
	}
	data, _ := json.Marshal(gs)
	t.Log(string(data))
}
