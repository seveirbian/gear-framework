package cmd

import (
	// "fmt"
	"github.com/docker/go-plugins-helpers/graphdriver"
	gearDriver "github.com/seveirbian/gear/graphdriver"
	"github.com/spf13/cobra"
)

var graphdriverUsage = `Usage:  gear graphdriver -m MONITORIP -p MONITORPORT

Options:
  If you need to monitor, then input the monitor ip and port
  -m, --monitor-ip          Monitor node's ip address
  -p, --monitor-port        Monitor node's port(default 2021)
`

var (
	monitorIP string
	monitorPort string
)

func init() {
	rootCmd.AddCommand(graphdriverCmd)
	graphdriverCmd.SetUsageTemplate(graphdriverUsage)
	graphdriverCmd.Flags().StringVarP(&monitorIP, "monitor-ip", "m", "", "Monitor node's ip address")
	graphdriverCmd.Flags().StringVarP(&monitorPort, "monitor-port", "p", "2021", "Monitor node's port")
}

var graphdriverCmd = &cobra.Command{
	Use:   "graphdriver",
	Short: "Start the gear graphdriver",
	Long:  `Start the gear graphdriver`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var monitorServer = ""

		if monitorIP != "" {
			monitorServer = monitorIP + ":" + monitorPort
		}

		gearGraphDriver := &gearDriver.Driver{MonitorServer: monitorServer}
		h := graphdriver.NewHandler(gearGraphDriver)

		h.ServeUnix("geargraphdriver", 0)
	},
}
