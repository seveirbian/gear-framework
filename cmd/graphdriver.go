package cmd

import (
    // "fmt"
    "github.com/spf13/cobra"
    gearDriver "github.com/seveirbian/gear/graphdriver"
    "github.com/docker/go-plugins-helpers/graphdriver"
)

var graphdriverUsage = `Usage:  gear graphdriver`

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