package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
	"github.com/hashwing/goansible/playbook"
)

var workDir = "."

type Pipeline struct {
	HTTPInput  *HTTPInput  `yaml:"http_input"`
	HTTPOutput *HTTPOutput `yaml:"http_output"`
	Playbook   string      `yaml:"playbook"`
}

type HTTPInput struct {
	Body   string `yaml:"body"`
	Query  string `yaml:"query"`
	Param  string `yaml:"param"`
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

type HTTPOutput struct {
	Body string `yaml:"body"`
}

func main() {
	if len(os.Args) > 1 {
		workDir = os.Args[1]
	}
	http.HandleFunc("/", server)
	http.ListenAndServe(":8095", nil)
}

func server(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Action, Module, Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	var ps []Pipeline
	piped, _ := ioutil.ReadFile(workDir + "/daemon.yaml")
	err := yaml.Unmarshal(piped, &ps)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	var p Pipeline
	for _, pv := range ps {
		if pv.HTTPInput.Method == "" {
			pv.HTTPInput.Method = http.MethodGet
		}
		if pv.HTTPInput.Path == r.URL.Path && r.Method == pv.HTTPInput.Method {
			p = pv
			break
		}
	}
	if p.HTTPInput == nil {
		w.WriteHeader(404)
		return
	}
	vars := &model.Vars{
		Values: make(map[string]interface{}),
	}
	if p.HTTPInput.Query != "" {
		queryVars := make(map[string]interface{})
		r.ParseForm()
		for k, vs := range r.Form {
			queryVars[k] = vs[0]
		}

		common.SetVar(p.HTTPInput.Query, queryVars, vars)
	}

	if p.HTTPInput.Body != "" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		if len(data) > 0 {
			if data[0] == '{' {
				var bmap map[string]interface{}
				err = json.Unmarshal(data, &bmap)
				if err != nil {
					w.WriteHeader(400)
					return
				}
				common.SetVar(p.HTTPInput.Body, bmap, vars)
			}
			if data[0] == '[' {
				var barr []interface{}
				err = json.Unmarshal(data, &barr)
				if err != nil {
					w.WriteHeader(400)
					return
				}
				common.SetVar(p.HTTPInput.Body, barr, vars)
			}
		}
	}
	if p.Playbook == "" {
		p.Playbook = "index.yaml"
	}
	cfg := model.Config{
		PlaybookFolder: workDir,
		PlaybookFile:   p.Playbook,
		InvFile:        "values.yaml",
	}
	res, err := playbook.Run(cfg, vars.Values, nil)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	resv, _ := common.GetVar(p.HTTPOutput.Body, &model.Vars{
		Values: res,
	})
	wbody, err := json.Marshal(resv)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(wbody)
}
