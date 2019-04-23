package cmd

import (
	"github.com/seveirbian/gear/manager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	managerUsage = `Usage:  gear manager
`
)

func init() {
	rootCmd.AddCommand(managerCmd)
	managerCmd.SetUsageTemplate(managerUsage)
}

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start a p2p cluster manager",
	Long:  `Start a p2p cluster manager`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := manager.Init()
		if err != nil {
			logrus.Fatal("Fail to init a manager...")
		}

		manager.Start()
	},
}
