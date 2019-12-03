package main

import (
	"os"
	"strings"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/pkg/termutil"
	"github.com/hashwing/goansible/playbook"
	log "github.com/sirupsen/logrus"
)

func main() {
	playbookFolder := "./"
	if len(os.Args) > 1 {
		playbookFolder = strings.Replace(os.Args[1], "\\", "/", -1)
	}
	ps, err := playbook.UnmarshalFromFile(playbookFolder + "/index.yaml")
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	inv, _ := inventory.NewFile(playbookFolder + "/hosts")
	gs, err := inv.Groups()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	defer func() {
		if err := recover(); err != nil {
			termutil.Errorf("erorr: %v", err)
		}
	}()
	start := time.Now()
	for _, p := range ps {
		err := p.Run(gs, model.Config{
			PlaybookFolder: playbookFolder,
		})
		if err != nil {
			log.Error(err)
			os.Exit(-1)
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
