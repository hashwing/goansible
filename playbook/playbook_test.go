package playbook

import (
	"testing"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
)

func TestNewPlaybook(t *testing.T) {
	ps, err := UnmarshalFromFile("testdata/index.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	inv, err := inventory.NewFile("testdata/hosts")
	if err != nil {
		t.Error(err)
		return
	}

	cfg := model.Config{
		PlaybookFolder: "./testdata",
		Tag:            "dd",
	}

	err = Run(cfg, ps, inv)
	if err != nil {
		t.Error(err)
	}
}
