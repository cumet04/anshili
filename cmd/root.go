package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "anshili",
	Short: "desc",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	cobra.OnInitialize()
	RootCmd.AddCommand(crowlCmd)
}
