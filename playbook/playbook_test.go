package playbook

import (
	"testing"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
)

func TestNewPlaybook(t *testing.T) {
	ps, err := UnmarshalFromFile("index.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	inv, _ := inventory.NewFile("hosts")
	gs, err := inv.Groups()
	if err != nil {
		t.Error(err)
		return
	}

	for _, p := range ps {
		err := p.Run(gs, model.Config{
			PlaybookFolder: "./testdata",
		})
		if err != nil {
			t.Error(err)
			return
		}
	}

}
