package command

import (
	"os"
	"strings"

	"github.com/hashwing/goansible/model"
	"github.com/hashwing/goansible/pkg/inventory"
	"github.com/hashwing/goansible/pkg/termutil"
	"github.com/hashwing/goansible/playbook"
	"github.com/spf13/cobra"
)

var cfg model.Config

//NewRoot ...
func NewRoot() {
	var rootCmd = &cobra.Command{
		Use:  "goansible",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cfg.PlaybookFile = args[0]
			}
			if cfg.Tag != "" {
				cfg.Tags = strings.Split(cfg.Tag, ",")
			}
			inv, err := inventory.NewYaml(cfg.PlaybookFolder + "/" + cfg.InvFile)
			if err != nil {
				if !os.IsNotExist(err) {
					termutil.Errorf(err.Error())
					os.Exit(-1)
				}
				termutil.Changedf("inventory file '%s' not found, use default inventory", cfg.InvFile)
				inv = &model.DefaultInventory{}
			}

			ps, err := playbook.UnmarshalFromFile(cfg.PlaybookFolder + "/" + cfg.PlaybookFile)
			if err != nil {
				termutil.Errorf(err.Error())
				os.Exit(-1)
			}
			err = playbook.Run(cfg, ps, inv)
			if err != nil {
				termutil.Errorf(err.Error())
				os.Exit(-1)
			}
		},
	}
	cfg = model.Config{}
	rootCmd.PersistentFlags().StringVar(&cfg.PlaybookFolder, "workdir", ".", "run playbook in specially dir")
	rootCmd.PersistentFlags().StringVar(&cfg.InvFile, "i", "values.yaml", "specify inventory file in a YAML file")
	rootCmd.PersistentFlags().StringVar(&cfg.PlaybookFile, "p", "index.yaml", "specify playbook file in a YAML file")
	rootCmd.PersistentFlags().StringVar(&cfg.Tag, "tags", "", "use to tag filter")
	rootCmd.AddCommand(newRunShellCmd())
	rootCmd.AddCommand(newRunInitCmd())
	if err := rootCmd.Execute(); err != nil {
		termutil.Errorf(err.Error())
		os.Exit(1)
	}
}
