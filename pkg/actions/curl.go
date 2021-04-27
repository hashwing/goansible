package actions

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

type CurlAction struct {
	URL     string            `yaml:"url"`
	Options map[string]string `yaml:"options"`
}

func (a *CurlAction) parse(vars *model.Vars, conf model.Config) (*CurlAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()
	options := make(map[string]string)
	for key, v := range a.Options {
		switch key {
		case "data-yaml":
			data, err := yaml.YAMLToJSON([]byte(common.ParseTplWithPanic(v, vars)))
			if err != nil {
				return nil, err
			}
			options["data"] = "'" + string(data) + "'"
			continue
		case "data-file":
			data, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, v))
			if err != nil {
				return nil, err
			}
			options["data"] = "'" + common.ParseTplWithPanic(string(data), vars) + "'"
		case "data-yaml-file":
			data, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, v))
			if err != nil {
				return nil, err
			}
			s := common.ParseTplWithPanic(string(data), vars)
			data, err = yaml.YAMLToJSON([]byte(s))
			if err != nil {
				return nil, err
			}
			options["data"] = "'" + string(data) + "'"
		case "data":
			options["data"] = "'" + v + "'"
		default:
			options[key] = common.ParseTplWithPanic(v, vars)
		}
	}
	return &CurlAction{
		URL:     common.ParseTplWithPanic(a.URL, vars),
		Options: options,
	}, gerr
}

func (a *CurlAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	newa, err := a.parse(vars, conf)
	if err != nil {
		return "", err
	}
	options := make([]string, 0)
	options = append(options, "curl")
	for k, v := range newa.Options {
		options = append(options, "--"+k)
		options = append(options, v)
	}
	options = append(options, newa.URL)
	return conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
		comm := strings.Join(options, " ")
		return sess.Start(comm), nil
	})
}
