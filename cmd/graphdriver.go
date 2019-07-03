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
  -p, --manager-port        Manager node's port(default 2019)`

var (
	driverManagerIp string
	driverManagerPort string
)

func init() {
	rootCmd.AddCommand(graphdriverCmd)
	graphdriverCmd.SetUsageTemplate(graphdriverUsage)
	graphdriverCmd.Flags().StringVarP(&driverManagerIp, "manager-ip", "m", "", "Manager node's ip address")
	graphdriverCmd.MarkFlagRequired("manager-ip")
	graphdriverCmd.Flags().StringVarP(&driverManagerPort, "manager-port", "p", "2019", "Manager node's port")
	
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
		}
		h := graphdriver.NewHandler(gearGraphDriver)

		h.ServeUnix("geargraphdriver", 0)
	},
}
