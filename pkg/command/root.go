package command

import (
	"fmt"
	"os"

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
			fmt.Println(cfg.PlaybookFolder + "/" + cfg.InvFile)
			inv, err := inventory.NewYaml(cfg.PlaybookFolder + "/" + cfg.InvFile)
			if err != nil {
				fmt.Println("aaaa")
				log.Error(err)
				os.Exit(-1)
			}
			ps, err := playbook.UnmarshalFromFile(cfg.PlaybookFolder + "/" + cfg.PlaybookFile)
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
	cfg = model.Config{}
	rootCmd.PersistentFlags().StringVar(&cfg.PlaybookFolder, "workdir", ".", "run playbook in specially dir")
	rootCmd.PersistentFlags().StringVar(&cfg.InvFile, "i", "values.yaml", "specify inventory file in a YAML file")
	rootCmd.PersistentFlags().StringVar(&cfg.PlaybookFile, "p", "index.yaml", "specify playbook file in a YAML file")
	rootCmd.PersistentFlags().StringVar(&cfg.Tag, "tag", "", "use to tag filter")

	rootCmd.AddCommand(newRunShellCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
