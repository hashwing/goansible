package main

import (
	"flag"
	"os"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/playbook"
	log "github.com/sirupsen/logrus"
)

func main() {
	workdir := flag.String("workdir", ".", "run playbook in specially dir")
	tag := flag.String("tag", "", "use to tag filter")
	flag.Parse()
	workFolder := strings.Replace(*workdir, "\\", "/", -1)
	ps, err := playbook.UnmarshalFromFile(workFolder + "/index.yaml")
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	inv, _ := inventory.NewFile(workFolder + "/hosts")
	cfg := model.Config{
		PlaybookFolder: workFolder,
		Tag:            *tag,
	}

	err = playbook.Run(cfg, ps, inv)
	if err != nil {
		log.Error(err)
	}
}
