package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/common"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBInitAction struct {
	Dns    string `yaml:"dns"`
	Driver string `yaml:"driver"`
	Name   string `yaml:"name"`
}

func (a *DBInitAction) parse(vars *model.Vars) (*DBInitAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	return &DBInitAction{
		Dns:    common.ParseTplWithPanic(a.Dns, vars),
		Driver: common.ParseTplWithPanic(a.Driver, vars),
		Name:   a.Name,
	}, gerr
}

func (a *DBInitAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	switch parseAction.Driver {
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(parseAction.Dns), &gorm.Config{})
		if err != nil {
			return "", err
		}
		fmt.Println("a")
		common.GlobalVars.Store("db@"+parseAction.Name, db)
	}

	return "", err
}

type DBAction struct {
	Name string        `yaml:"name"`
	SQL  string        `yaml:"sql"`
	Args []interface{} `yaml:"args"`
	Res  string        `yaml:"res"`
}

func (a *DBAction) parse(vars *model.Vars) (*DBAction, error) {
	var gerr error
	defer func() {
		if err := recover(); err != nil {
			gerr = err.(error)
		}
	}()

	return &DBAction{
		Name: common.ParseTplWithPanic(a.Name, vars),
		SQL:  common.ParseTplWithPanic(a.SQL, vars),
		Args: a.Args,
		Res:  a.Res,
	}, gerr
}

func (a *DBAction) Run(ctx context.Context, conn model.Connection, conf model.Config, vars *model.Vars) (string, error) {
	parseAction, err := a.parse(vars)
	if err != nil {
		return "", err
	}
	dbv, ok := common.GlobalVars.Load("db@" + parseAction.Name)
	if !ok {
		return "", errors.New("db not found")
	}
	if db, ok := dbv.(*gorm.DB); ok {
		var res []map[string]interface{}
		err := db.Raw(parseAction.SQL, parseAction.Args...).Scan(&res).Error
		if err != nil {
			return "", err
		}
		if parseAction.Res != "" {
			common.SetVar(parseAction.Res, res, vars)
		}
	}
	return "", nil
}
