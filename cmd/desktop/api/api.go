package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type config struct {
	PlaybookDir string `yaml:"playbooks_dir"`
}

type API struct {
	workDir string
	log     []string
	mu      *sync.Mutex
	cfg     *config
}

func New() (*API, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	workDir := dir + "/.goansible"
	err = os.MkdirAll(workDir, 0755)
	if err != nil {
		return nil, err
	}
	configFile := workDir + "/config"
	_, err = os.Stat(configFile)
	if err != nil {
		err = ioutil.WriteFile(configFile, []byte("playbooks_dir: "+workDir+"/playbooks"), 0664)
		if err != nil {
			return nil, err
		}
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	fmt.Println(cfg.PlaybookDir)
	err = os.MkdirAll(cfg.PlaybookDir, 0755)
	if err != nil {
		return nil, err
	}
	return &API{
		workDir: workDir,
		mu:      new(sync.Mutex),
		cfg:     &cfg,
	}, nil
}
