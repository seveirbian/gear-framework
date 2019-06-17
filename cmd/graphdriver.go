package cmd

import (
	// "fmt"
	"github.com/docker/go-plugins-helpers/graphdriver"
	gearDriver "github.com/seveirbian/gear/graphdriver"
	"github.com/spf13/cobra"
)

var graphdriverUsage = `Usage:  gear graphdriver`

var (
	monitorIP string
	monitorPort string
)

func init() {
	rootCmd.AddCommand(graphdriverCmd)
	graphdriverCmd.SetUsageTemplate(graphdriverUsage)
}

var graphdriverCmd = &cobra.Command{
	Use:   "graphdriver",
	Short: "Start the gear graphdriver",
	Long:  `Start the gear graphdriver`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		gearGraphDriver := &gearDriver.Driver{}
		h := graphdriver.NewHandler(gearGraphDriver)

		h.ServeUnix("geargraphdriver", 0)
	},
}
