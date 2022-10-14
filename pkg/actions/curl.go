package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"golang.org/x/sync/errgroup"
)

var datab = `@- << EOF
%s
EOF
`

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
		case "data-json":
			d := strings.TrimSpace(common.ParseTplWithPanic(v, vars))
			if strings.HasPrefix(d, "{") {
				var v map[string]interface{}
				err := json.Unmarshal([]byte(d), &v)
				if err != nil {
					return nil, err
				}
				j, _ := json.Marshal(v)
				b, _ := json.Marshal(string(j))
				options["data"] = string(b)
			}
			if strings.HasPrefix(d, "[") {
				var v []interface{}
				err := json.Unmarshal([]byte(d), &v)
				if err != nil {
					return nil, err
				}
				j, _ := json.Marshal(v)
				b, _ := json.Marshal(string(j))
				options["data"] = string(b)
			}
		case "data-json-file":
			data, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, v))
			if err != nil {
				return nil, err
			}
			d := strings.TrimSpace(common.ParseTplWithPanic(string(data), vars))
			if strings.HasPrefix(d, "{") {
				var v map[string]interface{}
				err := json.Unmarshal([]byte(d), &v)
				if err != nil {
					return nil, err
				}
				j, _ := json.Marshal(v)
				b, _ := json.Marshal(string(j))
				options["data"] = string(b)
			}
			if strings.HasPrefix(d, "[") {
				var v []interface{}
				err := json.Unmarshal([]byte(d), &v)
				if err != nil {
					return nil, err
				}
				j, _ := json.Marshal(v)
				b, _ := json.Marshal(string(j))
				options["data"] = string(b)
			}
		case "data-file":
			data, err := ioutil.ReadFile(filepath.Join(conf.PlaybookFolder, v))
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(common.ParseTplWithPanic(string(data), vars))
			if err != nil {
				return nil, err
			}
			options["data"] = string(b)
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
			options["data-binary"] = fmt.Sprintf(datab, common.ParseTplWithPanic(string(v), vars))
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
	dataBinary := ""
	for k, v := range newa.Options {
		if k == "data-binary" {
			dataBinary = v
			continue
		}
		options = append(options, "--"+k)
		options = append(options, v)
	}

	options = append(options, newa.URL)
	if dataBinary != "" {
		options = append(options, "--data-binary")
		options = append(options, dataBinary)
	}
	return conn.Exec(ctx, true, func(sess model.Session) (error, *errgroup.Group) {
		comm := strings.Join(options, " ")
		fmt.Println(comm)
		return sess.Start(comm), nil
	})
}
