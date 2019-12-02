package playbook

import (
	"fmt"
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

	gs, err := inv.Groups()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("", gs)

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
