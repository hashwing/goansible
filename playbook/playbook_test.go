package playbook

import (
	"testing"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/pkg/termutil"
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
	start := time.Now()
	for _, p := range ps {
		err := p.Run(gs, model.Config{
			PlaybookFolder: "./testdata",
		})
		if err != nil {
			t.Error(err)
			return
		}
	}
	end := time.Now()
	cost := end.Unix() - start.Unix()
	var m int64
	var s = cost
	if cost > 120 {
		s = cost % 60
		m = cost / 60
	}
	termutil.FullInfo("Finish playbooks", "=")
	termutil.Printf("start: %v", start)
	termutil.Printf("end: %v", end)
	termutil.Printf("cost: %dm%ds\n", m, s)

}
