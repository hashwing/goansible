package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	"github.com/hashwing/goansible/cmd/desktop/api"
	webui "github.com/hashwing/goansible/ui"
	"github.com/zserge/lorca"
)

func main() {
	args := getArgs()
	// Create UI with basic HTML passed via data URI
	ui, err := lorca.New("", "", 1000, 700, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()
	a, err := api.New()
	if err != nil {
		log.Fatal(err)
	}
	ui.Bind("createPlaybook", a.CreatePlaybook)
	ui.Bind("listPlaybook", a.ListPlaybook)
	ui.Bind("deletePlaybook", a.DeletePlaybook)
	ui.Bind("editPlaybook", a.EditPlaybook)
	ui.Bind("runPlaybook", a.RunPlaybook)
	ui.Bind("getLog", a.GetLog)

	webui.New()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(webui.New()))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))
	//ui.Load(fmt.Sprintf("http://%s", "127.0.0.1:8080"))

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}
	log.Println("exiting...")
}

// getArgs
func getArgs() (args []string) {
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	return args
}
