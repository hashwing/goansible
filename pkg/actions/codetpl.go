package actions

import (
	"bytes"
	"context"
	"html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"
)

type CodeTplAction struct {
	TplDir string `yaml:"tpl_dir"`
	DstDir string `yaml:"dst_dir"`
}

func (a *CodeTplAction) parse(vars *model.Vars) (*CodeTplAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()
	return &CodeTplAction{
		TplDir: common.ParseTplWithPanic(a.TplDir, vars),
		DstDir: common.ParseTplWithPanic(a.DstDir, vars),
	}, gerr
}

func (a *CodeTplAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	newa, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	tplDir := newa.TplDir
	dstDir := newa.DstDir
	tplFiles, err := getFiles(tplDir)
	if err != nil {
		return "", err
	}
	pipeStr := ""
	for _, tplF := range tplFiles {
		data, err := ioutil.ReadFile(tplF.Path)
		if err != nil {
			return "", err
		}
		dataStr := string(data)
		path := strings.TrimPrefix(tplF.Path, tplDir+"/")
		if path == ".drone.yml" {
			pipeStr = dataStr
			continue
		}
		line, err := bytes.NewBuffer(data).ReadString('\n')
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(line, "//@") {
			path = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(line, "\n", ""), "\r", ""), "//@", "")
			dataStr = strings.Replace(dataStr, line, "", 1)
		}

		fdata, err := parseTpl(dataStr, vars)
		if err != nil {
			return "", err
		}
		fpath, err := parseTpl(path, vars)
		if err != nil {
			return "", err
		}
		fpath = dstDir + "/" + fpath
		err = os.MkdirAll(filepath.Dir(fpath), 0775)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(fpath, []byte(fdata), 0664)
		if err != nil {
			return "", err
		}
	}
	return pipeStr, nil
}

type tplInfo struct {
	Path string
	Name string
}

func getFiles(dir string) ([]tplInfo, error) {
	res := make([]tplInfo, 0)
	tplFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, tplF := range tplFiles {
		if tplF.IsDir() {
			subFiles, err := getFiles(dir + "/" + tplF.Name())
			if err != nil {
				return nil, err
			}
			res = append(res, subFiles...)
			continue
		}
		path := dir + "/" + tplF.Name()
		name := strings.Replace(strings.Replace(path, ".", "_", -1), "/", "_", -1)
		res = append(res, tplInfo{Path: path, Name: name})
	}
	return res, nil
}

func parseTpl(tpl string, vars *model.Vars) (string, error) {
	tmpl, err := newTpl().Parse(tpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, vars)
	return b.String(), err
}

func newTpl() *template.Template {
	tmpl := template.New("tpl")
	tmpl = tmpl.Funcs(funcMap)
	return tmpl
}

// funcMap provides extra functions for the templates.
var funcMap = template.FuncMap{
	"substr":    substr,
	"replace":   replace,
	"randomStr": randomStr,
}

func substr(s string, i int) string {
	return s[:i]
}

func replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func randomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
