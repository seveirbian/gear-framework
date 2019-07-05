package cmd

import (
	// "fmt"
	"github.com/docker/go-plugins-helpers/graphdriver"
	gearDriver "github.com/seveirbian/gear/graphdriver"
	"github.com/spf13/cobra"
)

var graphdriverUsage = `Usage:  gear graphdriver

Options:
  -m, --manager-ip          Manager node's ip address
  -p, --manager-port        Manager node's port(default 2019)
  -t, --monitor-ip          Monitor node's ip address
      --monitor-port        Monitor node's port(default 2021)
  `

var (
	driverManagerIp string
	driverManagerPort string
	driverMonitorIp string
	driverMonitorPort string
)

func init() {
	rootCmd.AddCommand(graphdriverCmd)
	graphdriverCmd.SetUsageTemplate(graphdriverUsage)
	graphdriverCmd.Flags().StringVarP(&driverManagerIp, "manager-ip", "m", "", "Manager node's ip address")
	graphdriverCmd.MarkFlagRequired("manager-ip")
	graphdriverCmd.Flags().StringVarP(&driverManagerPort, "manager-port", "p", "2019", "Manager node's port")
	graphdriverCmd.Flags().StringVarP(&driverMonitorIp, "monitor-ip", "t", "", "Monitor node's ip address")
	graphdriverCmd.MarkFlagRequired("monitor-ip")
	graphdriverCmd.Flags().StringVarP(&driverMonitorPort, "monitor-port", "", "2021", "Monitor node's port")
	
}

var graphdriverCmd = &cobra.Command{
	Use:   "graphdriver",
	Short: "Start the gear graphdriver",
	Long:  `Start the gear graphdriver`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		gearGraphDriver := &gearDriver.Driver{
			ManagerIp: driverManagerIp, 
			ManagerPort: driverManagerPort, 
			MonitorIp: driverMonitorIp, 
			MonitorPort: driverMonitorPort, 
		}
		h := graphdriver.NewHandler(gearGraphDriver)

		h.ServeUnix("geargraphdriver", 0)
	},
}
