package command

import (
	"github.com/spf13/cobra"
)

func newChartCmd() *cobra.Command {
	//valuesFile := ""
	cmd := &cobra.Command{
		Use: "chart",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	//valuesFile = *cmd.Flags().StringP("values", "f", "values.yaml", "specify values in a YAML file")
	return cmd
}
