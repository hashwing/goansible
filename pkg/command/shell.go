package command

import (
	"os"

	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/playbook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newRunShellCmd() *cobra.Command {
	return &cobra.Command{
		Use: "run group_name shell_command",
		Run: func(cmd *cobra.Command, args []string) {
			inv, err := inventory.NewYaml(cfg.PlaybookFolder + "/" + cfg.InvFile)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
			err = playbook.RunShell(cfg, inv, args[0], args[1])
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
		},
	}
}
