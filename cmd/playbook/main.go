package main

import (
	"os"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
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

	for _, p := range ps {
		err := p.Run(gs, model.Config{
			PlaybookFolder: playbookFolder,
		})
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	}
}
