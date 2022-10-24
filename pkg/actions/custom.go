package actions

import (
	"sync"

	"github.com/hashwing/goansible/model"
)

var CustomActions sync.Map

func AddCustomActions(key string, action model.Action) {
	CustomActions.Store(key, action)
}
