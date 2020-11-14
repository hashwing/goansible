package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/playbook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfg model.Config

//NewRoot ...
func NewRoot() {
	var rootCmd = &cobra.Command{
		Use: "goansible",
		Run: func(cmd *cobra.Command, args []string) {
			inv, err := inventory.NewYaml(cfg.PlaybookFolder + "/" + cfg.InvFile)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
			ps, err := playbook.UnmarshalFromFile(cfg.PlaybookFolder + "/index.yaml")
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
			err = playbook.Run(cfg, ps, inv)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
		},
	}
	workdir := rootCmd.PersistentFlags().String("workdir", ".", "run playbook in specially dir")
	invFile := rootCmd.PersistentFlags().String("i", "values.yaml", "specify inventory file in a YAML file")
	tag := rootCmd.PersistentFlags().String("tag", "", "use to tag filter")
	workFolder := strings.Replace(*workdir, "\\", "/", -1)
	cfg = model.Config{
		PlaybookFolder: workFolder,
		Tag:            *tag,
		InvFile:        *invFile,
	}
	rootCmd.AddCommand(newRunShellCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
