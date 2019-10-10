package actions

import (
	"context"
	"testing"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/transport"
)

func TestFile(t *testing.T) {
	fa := FileAction{
		Src:  "file.go",
		Dest: "/root/file.go",
	}

	conn, err := transport.Connect("root", "sunrunvas", "", "188.8.2.130:22")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = fa.Run(context.Background(), conn, model.Config{
		PlaybookFolder: "./",
	}, &model.Vars{})
	if err != nil {
		t.Error(err)
		return
	}
}
